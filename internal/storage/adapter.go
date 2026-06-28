package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type MinioAdapter struct {
	client *minio.Client
}

func NewMinioAdapter(client *minio.Client) *MinioAdapter {
	return &MinioAdapter{client: client}
}

func (m *MinioAdapter) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	_, err := m.client.PutObject(ctx, bucketName, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}
