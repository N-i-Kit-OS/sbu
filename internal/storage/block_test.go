package storage

import (
	"context"
	"database/sql"
	"diplom/internal/sbudb"
	"testing"
)

func TestProcessBlock(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := sbudb.InitSchema(db); err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}

	mockStorage := NewMockStorage()
	processor, err := NewBlockProcessor(mockStorage, "test-bucket", db)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}
	defer processor.Close()

	snapshotID, err := createTestSnapshot(db, "test-snapshot")
	if err != nil {
		t.Fatalf("failed to create snapshot: %v", err)
	}
	fileID, err := insertTestFile(db, "test.txt", snapshotID)
	if err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	ctx := context.Background()
	data := []byte("hello world")

	err = processor.ProcessBlock(ctx, data, fileID, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func createTestSnapshot(db *sql.DB, name string) (int64, error) {
	res, err := db.Exec("INSERT INTO snapshots (timestamp, name) VALUES (datetime('now'), ?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func insertTestFile(db *sql.DB, path string, snapshotID int64) (int64, error) {
	res, err := db.Exec("INSERT INTO files (path_file, id_snapshot) VALUES (?, ?)", path, snapshotID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
