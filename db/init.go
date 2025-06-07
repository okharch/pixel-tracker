package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
)

func MustConnect(ctx context.Context) *pgx.Conn {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/pixel_tracker"
	}

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	return conn
}
