package events

import (
	"context"
	"github.com/okharch/pixel-tracker/db"
	"github.com/okharch/pixel-tracker/model"
	"github.com/okharch/pixel-tracker/users"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPutSingleEvent(t *testing.T) {
	// Initialize the user manager and database connection
	ctx := context.Background()
	db := db.MustConnect(ctx)
	userManager, err := users.NewUserManager(ctx, db)
	if err != nil {
		t.Fatalf("Failed to initialize user manager: %v", err)
	}

	// truncate the track_events table to start fresh
	_, err = db.Exec(ctx, "TRUNCATE TABLE track_events")
	if err != nil {
		t.Fatalf("Failed to truncate track_events table: %v", err)
	}

	// Create a sample event
	event := model.TrackEvent{
		EmailHash: "test@email.com",
		EventHash: "event123",
	}
	// Add the event to the event pool
	AddEvents(ctx, db, []model.TrackEvent{event}, userManager)
	// Flush the events to the database
	if err := FlushEvents(ctx, db, userManager, true); err != nil {
		t.Fatalf("Failed to flush events: %v", err)
	}
	// Verify that the event was added to the database
	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM track_events WHERE event_hash = $1", event.EventHash).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query track_events: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 event in track_events, got %d", count)
	}
}

func TestPutMultipleEvent(t *testing.T) {
	// Initialize the user manager and database connection
	ctx := context.Background()
	umDb := db.MustConnect(ctx) // Initialize the database connection
	evDb := db.MustConnect(ctx) // Create a separate connection for event processing
	userManager, err := users.NewUserManager(ctx, umDb)
	if err != nil {
		t.Fatalf("Failed to initialize user manager: %v", err)
	}

	// truncate the track_events table to start fresh
	_, err = evDb.Exec(ctx, "TRUNCATE TABLE track_events")
	if err != nil {
		t.Fatalf("Failed to truncate track_events table: %v", err)
	}
	go CopyWorker(ctx, evDb, userManager, time.Millisecond*10)

	var wg sync.WaitGroup
	// Create multiple sample events
	var c atomic.Int64
	for i := 0; i < 1000; i++ {
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cc := c.Add(1)
				n := strconv.Itoa(int(cc))
				event := model.TrackEvent{
					EmailHash: "test" + n + "@email.com",
					EventHash: "event" + n,
				}
				// Add the event to the event pool
				AddEvents(ctx, evDb, []model.TrackEvent{event}, userManager)
			}()
		}
		time.Sleep(time.Millisecond)
	}
	wg.Wait()
	// Flush the events to the database
	if err := FlushEvents(ctx, evDb, userManager, true); err != nil {
		t.Fatalf("Failed to flush events: %v", err)
	}
	// Verify that the event was added to the database
	var count int64
	testDb := db.MustConnect(ctx)
	err = testDb.QueryRow(ctx, "SELECT COUNT(*) FROM track_events").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query track_events: %v", err)
	}
	expectedCount := c.Load()
	if count != expectedCount {
		t.Errorf("Expected %d event in track_events, got %d", expectedCount, count)
	}
}
