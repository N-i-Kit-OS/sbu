package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"diplom/internal/constants"
	"diplom/internal/sbudb"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/restic/chunker"
)

type BlockProcessor struct {
	minioClient *minio.Client
	bucket      string
	db          *sql.DB
	insertStmt  *sql.Stmt
}

func hashBlock(data []byte) string {
	hash := sha256.Sum256(data)
	hash256 := hex.EncodeToString(hash[:])
	return hash256
}

func NewBlockProcessor(minioClient *minio.Client, bucket string, db *sql.DB) (*BlockProcessor, error) {
	if minioClient == nil {
		return nil, fmt.Errorf("minioClient is nil")
	}
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if bucket == "" {
		return nil, fmt.Errorf("bucket is empty")
	}

	stmt, err := db.Prepare("INSERT INTO blocks (id_file, hash, block_index) VALUES (?,?,?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert block statement: %w", err)
	}

	return &BlockProcessor{

		minioClient: minioClient,
		bucket:      bucket,
		db:          db,
		insertStmt:  stmt,
	}, nil
}

func (b *BlockProcessor) ProcessBlock(ctx context.Context, data []byte, fileID int64, blockIndex int) error {
	blockHash := hashBlock(data)

	existBlock, err := sbudb.BlockExists(ctx, b.db, blockHash)
	if err != nil {
		return fmt.Errorf("failed to check block %s: %w", blockHash, err)
	}
	if existBlock {
		return b.InsertBlock(ctx, blockHash, fileID, blockIndex)
	}

	err = b.UploadBlock(ctx, data, blockHash)
	if err != nil {
		return fmt.Errorf("failed to upload block %s: %w", blockHash, err)
	}

	return b.InsertBlock(ctx, blockHash, fileID, blockIndex)
}

func (b *BlockProcessor) UploadBlock(ctx context.Context, data []byte, blockHash string) error {
	_, err := b.minioClient.PutObject(ctx, b.bucket, constants.ObjectPrefix+blockHash, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload block %s: %w", blockHash, err)
	}

	return nil
}

func (b *BlockProcessor) InsertBlock(ctx context.Context, blockHash string, fileID int64, blockIndex int) error {
	_, err := b.insertStmt.ExecContext(ctx, fileID, blockHash, blockIndex)
	if err != nil {
		return fmt.Errorf("failed to insert block %s: %w", blockHash, err)
	}

	return nil
}

func (b *BlockProcessor) Close() error {
	if b.insertStmt != nil {
		return b.insertStmt.Close()
	}

	return nil
}

func (b *BlockProcessor) ProcessFileChunks(ctx context.Context, content io.Reader, fileID int64) error {
	splitter := chunker.NewWithBoundaries(content, chunker.Pol(polynomial), minBlockSize, maxBlockSize)
	buf := make([]byte, bufSize)

	blockIndex := 0
	for {
		ch, err := splitter.Next(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := b.ProcessBlock(ctx, ch.Data, fileID, blockIndex); err != nil {
			return fmt.Errorf("failed to process block: %d: %w", blockIndex, err)
		}

		blockIndex++
	}

	return nil
}
