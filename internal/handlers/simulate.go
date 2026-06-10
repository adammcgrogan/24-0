package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/adammcgrogan/24-0/internal/db"
	"github.com/adammcgrogan/24-0/internal/game"
)

// SimulatePage shows the "team complete" screen where the player reviews
// their roster and launches the season simulation.
func SimulatePage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		http.Redirect(w, r, "/result/"+id, http.StatusSeeOther)
		return
	}
	if !s.IsComplete() {
		http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
		return
	}
	renderTemplate(w, "simulate.html", s)
}

// RunSimulation handles POST /game/{id}/simulate.
// Runs the season simulation, persists the result, and redirects to the
// result page.
func RunSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		// Already simulated — just go to results.
		http.Redirect(w, r, "/result/"+id, http.StatusSeeOther)
		return
	}
	if !s.IsComplete() {
		http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
		return
	}

	result := game.SimulateSeason(s.Picks, game.CachedFieldAverage)
	tier := game.TierForWins(result.Wins).Name

	if err := db.Complete(r.Context(), id, result.Wins, tier, result.Races); err != nil {
		log.Printf("RunSimulation Complete error for session %s: %v", id, err)
		http.Error(w, "simulation failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/result/"+id, http.StatusSeeOther)
}
