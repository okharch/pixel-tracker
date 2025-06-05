package worker

import (
    "bytes"
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/jackc/pgx/v5"
    "pixel-tracker/main"
)

func CopyWorker(db *pgx.Conn, buffer *[]main.TrackEvent, mu *sync.Mutex, flushRunning *bool, flushMu *sync.Mutex) {
    ticker := time.NewTicker(500 * time.Millisecond)
    for range ticker.C {
        flushMu.Lock()
        if *flushRunning {
            flushMu.Unlock()
            continue
        }
        flushMu.Unlock()

        mu.Lock()
        if len(*buffer) == 0 {
            mu.Unlock()
            continue
        }

        batch := make([]main.TrackEvent, len(*buffer))
        copy(batch, *buffer)
        *buffer = (*buffer)[:0]
        mu.Unlock()

        copyToStaging(db, batch)
    }
}

func copyToStaging(db *pgx.Conn, batch []main.TrackEvent) {
    var buf bytes.Buffer
    for _, e := range batch {
        fmt.Fprintf(&buf, "%s	%s
", e.EmailHash, e.EventHash)
    }

    _, err := db.PgConn().CopyFrom(context.Background(), []byte(`COPY staging_user_events (email_hash, event_hash) FROM STDIN`), &buf)
    if err != nil {
        fmt.Println("COPY error:", err)
    }
}
