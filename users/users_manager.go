package users

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

type UserManager struct {
	mu       sync.Mutex
	existing sync.Map      // map[int64]struct{}
	newUsers *bytes.Buffer // Buffer to hold new users before flushing
	db       *pgx.Conn
	pool     sync.Pool
}

func NewUserManager(ctx context.Context, db *pgx.Conn) (*UserManager, error) {
	rows, err := db.Query(ctx, "SELECT id FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to preload users: %w", err)
	}
	defer rows.Close()

	um := UserManager{
		db: db,
	}
	um.pool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	um.newUsers = um.pool.Get().(*bytes.Buffer)

	// Step 1: Read all IDs into a slice
	ids := make([]int64, 0, 1024*1024) // Preallocate a large enough slice
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		ids = append(ids, id)
	}

	// Step 2: Load sync.Map in a goroutine
	go func() {
		for _, id := range ids {
			um.existing.Store(id, struct{}{})
		}
	}()

	return &um, nil
}

func (um *UserManager) hashToID(email string) int64 {
	hash := sha256.Sum256([]byte(email))
	return int64(binary.LittleEndian.Uint64(hash[:8]))
}

func (um *UserManager) AddUser(email string) (id int64) {
	id = um.hashToID(email)

	if _, exists := um.existing.LoadOrStore(id, struct{}{}); exists {
		return
	}

	um.mu.Lock()
	um.newUsers.WriteString(fmt.Sprintf("%d\t%s\n", id, email))
	um.mu.Unlock()

	return
}

func (um *UserManager) FlushUsers(ctx context.Context) error {
	um.mu.Lock()
	if um.newUsers.Len() == 0 {
		um.mu.Unlock()
		return nil
	}
	// Prepare buffer for COPY
	flushBuffer := um.newUsers
	um.newUsers = um.pool.Get().(*bytes.Buffer)
	um.mu.Unlock()

	// COPY into users_staging
	_, err := um.db.PgConn().CopyFrom(
		ctx,
		flushBuffer,
		"COPY users_staging (id, email_hash) FROM STDIN",
	)
	flushBuffer.Reset()
	um.pool.Put(flushBuffer) // Return buffer to pool
	if err != nil {
		return fmt.Errorf("copy to users_staging failed: %w", err)
	}

	// INSERT into users from staging
	_, err = um.db.Exec(ctx, `
        INSERT INTO users (id, email_hash)
        SELECT DISTINCT id, email_hash FROM users_staging
        ON CONFLICT DO NOTHING
    `)
	if err != nil {
		return fmt.Errorf("insert into users failed: %w", err)
	}

	// Clear staging
	_, err = um.db.Exec(ctx, `TRUNCATE users_staging`)
	if err != nil {
		return fmt.Errorf("truncate users_staging failed: %w", err)
	}

	return nil
}
