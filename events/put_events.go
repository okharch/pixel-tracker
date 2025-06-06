package events

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/okharch/pixel-tracker/model"
	"github.com/okharch/pixel-tracker/users"
	"log"
	"sync"
)

var mu sync.Mutex
var pool *sync.Pool

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	eventBuffer = pool.Get().(*bytes.Buffer)
}

var eventBuffer *bytes.Buffer

func AddEvents(events []model.TrackEvent, um *users.UserManager) error {
	if len(events) == 0 {
		return nil
	}
	// Convert events to TrackEventU
	converted := make([]model.TrackEventU, len(events))
	for i, e := range events {
		converted[i].UserId = um.AddUser(e.EmailHash)
		converted[i].EventHash = e.EventHash
	}
	mu.Lock()
	for i := 0; i < len(converted); i++ {
		eventBuffer.WriteString(fmt.Sprintf("%d\t%s\n",
			converted[i].UserId,
			converted[i].EventHash,
		))
	}
	mu.Unlock()
	return nil
}

func FlushEvents(ctx context.Context, db *pgx.Conn, userManager *users.UserManager) {
	// before flushing events, we need to ensure that all users are preloaded
	if err := userManager.FlushUsers(ctx); err != nil {
		log.Fatalf("Failed to flush users: %v", err)
	}
	mu.Lock()
	flush := eventBuffer
	eventBuffer = pool.Get().(*bytes.Buffer)
	mu.Unlock()

	_, err := db.PgConn().CopyFrom(
		ctx,
		flush,
		`COPY track_events (user_id, event_hash) FROM STDIN`,
	)
	flush.Reset()   // Reset the buffer for reuse
	pool.Put(flush) // Return buffer to pool

	if err != nil {
		fmt.Println("COPY error:", err)
	}
}
