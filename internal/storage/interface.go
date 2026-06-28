package storage

import (
	"context"
)

type ObjectStorage interface {
	PutObject(ctx context.Context, bucketName, objectName string, data []byte) error
}
