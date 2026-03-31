package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func BackupToS3(file string) error {
	endpoint := "localhost:9000"
	accessKeyID := "admin"
	secretKey := "admin123"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}

	bucketExists, err := minioClient.BucketExists(context.Background(), "test")
	if bucketExists {
		_, err := minioClient.FPutObject(context.Background(), "test", file, file, minio.PutObjectOptions{})

		if err != nil {
			return err
		}
	} else {
		err := minioClient.MakeBucket(context.Background(), "test", minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
