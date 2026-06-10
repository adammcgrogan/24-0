package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/adammcgrogan/24-0/internal/f1"
)

// ErrNotFound is returned when a session does not exist.
var ErrNotFound = errors.New("session not found")

func newID() (string, error) {
	b := make([]byte, 8) // 64-bit entropy — not brute-forceable at scale
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession inserts a fresh session and returns its ID.
func CreateSession(ctx context.Context) (string, error) {
	if err := checkPool(); err != nil {
		return "", err
	}
	id, err := newID()
	if err != nil {
		return "", err
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO sessions (id, picks, constructor_skips_left, era_skips_left)
		 VALUES ($1, '[]', 1, 1)`, id)
	if err != nil {
		return "", fmt.Errorf("CreateSession: %w", err)
	}
	return id, nil
}

// GetSession loads a session by ID. Returns ErrNotFound when the row is absent.
func GetSession(ctx context.Context, id string) (*f1.Session, error) {
	if err := checkPool(); err != nil {
		return nil, err
	}
	row := pool.QueryRow(ctx,
		`SELECT id, picks, constructor_skips_left, era_skips_left,
		        pending_spin, pending_component_spin,
		        COALESCE(wins, 0),
		        COALESCE(tier, ''),
		        completed,
		        race_results
		 FROM sessions WHERE id = $1`, id)

	var s f1.Session
	var picksJSON, pendingJSON, pendingComponentJSON, raceResultsJSON []byte

	err := row.Scan(&s.ID, &picksJSON, &s.ConstructorSkipsLeft, &s.EraSkipsLeft,
		&pendingJSON, &pendingComponentJSON, &s.Wins, &s.Tier, &s.Completed, &raceResultsJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetSession: %w", err)
	}

	if len(picksJSON) > 0 {
		if err := json.Unmarshal(picksJSON, &s.Picks); err != nil {
			return nil, fmt.Errorf("GetSession: corrupt picks JSON: %w", err)
		}
	}
	if len(pendingJSON) > 0 {
		var spin f1.SpinResult
		if err := json.Unmarshal(pendingJSON, &spin); err != nil {
			return nil, fmt.Errorf("GetSession: corrupt pending_spin JSON: %w", err)
		}
		s.PendingSpin = &spin
	}
	if len(pendingComponentJSON) > 0 {
		var cspin f1.ComponentSpin
		if err := json.Unmarshal(pendingComponentJSON, &cspin); err != nil {
			return nil, fmt.Errorf("GetSession: corrupt pending_component_spin JSON: %w", err)
		}
		s.PendingComponentSpin = &cspin
	}
	if len(raceResultsJSON) > 0 {
		if err := json.Unmarshal(raceResultsJSON, &s.RaceResults); err != nil {
			return nil, fmt.Errorf("GetSession: corrupt race_results JSON: %w", err)
		}
	}
	return &s, nil
}

// SaveSpin stores the pending spin result on the session.
func SaveSpin(ctx context.Context, sessionID string, spin f1.SpinResult) error {
	if err := checkPool(); err != nil {
		return err
	}
	b, err := json.Marshal(spin)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`UPDATE sessions SET pending_spin = $1 WHERE id = $2`, b, sessionID)
	return err
}

// DecrementConstructorSkip atomically decrements the counter only if it is
// still > 0. Returns false (no error) if the skip was already exhausted.
func DecrementConstructorSkip(ctx context.Context, sessionID string) (bool, error) {
	if err := checkPool(); err != nil {
		return false, err
	}
	tag, err := pool.Exec(ctx,
		`UPDATE sessions
		 SET constructor_skips_left = constructor_skips_left - 1
		 WHERE id = $1 AND constructor_skips_left > 0`,
		sessionID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// DecrementEraSkip atomically decrements the counter only if it is still > 0.
func DecrementEraSkip(ctx context.Context, sessionID string) (bool, error) {
	if err := checkPool(); err != nil {
		return false, err
	}
	tag, err := pool.Exec(ctx,
		`UPDATE sessions
		 SET era_skips_left = era_skips_left - 1
		 WHERE id = $1 AND era_skips_left > 0`,
		sessionID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// AddPick atomically appends a single pick to the session's JSONB array and
// clears the pending spin. This avoids the lost-update race of read-modify-write.
func AddPick(ctx context.Context, sessionID string, pick f1.Pick) error {
	if err := checkPool(); err != nil {
		return err
	}
	b, err := json.Marshal(pick)
	if err != nil {
		return err
	}
	tag, err := pool.Exec(ctx,
		`UPDATE sessions
		 SET picks = picks || $1::jsonb, pending_spin = NULL
		 WHERE id = $2 AND completed = FALSE`,
		// Wrap in a JSON array element so || appends one object, not merges keys
		json.RawMessage("["+string(b)+"]"), sessionID)
	if err != nil {
		return fmt.Errorf("AddPick: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("AddPick: session not found or already completed")
	}
	return nil
}

// SaveComponentSpin stores the pending component spin result.
func SaveComponentSpin(ctx context.Context, sessionID string, spin f1.ComponentSpin) error {
	if err := checkPool(); err != nil {
		return err
	}
	b, err := json.Marshal(spin)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`UPDATE sessions SET pending_component_spin = $1 WHERE id = $2`, b, sessionID)
	return err
}

// AddComponentPick atomically appends a component pick and clears pending_component_spin.
func AddComponentPick(ctx context.Context, sessionID string, pick f1.Pick) error {
	if err := checkPool(); err != nil {
		return err
	}
	b, err := json.Marshal(pick)
	if err != nil {
		return err
	}
	tag, err := pool.Exec(ctx,
		`UPDATE sessions
		 SET picks = picks || $1::jsonb, pending_component_spin = NULL
		 WHERE id = $2 AND completed = FALSE`,
		json.RawMessage("["+string(b)+"]"), sessionID)
	if err != nil {
		return fmt.Errorf("AddComponentPick: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("AddComponentPick: session not found or already completed")
	}
	return nil
}

// Complete marks the session as done and records its result.
func Complete(ctx context.Context, sessionID string, wins int, tier string, races []f1.RaceResult) error {
	if err := checkPool(); err != nil {
		return err
	}
	racesJSON, err := json.Marshal(races)
	if err != nil {
		return fmt.Errorf("Complete: marshal races: %w", err)
	}
	_, err = pool.Exec(ctx,
		`UPDATE sessions SET completed = TRUE, wins = $1, tier = $2, pending_spin = NULL, race_results = $3
		 WHERE id = $4`, wins, tier, racesJSON, sessionID)
	return err
}
