package worker

import (
    "context"
    "log"
    "os"

    "github.com/jackc/pgx/v5"
)

func MustConnect() *pgx.Conn {
    url := os.Getenv("DATABASE_URL")
    if url == "" {
        url = "postgres://postgres:postgres@localhost:5432/postgres"
    }

    conn, err := pgx.Connect(context.Background(), url)
    if err != nil {
        log.Fatal("Unable to connect to database:", err)
    }

    return conn
}
