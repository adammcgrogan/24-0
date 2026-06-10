package handlers

import (
	"net/http"

	"github.com/adammcgrogan/24-0/internal/db"
)

func LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	entries, err := db.TopScores(r.Context(), 50)
	if err != nil {
		http.Error(w, "leaderboard unavailable", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "leaderboard.html", map[string]any{
		"Entries": entries,
	})
}
