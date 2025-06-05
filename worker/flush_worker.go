package worker

import (
    "context"
    "sync"
    "time"

    "github.com/jackc/pgx/v5"
)

func FlushWorker(db *pgx.Conn, flushRunning *bool, mu *sync.Mutex) {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        mu.Lock()
        *flushRunning = true
        mu.Unlock()

        _, err := db.Exec(context.Background(), "CALL process_staging_user_events()")
        if err != nil {
            println("Flush error:", err.Error())
        }

        mu.Lock()
        *flushRunning = false
        mu.Unlock()
    }
}
