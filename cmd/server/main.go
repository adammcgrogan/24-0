package main

import (
	"log"
	"net/http"
	"os"

	"github.com/adammcgrogan/24-0/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server.New()))
}
