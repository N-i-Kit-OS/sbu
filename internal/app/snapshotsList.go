package app

import (
	"context"
	"diplom/internal/sbudb"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

func GetAllSnapshots(client *minio.Client, bucket string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := sbudb.DownloadFromS3(ctx, bucket, client); err != nil {
		return fmt.Errorf(" DB exist but not downloaded: %w", err)
	}

	db, err := sbudb.OpenLocal()
	if err != nil {
		return fmt.Errorf("failed to open local DB: %w", err)
	}
	defer db.Close()

	snapshots, err := sbudb.GetSnapshots(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get all snapshots: %w", err)
	}

	for _, s := range snapshots {
		fmt.Println(s)
	}
	return nil
}
