package handlers

import (
	"errors"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"github.com/adammcgrogan/24-0/internal/db"
	"github.com/adammcgrogan/24-0/internal/game"
)

func ResultPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if !s.Completed {
		http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
		return
	}

	tier := game.TierForWins(s.Wins)
	renderTemplate(w, "result.html", map[string]any{
		"Session": s,
		"Tier":    tier,
	})
}

func SubmitScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Limit body to 8 KB — far more than any leaderboard name needs.
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	// Validate by rune count, not byte count, so multi-byte characters are
	// treated the same as ASCII. Postgres VARCHAR(50) counts characters.
	runeCount := utf8.RuneCountInString(name)
	if runeCount < 1 || runeCount > 50 {
		http.Error(w, "name must be 1–50 characters", http.StatusBadRequest)
		return
	}

	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, "session not found", http.StatusBadRequest)
		} else {
			log.Printf("SubmitScore GetSession error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	if !s.Completed {
		http.Error(w, "game not completed", http.StatusBadRequest)
		return
	}

	if err := db.SubmitScore(r.Context(), id, name, s.Wins, s.Tier); err != nil {
		log.Printf("SubmitScore error for session %s: %v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/leaderboard", http.StatusSeeOther)
}
