package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"diplom/internal/config"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/restic/chunker"
	_ "modernc.org/sqlite"
)

const (
	polynomial = 0x3DA3358B4DC173
	nameDB     = "sbu.db"
	bufSize    = 8 * 1024 * 1024
	prefObj    = "object/"
)

// upload file to s3
func UploadFile(conf config.Config) error {

	// get time
	nowISO := time.Now().UTC().Format(time.RFC3339)

	// read config
	endpoint := conf.Endpoint
	accessKeyID := conf.AccessKeyID
	secretKey := conf.SecretKey
	dir := conf.Source
	bucket := conf.Bucket
	useSSL := conf.UseSSL

	// create minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
		Region: "ru-central-1",
	})
	if err != nil {
		return err
	}

	// create db
	db, err := sql.Open("sqlite", nameDB)
	if err != nil {
		return err
	}

	// init db
	err = initDB(db)
	if err != nil {
		return err
	}

	defer db.Close()

	// check bucket
	err = ensureBucket(minioClient, bucket)
	if err != nil {
		return err
	}

	// create snapshot
	_, err = db.Exec("INSERT INTO snapshots (timestamp) VALUES (?)", nowISO)
	if err != nil {
		return err
	}

	// upload files and snapshot
	err = findFile(dir, minioClient, bucket, db)
	if err != nil {
		return err
	}

	// upload db
	_, err = minioClient.FPutObject(context.Background(), bucket, nameDB, nameDB, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func findFile(dirName string, client *minio.Client, bucket string, db *sql.DB) error {

	// read dir
	files, err := os.ReadDir(dirName)
	if err != nil {
		return err
	}

	// if dir not empty
	if len(files) != 0 {
		for _, file := range files {
			if file.IsDir() {
				err := findFile(filepath.Join(dirName, file.Name()), client, bucket, db)
				if err != nil {
					return err
				}
			} else {
				err := uploadBlocks(filepath.Join(dirName, file.Name()), client, bucket, db)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func uploadBlocks(filePath string, client *minio.Client, bucket string, db *sql.DB) error {

	res, err := db.Exec("INSERT INTO files (path_file, id_snapshot) VALUES (?, (SELECT id FROM snapshots ORDER BY id DESC LIMIT 1))", filePath)

	fileID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// open file
	content, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer content.Close()

	// create chunker and buffer
	chunk := chunker.New(content, chunker.Pol(polynomial))
	buf := make([]byte, bufSize)

	blockIndex := 0
	for {

		// read chunk
		ch, err := chunk.Next(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// create hash from data and add to blob
		hash := sha256.Sum256(ch.Data)
		hash256 := hex.EncodeToString(hash[:])

		// upload to s3
		_, err = client.PutObject(context.Background(), bucket, prefObj+hash256, bytes.NewReader(ch.Data), int64(len(ch.Data)), minio.PutObjectOptions{})
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO blocks (id_file,hash,block_index) VALUES (?,?,?)", fileID, hash256, blockIndex)
		if err != nil {
			return err
		}

		blockIndex++

	}
	return nil
}

func initDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS snapshots (id INTEGER PRIMARY KEY AUTOINCREMENT,timestamp TEXT)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS files (id_file INTEGER PRIMARY KEY AUTOINCREMENT,path_file TEXT,id_snapshot INTEGER,FOREIGN KEY (id_snapshot) REFERENCES snapshots(id))")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS blocks (id_file INTEGER,hash TEXT,block_index INTEGER NOT NULL,FOREIGN KEY (id_file) REFERENCES files(id_file),PRIMARY KEY (id_file, block_index))")
	if err != nil {
		return err
	}
	return nil
}

func ensureBucket(client *minio.Client, bucket string) error {

	// check bucket
	exist, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}

	// if bucket exist
	if exist {
		return nil
	}

	// create bucket
	err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}

	return nil
}
