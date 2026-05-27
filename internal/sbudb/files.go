package sbudb

import (
	"context"
	"database/sql"
	"fmt"
)

func InsertFile(ctx context.Context, db *sql.DB, filePath string, snapshotID int64) (int64, error) {
	if filePath == "" {
		return 0, fmt.Errorf("file path is empty")
	}
	if snapshotID <= 0 {
		return 0, fmt.Errorf("invalid snapshot ID: %d <= 0", snapshotID)
	}

	res, err := db.ExecContext(ctx, "INSERT INTO files (path_file, id_snapshot) VALUES (?, ?)", filePath, snapshotID)
	if err != nil {
		return 0, fmt.Errorf("insert file %s into snapshot %d: %w", filePath, snapshotID, err)
	}

	return res.LastInsertId()
}
