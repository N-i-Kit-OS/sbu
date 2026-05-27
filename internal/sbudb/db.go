package sbudb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/minio/minio-go/v7"
	_ "modernc.org/sqlite"
)

const nameDB = "sbu.db"

func InitSchema(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS snapshots (id INTEGER PRIMARY KEY AUTOINCREMENT,timestamp TEXT, name TEXT)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (id_file INTEGER PRIMARY KEY AUTOINCREMENT,path_file TEXT,id_snapshot INTEGER,FOREIGN KEY (id_snapshot) REFERENCES snapshots(id))")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS blocks (id_file INTEGER,hash TEXT,block_index INTEGER NOT NULL,FOREIGN KEY (id_file) REFERENCES files(id_file),PRIMARY KEY (id_file, block_index))")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash)")
	if err != nil {
		return err
	}
	return nil
}

func DownloadFromS3(ctx context.Context, bucket string, client *minio.Client) error {
	_, err := client.StatObject(ctx, bucket, nameDB, minio.StatObjectOptions{})
	if err != nil {
		return nil
	}

	return client.FGetObject(ctx, bucket, nameDB, nameDB, minio.GetObjectOptions{})
}

func OpenLocal() (*sql.DB, error) {
	db, err := sql.Open("sqlite", nameDB)
	if err != nil {
		return nil, err
	}

	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func BlockExists(ctx context.Context, db *sql.DB, blockHash string) (bool, error) {
	err := db.QueryRowContext(ctx, "SELECT 1 FROM blocks WHERE hash = ? LIMIT 1", blockHash).Scan(new(int))
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check block %s: %w", blockHash, err)
	}

	return true, nil
}

func UploadToS3(ctx context.Context, bucket string, client *minio.Client) error {
	if _, err := client.FPutObject(ctx, bucket, nameDB, nameDB, minio.PutObjectOptions{}); err != nil {
		return fmt.Errorf("failed to upload DB: %w", err)
	}

	return nil
}
