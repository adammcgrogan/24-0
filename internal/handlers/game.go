package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/adammcgrogan/24-0/internal/db"
	"github.com/adammcgrogan/24-0/internal/f1"
	"github.com/adammcgrogan/24-0/internal/game"
)

func GamePage(w http.ResponseWriter, r *http.Request) {
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
	renderTemplate(w, "draft.html", s)
}

func Spin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		http.Error(w, "game already completed", http.StatusBadRequest)
		return
	}

	remaining := s.RemainingEras()
	if len(remaining) == 0 {
		http.Error(w, "all eras already drafted", http.StatusBadRequest)
		return
	}

	spin, err := game.Spin(f1.All(), remaining, nil)
	if err != nil {
		log.Printf("Spin error for session %s: %v", id, err)
		http.Error(w, "spin failed", http.StatusInternalServerError)
		return
	}
	if err := db.SaveSpin(r.Context(), id, spin); err != nil {
		log.Printf("SaveSpin error for session %s: %v", id, err)
		http.Error(w, "spin failed", http.StatusInternalServerError)
		return
	}

	renderPartial(w, "spin_result.html", map[string]any{
		"Session": s,
		"Spin":    spin,
	})
}

func ConstructorSkip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.PendingSpin == nil {
		http.Error(w, "no active spin to skip", http.StatusBadRequest)
		return
	}

	// Atomic conditional decrement — returns false if already exhausted.
	decremented, err := db.DecrementConstructorSkip(r.Context(), id)
	if err != nil {
		log.Printf("DecrementConstructorSkip error: %v", err)
		http.Error(w, "skip failed", http.StatusInternalServerError)
		return
	}
	if !decremented {
		http.Error(w, "no constructor skips remaining", http.StatusBadRequest)
		return
	}

	// Respin in the same era.
	era := s.PendingSpin.Era
	spin, err := game.Spin(f1.All(), s.RemainingEras(), &era)
	if err != nil {
		log.Printf("Spin (constructor skip) error: %v", err)
		http.Error(w, "respin failed", http.StatusInternalServerError)
		return
	}
	if err := db.SaveSpin(r.Context(), id, spin); err != nil {
		log.Printf("SaveSpin (constructor skip) error: %v", err)
		http.Error(w, "respin failed", http.StatusInternalServerError)
		return
	}

	// Reflect the decrement locally for the template.
	s.ConstructorSkipsLeft--
	renderPartial(w, "spin_result.html", map[string]any{
		"Session": s,
		"Spin":    spin,
	})
}

func EraSkip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	// Era skip only makes sense when a spin is pending.
	if s.PendingSpin == nil {
		http.Error(w, "no active spin to skip", http.StatusBadRequest)
		return
	}

	// Atomic conditional decrement.
	decremented, err := db.DecrementEraSkip(r.Context(), id)
	if err != nil {
		log.Printf("DecrementEraSkip error: %v", err)
		http.Error(w, "era skip failed", http.StatusInternalServerError)
		return
	}
	if !decremented {
		http.Error(w, "no era skips remaining", http.StatusBadRequest)
		return
	}

	spin, err := game.Spin(f1.All(), s.RemainingEras(), nil)
	if err != nil {
		log.Printf("Spin (era skip) error: %v", err)
		http.Error(w, "respin failed", http.StatusInternalServerError)
		return
	}
	if err := db.SaveSpin(r.Context(), id, spin); err != nil {
		log.Printf("SaveSpin (era skip) error: %v", err)
		http.Error(w, "respin failed", http.StatusInternalServerError)
		return
	}

	s.EraSkipsLeft--
	renderPartial(w, "spin_result.html", map[string]any{
		"Session": s,
		"Spin":    spin,
	})
}

func Pick(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	driverIndex := chi.URLParam(r, "index") // "a" or "b"

	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		http.Error(w, "game already completed", http.StatusBadRequest)
		return
	}
	if s.PendingSpin == nil {
		http.Error(w, "no active spin", http.StatusBadRequest)
		return
	}

	var chosen f1.Driver
	switch driverIndex {
	case "a":
		chosen = s.PendingSpin.DriverA
	case "b":
		chosen = s.PendingSpin.DriverB
	default:
		http.Error(w, "invalid driver index", http.StatusBadRequest)
		return
	}

	pick := f1.Pick{Driver: chosen, Era: s.PendingSpin.Era}

	// AddPick atomically appends via JSONB — no stale-read lost-update.
	if err := db.AddPick(r.Context(), id, pick); err != nil {
		log.Printf("AddPick error for session %s: %v", id, err)
		http.Error(w, "pick failed", http.StatusInternalServerError)
		return
	}

	// Re-read session from DB so picks count is authoritative.
	s, err = db.GetSession(r.Context(), id)
	if err != nil {
		log.Printf("GetSession after AddPick error: %v", err)
		http.Error(w, "pick failed", http.StatusInternalServerError)
		return
	}

	// If all 5 eras filled, run simulation and complete.
	if len(s.Picks) == len(f1.Eras) {
		lineup := make([]f1.Driver, len(s.Picks))
		for i, p := range s.Picks {
			lineup[i] = p.Driver
		}

		result := game.SimulateSeason(lineup, game.CachedFieldAverage)
		tier := game.TierForWins(result.Wins).Name

		if err := db.Complete(r.Context(), id, result.Wins, tier, result.Races); err != nil {
			log.Printf("Complete error for session %s: %v", id, err)
			http.Error(w, "simulation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/result/"+id)
		w.WriteHeader(http.StatusOK)
		return
	}

	renderPartial(w, "slot.html", map[string]any{
		"Session": s,
		"Pick":    pick,
		"SlotNum": len(s.Picks),
	})
}

// handleSessionError distinguishes not-found from transient DB errors so
// monitoring systems see 5xx for real failures.
func handleSessionError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, db.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	log.Printf("session error for %s: %v", r.URL.Path, err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
