package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/okharch/pixel-tracker/db"
	"github.com/okharch/pixel-tracker/events"
	"github.com/okharch/pixel-tracker/model"
	"github.com/okharch/pixel-tracker/users"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx := context.TODO()
	umDb := db.MustConnect(ctx) // Initialize the database connection
	evDb := db.MustConnect(ctx) // Create a separate connection for event processing
	userManager, err := users.NewUserManager(ctx, umDb)
	if err != nil {
		log.Fatalf("Failed to initialize user manager: %v", err)
	}
	go events.CopyWorker(ctx, evDb, userManager, time.Second) // Start the event flushing worker

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
		events.AddEvents(ctx, evDb, rawEvents, userManager) // Assumes `AddEvents` is globally accessible
		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
