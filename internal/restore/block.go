package restore

import (
	"database/sql"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type BlockProcessor struct {
	minioClient *minio.Client
	bucket      string
	db          *sql.DB
	selectStmt  *sql.Stmt
}

func NewRestorProcessor(minioClient *minio.Client, bucket string, db *sql.DB) (*BlockProcessor, error) {
	if minioClient == nil {
		return nil, fmt.Errorf("minioClient is nil")
	}
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if bucket == "" {
		return nil, fmt.Errorf("bucket is empty")
	}

	stmt, err := db.Prepare("SELECT hash FROM blocks WHERE id_file = ? ORDER BY block_index")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare slect block statement: %w", err)
	}

	return &BlockProcessor{
		minioClient: minioClient,
		bucket:      bucket,
		db:          db,
		selectStmt:  stmt,
	}, nil
}
