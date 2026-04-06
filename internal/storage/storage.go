package storage

import (
	"context"
	"diplom/internal/config"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func UploadFile(conf config.Config) error {

	endpoint := conf.Endpoint
	accessKeyID := conf.AccessKeyID
	secretKey := conf.SecretKey
	dir := conf.Source
	bucket := conf.Bucket
	useSSL := conf.UseSSL

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}

	bucketExists, err := minioClient.BucketExists(context.Background(), bucket)
	if bucketExists {
		err := findFile(dir, minioClient, bucket)
		if err != nil {
			return err
		}

	} else {
		err := minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func findFile(dirName string, client *minio.Client, bucket string) error {
	files, err := os.ReadDir(dirName)
	if err != nil {
		return err
	}
	if len(files) != 0 {
		for _, file := range files {

			if file.IsDir() {
				findFile(dirName+string(os.PathSeparator)+file.Name(), client, bucket)

			} else {
				_, err := client.FPutObject(context.Background(), bucket, dirName+string(os.PathSeparator)+file.Name(), dirName+string(os.PathSeparator)+file.Name(), minio.PutObjectOptions{})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
