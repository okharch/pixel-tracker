package events

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/okharch/pixel-tracker/users"
	"log"
	"time"
)

func CopyWorker(ctx context.Context, db *pgx.Conn, userManager *users.UserManager, delay time.Duration) {
	ticker := time.NewTicker(delay)
	for range ticker.C {
		err := FlushEvents(ctx, db, userManager, true)
		if err != nil {
			log.Printf("Error flushing events: %v", err)
		}
	}
}
