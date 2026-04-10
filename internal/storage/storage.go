package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"diplom/internal/config"
	"encoding/hex"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/restic/chunker"
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

				file, err := os.Open(dirName + string(os.PathSeparator) + file.Name())
				if err != nil {
					return err
				}
				defer file.Close()

				chunk := chunker.New(file, chunker.Pol(0x3DA3358B4DC173))

				buf := make([]byte, 8*1024*1024)

				for {
					ch, err := chunk.Next(buf)

					if err == io.EOF {
						break
					}

					if err != nil {
						return err
					}

					hash := sha256.Sum256(ch.Data)

					_, err = client.PutObject(context.Background(), bucket, "object/"+string(hex.EncodeToString(hash[:])), bytes.NewReader(ch.Data), int64(len(ch.Data)), minio.PutObjectOptions{})
					if err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}
