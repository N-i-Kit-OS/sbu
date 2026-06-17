package storage

import (
	"context"
	"diplom/internal/config"
	"diplom/internal/sbudb"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sync"
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

	processor, err := NewBlockProcessor(client, conf.Bucket, db)
	if err != nil {
		return fmt.Errorf("failed to create block processor: %w", err)
	}
	defer processor.Close()

	snapshotID, err := sbudb.CreateSnapshot(ctx, db, conf.SnapshotName)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	var files []string

	err = filepath.WalkDir(conf.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error reading file %s: %w", path, err)
		}

		if !d.IsDir() {
			files = append(files, path)
			return nil
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk dir: %w", err)
	}

	numCPUs := runtime.NumCPU()

	jobs := make(chan string, numCPUs)
	errorChan := make(chan error, numCPUs)

	var wg sync.WaitGroup

	for w := 0; w < numCPUs; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for file := range jobs {
				if err := processFile(ctx, db, processor, file, snapshotID); err != nil {
					errorChan <- fmt.Errorf("failed to process file: %w", err)
				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}

	close(jobs)
	wg.Wait()
	close(errorChan)

	errs := []error{}
	for err := range errorChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to process files: %v", errs)
	}

	if err := sbudb.UploadToS3(ctx, conf.Bucket, client); err != nil {
		return fmt.Errorf("failed to upload DB: %w", err)
	}
	return nil
}
