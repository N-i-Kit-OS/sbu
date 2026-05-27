package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func ensureBucket(ctx context.Context, client *minio.Client, bucket string) error {
	if bucket == "" {
		return fmt.Errorf("bucket name is empty")
	}
	if client == nil {
		return fmt.Errorf("minio client is nil")
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s existence: %w", bucket, err)
	}

	if !exists {
		err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("create bucket %s: %w", bucket, err)
		}
	}

	return nil
}
