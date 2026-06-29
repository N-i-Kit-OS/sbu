package storage

import (
	"context"
	"database/sql"
	"diplom/internal/config"
	"diplom/internal/sbudb"
	"diplom/internal/sbufs"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
)

type backupEnv struct {
	db         *sql.DB
	processor  *BlockProcessor
	snapshotID int64
}

func Backup(conf config.BackupConfig, client *minio.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	env, err := prepareBackupEnv(ctx, conf, client)
	if err != nil {
		return fmt.Errorf("failed to prepare backup env: %w", err)
	}
	defer env.db.Close()
	defer env.processor.Close()

	files, err := sbufs.GetListFiles(conf.Source)
	if err != nil {
		return fmt.Errorf("failed to get list files: %w", err)
	}

	numCPUs := runtime.NumCPU()

	jobs := make(chan string, numCPUs*4)
	errorChan := make(chan error, numCPUs*2)

	var wg sync.WaitGroup

	for w := 0; w < numCPUs*2; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for file := range jobs {
				if err := processFile(ctx, env.db, env.processor, file, env.snapshotID); err != nil {
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

func prepareBackupEnv(ctx context.Context, conf config.BackupConfig, client *minio.Client) (*backupEnv, error) {
	if err := ensureBucket(ctx, client, conf.Bucket); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	if err := sbudb.DownloadFromS3(ctx, conf.Bucket, client); err != nil {
		return nil, fmt.Errorf(" DB exist but not downloaded: %w", err)
	}

	db, err := sbudb.OpenLocal()
	if err != nil {
		return nil, fmt.Errorf("failed to open local DB: %w", err)
	}

	MAdapter := NewMinioAdapter(client)

	processor, err := NewBlockProcessor(MAdapter, conf.Bucket, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create block processor: %w", err)
	}

	snapshotID, err := sbudb.CreateSnapshot(ctx, db, conf.SnapshotName)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &backupEnv{
		db:         db,
		processor:  processor,
		snapshotID: snapshotID,
	}, nil
}
