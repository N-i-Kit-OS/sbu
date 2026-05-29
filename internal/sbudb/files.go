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

func GetFilesByPrefix(ctx context.Context, db *sql.DB, snapshotTimestamp, prefix string) ([]string, error) {
	query := "SELECT f.path_file FROM files f JOIN snapshots s ON f.id_snapshot = s.id WHERE s.timestamp = ? AND f.path_file LIKE ?"

	rows, err := db.QueryContext(ctx, query, snapshotTimestamp, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("query files by prefix: %w", err)
	}
	defer rows.Close()

	var files []string

	for rows.Next() {

		var file string
		err = rows.Scan(&file)
		if err != nil {
			return nil, fmt.Errorf("scan path: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func GetBlockHashesByPath(ctx context.Context, db *sql.DB, snapshotTimestamp, filePath string) ([]string, error) {
	query := "SELECT b.hash FROM blocks b JOIN files f ON f.id_file = b.id_file JOIN snapshots s ON f.id_snapshot = s.id WHERE s.timestamp = ? AND f.path_file = ? ORDER BY b.block_index ASC"

	rows, err := db.QueryContext(ctx, query, snapshotTimestamp, filePath)
	if err != nil {
		return nil, fmt.Errorf("query block hashes: %w", err)
	}
	defer rows.Close()

	var hashes []string

	for rows.Next() {

		var hash string
		err = rows.Scan(&hash)
		if err != nil {
			return nil, fmt.Errorf("scan hash: %w", err)
		}

		hashes = append(hashes, hash)
	}

	return hashes, nil
}
