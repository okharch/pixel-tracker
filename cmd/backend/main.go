package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/okharch/pixel-tracker/events"
	"github.com/okharch/pixel-tracker/model"
	"github.com/okharch/pixel-tracker/users"
	"log"
	"net/http"
	"os"
	"time"
)

func MustConnect() *pgx.Conn {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/pixel_tracker"
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	return conn
}
func CopyWorker(ctx context.Context, db *pgx.Conn, userManager *users.UserManager) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		events.FlushEvents(ctx, db, userManager)
	}
}

func main() {
	db := MustConnect()
	ctx := context.Background()
	userManager, err := users.NewUserManager(ctx, db)
	if err != nil {
		log.Fatalf("Failed to initialize user manager: %v", err)
	}
	go CopyWorker(ctx, db, userManager)

	r := chi.NewRouter()
	r.Post("/track", func(w http.ResponseWriter, r *http.Request) {
		var rawEvents []model.TrackEvent
		if err := json.NewDecoder(r.Body).Decode(&rawEvents); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(rawEvents) == 0 {
			http.Error(w, "no events provided", http.StatusBadRequest)
			return
		}
		events.AddEvents(rawEvents, userManager) // Assumes `AddEvents` is globally accessible
		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
