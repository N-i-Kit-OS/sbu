package storage

import (
	"context"
	"diplom/internal/config"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectToS3(cfg config.S3Config) (*minio.Client, error) {

	// validation
	if cfg.Endpoint == "" {
		return nil, errors.New("S3 endpoint is empty")
	}
	if cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, errors.New("S3 credentials is empty")
	}

	// create client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// test connection
	if _, err := client.ListBuckets(context.Background()); err != nil {
		return nil, fmt.Errorf("S3 connection test failed: %w", err)
	}

	return client, nil
}
