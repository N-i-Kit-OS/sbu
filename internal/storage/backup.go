package storage

import (
	"context"
	"diplom/internal/config"
	"diplom/internal/sbudb"
	"diplom/internal/sbufs"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

func Backup(conf config.BackupConfig, client *minio.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := ensureBucket(ctx, client, conf.Bucket); err != nil {
		return fmt.Errorf("failed to ensure bucket: %w", err)
	}

	if err := sbudb.DownloadFromS3(ctx, conf.Bucket, client); err != nil {
		return fmt.Errorf(" DB exist but not downloaded: %w", err)
	}

	db, err := sbudb.OpenLocal()
	if err != nil {
		return fmt.Errorf("failed to open local DB: %w", err)
	}
	defer db.Close()

	MAdapter := NewMinioAdapter(client)

	processor, err := NewBlockProcessor(MAdapter, conf.Bucket, db)
	if err != nil {
		return fmt.Errorf("failed to create block processor: %w", err)
	}
	defer processor.Close()

	snapshotID, err := sbudb.CreateSnapshot(ctx, db, conf.SnapshotName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	err = filepath.WalkDir(conf.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error reading file %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

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
	})
	if err != nil {
		return fmt.Errorf(" failed to walk source directory: %w", err)
	}

	if err := sbudb.UploadToS3(ctx, conf.Bucket, client); err != nil {
		return fmt.Errorf("failed to upload DB: %w", err)
	}
	return nil
}
