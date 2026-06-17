package storage

import (
	"context"
	"database/sql"
	"diplom/internal/sbudb"
	"diplom/internal/sbufs"
	"fmt"
	"io"
	"os"
)

func processFile(ctx context.Context, db *sql.DB, processor *BlockProcessor, path string, snapshotID int64) error {
	fileID, err := sbudb.InsertFile(ctx, db, sbufs.NormalizePath(path), snapshotID)
	if err != nil {
		return fmt.Errorf("failed to insert file %s into snapshot %d: %w", path, snapshotID, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info %s: %w", path, err)
	}
	size := info.Size()

	if size < minFileSize {
		data, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		if err := processor.ProcessBlock(ctx, data, fileID, 0); err != nil {
			return fmt.Errorf("failed to process block: %w", err)
		}

		return nil
	}

	err = processor.ProcessFileChunks(ctx, file, fileID)
	if err != nil {
		return fmt.Errorf(" failed to process file %s: %w", path, err)
	}

	return nil
}
