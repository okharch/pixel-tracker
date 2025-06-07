package events

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/okharch/pixel-tracker/model"
	"github.com/okharch/pixel-tracker/users"
	"log"
	"sync"
)

var pool *sync.Pool
var rowsMu sync.Mutex
var curRow int
var rows [][]interface{}

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return make([][]interface{}, 8*1024*1024)
		},
	}
	rows = pool.Get().([][]interface{})
}

func AddEvents(ctx context.Context, db *pgx.Conn, events []model.TrackEvent, um *users.UserManager) {
	eventsU := make([]model.TrackEventU, len(events))
	for i, e := range events {
		eventsU[i].UserId = um.AddUser(e.EmailHash)
		eventsU[i].EventHash = e.EventHash
	}
	rowsMu.Lock()
	for _, e := range eventsU {
		rows[curRow] = []interface{}{e.UserId, e.EventHash}
		curRow++
		if curRow == len(rows) {
			err := FlushEvents(ctx, db, um, false)
			if err != nil {
				log.Printf("flush events: %w", err)
			}
		}
	}
	rowsMu.Unlock()
}

var flushMu sync.Mutex

func FlushEvents(ctx context.Context, db *pgx.Conn, userManager *users.UserManager, lockRows bool) error {
	flushMu.Lock()
	defer flushMu.Unlock()
	if err := userManager.FlushUsers(ctx); err != nil {
		return fmt.Errorf("flush users: %w", err)
	}
	if lockRows {
		rowsMu.Lock()
	}
	cur := curRow
	curRow = 0
	flush := rows
	if cur != 0 {
		rows = pool.Get().([][]interface{})
	}
	if lockRows {
		rowsMu.Unlock()
	}
	if cur == 0 {
		return nil // nothing to flush
	}
	_, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"track_events"},
		[]string{"user_id", "event_hash"},
		pgx.CopyFromRows(flush[:cur]),
	)
	if err != nil {
		return fmt.Errorf("copy from track_events: %w", err)
	}
	pool.Put(flush)
	return err
}
