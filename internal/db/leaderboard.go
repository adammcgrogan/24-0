package db

import (
	"context"
)

type LeaderboardEntry struct {
	Rank      int
	Name      string
	Wins      int
	Tier      string
	SessionID string
}

// SubmitScore adds an entry to the leaderboard for a completed session.
// If the session has already been submitted the call is a no-op (idempotent).
// The UNIQUE constraint on session_id in the schema enforces this atomically.
func SubmitScore(ctx context.Context, sessionID, name string, wins int, tier string) error {
	if err := checkPool(); err != nil {
		return err
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO leaderboard (session_id, name, wins, tier)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (session_id) DO NOTHING`,
		sessionID, name, wins, tier)
	return err
}

// TopScores returns the top N all-time leaderboard entries.
func TopScores(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	if err := checkPool(); err != nil {
		return nil, err
	}
	rows, err := pool.Query(ctx,
		`SELECT name, wins, tier, session_id,
		        RANK() OVER (ORDER BY wins DESC, created_at ASC) AS rank
		 FROM leaderboard
		 ORDER BY wins DESC, created_at ASC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Name, &e.Wins, &e.Tier, &e.SessionID, &e.Rank); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
