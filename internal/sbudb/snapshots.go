package sbudb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const SnapshotNamePrefix = "sbu"

func CreateSnapshot(ctx context.Context, db *sql.DB, snapshotName string) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	name := snapshotName
	if name == "" {
		name = SnapshotNamePrefix + now.Format("2006-01-02_15:04:05")
	}

	res, err := db.ExecContext(ctx, "INSERT INTO snapshots (timestamp, name) VALUES (?, ?)", timestamp, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return res.LastInsertId()
}
