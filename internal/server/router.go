package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/adammcgrogan/24-0/internal/db"
	"github.com/adammcgrogan/24-0/internal/handlers"
)

func New() http.Handler {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Connect(ctx); err != nil {
		log.Printf("WARNING: db connect failed: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)

	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("public/static"))))

	r.Get("/", handlers.Home)
	r.Post("/game/start", handlers.StartGame)
	r.Get("/game/{id}", handlers.GamePage)
	r.Post("/game/{id}/spin", handlers.Spin)
	r.Post("/game/{id}/skip/constructor", handlers.ConstructorSkip)
	r.Post("/game/{id}/skip/era", handlers.EraSkip)
	r.Post("/game/{id}/pick/{index}", handlers.Pick)
	r.Get("/result/{id}", handlers.ResultPage)
	r.Post("/result/{id}/submit", handlers.SubmitScore)
	r.Get("/leaderboard", handlers.LeaderboardPage)

	return r
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' https://cdn.tailwindcss.com; "+
				"style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
				"img-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		next.ServeHTTP(w, r)
	})
}
