package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"diplom/internal/config"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
func Backup(conf config.BackupConfig, client *minio.Client) error {

	// init db
	db, err := setupDB(conf.Bucket, client)
	if err != nil {
		return err
	}
	defer db.Close()

	// ensure bucket
	if err := ensureBucket(client, conf.Bucket); err != nil {
		return err
	}

	// create snapshot
	snapshotID, err := createSnapshot(db)
	if err != nil {
		return err
	}

	// backup files
	if err := backupFiles(conf.Source, client, conf.Bucket, db, snapshotID); err != nil {
		return err
	}

	// upload db
	return uploadDB(client, conf.Bucket)
}

func createSnapshot(db *sql.DB) (int64, error) {

	// get time
	nowISO := time.Now().UTC().Format(time.RFC3339)

	// create snapshot
	res, err := db.Exec("INSERT INTO snapshots (timestamp) VALUES (?)", nowISO)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()

}

func backupFiles(dirName string, client *minio.Client, bucket string, db *sql.DB, snapshotID int64) error {

	// read dir
	files, err := os.ReadDir(dirName)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullPath := filepath.Join(dirName, file.Name())

		// backup files
		if file.IsDir() {
			if err := backupFiles(fullPath, client, bucket, db, snapshotID); err != nil {
				return err
			}
		} else {
			if err := uploadBlocks(fullPath, client, bucket, db, snapshotID); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupDB(bucket string, client *minio.Client) (*sql.DB, error) {

	// check db
	_, err := client.StatObject(context.Background(), bucket, nameDB, minio.StatObjectOptions{})

	// download db
	if err == nil {

		err = client.FGetObject(context.Background(), bucket, nameDB, nameDB, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to download db: %w", err)
		}

	}

	db, err := sql.Open("sqlite", nameDB)
	if err != nil {
		return nil, err
	}

	if err := initDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func uploadDB(client *minio.Client, bucket string) error {

	_, err := client.FPutObject(context.Background(), bucket, nameDB, nameDB, minio.PutObjectOptions{})
	return err

}
func insertFile(db *sql.DB, filePath string, snapshotID int64) (int64, error) {

	// insert file to db
	res, err := db.Exec("INSERT INTO files (path_file, id_snapshot) VALUES (?, ?)", filePath, snapshotID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func normalizePath(path string) string {
	volume := filepath.VolumeName(path)
	withoutVolume := strings.TrimPrefix(path, volume)
	cleanedPath := filepath.Clean(withoutVolume)
	sepSlashPath := filepath.ToSlash(cleanedPath)
	return strings.TrimPrefix(sepSlashPath, "/")
}

func uploadBlocks(filePath string, client *minio.Client, bucket string, db *sql.DB, snapshotID int64) error {

	// insert file to db
	fileID, err := insertFile(db, normalizePath(filePath), snapshotID)
	if err != nil {
		return err
	}

	// read file
	content, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer content.Close()

	// split file into blocks
	chunk := chunker.New(content, chunker.Pol(polynomial))
	buf := make([]byte, bufSize)

	// insert blocks
	stmt, err := db.Prepare("INSERT INTO blocks (id_file, hash, block_index) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// process split file
	blockIndex := 0
	for {
		ch, err := chunk.Next(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// hashing and uploading block
		if err := processBlock(ch, client, bucket, stmt, fileID, blockIndex); err != nil {
			return err
		}

		blockIndex++
	}
	return nil
}

func processBlock(ch chunker.Chunk, client *minio.Client, bucket string, stmt *sql.Stmt, fileID int64, blockIndex int) error {

	// hashing
	hash := sha256.Sum256(ch.Data)
	hash256 := hex.EncodeToString(hash[:])

	// uploading block
	_, err := client.PutObject(context.Background(), bucket, prefObj+hash256, bytes.NewReader(ch.Data), int64(len(ch.Data)), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	_, err = stmt.Exec(fileID, hash256, blockIndex)
	return err
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
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if !exists {
		return client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	}
	return nil
}

func ConnectToS3(cfg config.S3Config) (*minio.Client, error) {

	// read config
	endpoint := cfg.Endpoint
	accessKeyID := cfg.AccessKeyID
	secretKey := cfg.SecretKey
	useSSL := cfg.UseSSL
	region := cfg.Region

	// create minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
