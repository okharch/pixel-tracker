package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/go-chi/chi/v5"
    "pixel-tracker/worker"
)

type TrackEvent struct {
    EmailHash string `json:"email_hash"`
    EventHash string `json:"event_hash"`
}

var (
    eventBuffer   = make([]TrackEvent, 0, 10000)
    bufferMutex   sync.Mutex
    flushRunning  bool
    flushMutex    sync.Mutex
)

func main() {
    db := worker.MustConnect()
    go worker.CopyWorker(db, &eventBuffer, &bufferMutex, &flushRunning, &flushMutex)
    go worker.FlushWorker(db, &flushRunning, &flushMutex)

    r := chi.NewRouter()
    r.Post("/track", func(w http.ResponseWriter, r *http.Request) {
        var events []TrackEvent
        if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
            http.Error(w, "invalid JSON", http.StatusBadRequest)
            return
        }

        bufferMutex.Lock()
        eventBuffer = append(eventBuffer, events...)
        bufferMutex.Unlock()

        w.WriteHeader(http.StatusNoContent)
    })

    log.Println("Listening on :8080")
    http.ListenAndServe(":8080", r)
}
