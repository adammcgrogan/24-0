package handlers

import (
	"log"
	"net/http"

	"github.com/adammcgrogan/24-0/internal/db"
)

func Home(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home.html", nil)
}

func StartGame(w http.ResponseWriter, r *http.Request) {
	// Drain and discard body (there is none for this POST, but cap it defensively).
	r.Body = http.MaxBytesReader(w, r.Body, 512)

	id, err := db.CreateSession(r.Context())
	if err != nil {
		log.Printf("CreateSession error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/game/"+id, http.StatusSeeOther)
}
