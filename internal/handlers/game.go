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
	if s.IsComplete() {
		http.Redirect(w, r, "/game/"+id+"/simulate", http.StatusSeeOther)
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

	if !s.NeedsDriver() {
		http.Error(w, "already have 2 drivers", http.StatusBadRequest)
		return
	}

	spin, err := game.Spin()
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

	spin, err := game.Spin()
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

	pick := f1.Pick{Type: "driver", Driver: chosen}

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

	// All picks made — send player to the simulate page before revealing results.
	if s.IsComplete() {
		w.Header().Set("HX-Redirect", "/game/"+id+"/simulate")
		w.WriteHeader(http.StatusOK)
		return
	}

	renderPartial(w, "slot.html", map[string]any{
		"Session": s,
		"Pick":    pick,
		"SlotNum": len(s.Picks),
	})
}

// SpinComponent handles POST /game/{id}/spin/component/{category}.
// Generates a pair of component options for the player to choose from.
func SpinComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	category := chi.URLParam(r, "category")

	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		http.Error(w, "game already completed", http.StatusBadRequest)
		return
	}

	// Validate this category hasn't been picked yet.
	valid := false
	for _, cat := range s.RemainingComponentCategories() {
		if cat.ID == category {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "component category not available", http.StatusBadRequest)
		return
	}

	cspin, err := game.SpinComponent(category)
	if err != nil {
		log.Printf("SpinComponent error for session %s: %v", id, err)
		http.Error(w, "spin failed", http.StatusInternalServerError)
		return
	}
	if err := db.SaveComponentSpin(r.Context(), id, cspin); err != nil {
		log.Printf("SaveComponentSpin error for session %s: %v", id, err)
		http.Error(w, "spin failed", http.StatusInternalServerError)
		return
	}

	renderPartial(w, "component_spin.html", map[string]any{
		"Session": s,
		"Spin":    cspin,
	})
}

// PickComponent handles POST /game/{id}/pick/component/{index}.
// Records the player's component choice (a or b).
func PickComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	index := chi.URLParam(r, "index") // "a" or "b"

	s, err := db.GetSession(r.Context(), id)
	if err != nil {
		handleSessionError(w, r, err)
		return
	}
	if s.Completed {
		http.Error(w, "game already completed", http.StatusBadRequest)
		return
	}
	if s.PendingComponentSpin == nil {
		http.Error(w, "no active component spin", http.StatusBadRequest)
		return
	}

	var chosen f1.TeamComponent
	switch index {
	case "a":
		chosen = s.PendingComponentSpin.OptionA
	case "b":
		chosen = s.PendingComponentSpin.OptionB
	default:
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	pick := f1.Pick{
		Type:      "component",
		Component: &chosen,
		Era:       s.PendingComponentSpin.Era,
	}

	if err := db.AddComponentPick(r.Context(), id, pick); err != nil {
		log.Printf("AddComponentPick error for session %s: %v", id, err)
		http.Error(w, "pick failed", http.StatusInternalServerError)
		return
	}

	s, err = db.GetSession(r.Context(), id)
	if err != nil {
		log.Printf("GetSession after AddComponentPick error: %v", err)
		http.Error(w, "pick failed", http.StatusInternalServerError)
		return
	}

	if s.IsComplete() {
		w.Header().Set("HX-Redirect", "/game/"+id+"/simulate")
		w.WriteHeader(http.StatusOK)
		return
	}

	renderPartial(w, "slot.html", map[string]any{
		"Session": s,
		"Pick":    pick,
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
