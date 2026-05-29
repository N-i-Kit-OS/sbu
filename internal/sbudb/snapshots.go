package sbudb

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const snapshotNamePrefix = "sbu"

type SnapshotInfo struct {
	Timestamp string
	Name      string
}

func CreateSnapshot(ctx context.Context, db *sql.DB, snapshotName string) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	name := snapshotName
	if name == "" {
		name = snapshotNamePrefix + now.Format("2006-01-02_15:04:05")
	}

	res, err := db.ExecContext(ctx, "INSERT INTO snapshots (timestamp, name) VALUES (?, ?)", timestamp, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return res.LastInsertId()
}

func GetSnapshots(ctx context.Context, db *sql.DB) ([]SnapshotInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}

	rows, err := db.QueryContext(ctx, "SELECT timestamp, name FROM snapshots")
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}
	defer rows.Close()

	var res []SnapshotInfo

	for rows.Next() {
		var timestamp, name string
		err = rows.Scan(&timestamp, &name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		res = append(res, SnapshotInfo{Timestamp: timestamp, Name: name})
	}

	return res, nil
}
