package restore

import (
	"context"
	"database/sql"
	"diplom/internal/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	nameDB  = "sbu.db"
	prefObj = "object/"
)

func Restore(conf config.Config) error {

	// connect to s3
	miClient, err := connectToS3(conf)
	if err != nil {
		return err
	}

	// download db
	err = miClient.FGetObject(context.Background(), conf.Bucket, nameDB, nameDB, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	// restore file
	err = restoreFile(conf, miClient)
	if err != nil {
		return err
	}

	return nil
}

func restoreFile(config config.Config, minioClient *minio.Client) error {

	// read and parse config
	dateOfRecovery := strings.Replace(config.DateRecovery, "_", "T", -1) + "Z"
	fromRecovery := config.FromRecovery

	// connect to db
	db, err := sql.Open("sqlite", nameDB)
	if err != nil {
		return err
	}

	// get hashes of file's blocks
	hashes, err := getBlocksHashes(db, fromRecovery, dateOfRecovery)
	if err != nil {
		return err
	}

	// download file
	err = downloadFile(minioClient, config, hashes)
	if err != nil {
		return err
	}

	return nil
}

func downloadFile(minioClient *minio.Client, conf config.Config, hashes []string) error {

	pathSlice := strings.Split(filepath.Join(conf.PathToRecovery, conf.FromRecovery), string(os.PathSeparator))

	fmt.Println(pathSlice)
	// create dir
	err := os.MkdirAll(strings.Join(pathSlice[:len(pathSlice)-1], string(os.PathSeparator)), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(pathSlice[:len(pathSlice)-1], string(os.PathSeparator)))
	// create file
	file, err := os.Create(filepath.Join(pathSlice...))
	if err != nil {
		return err
	}

	fmt.Println(filepath.Join(pathSlice...))

	defer file.Close()

	// download file
	for _, hash := range hashes {

		data, err := minioClient.GetObject(context.Background(), conf.Bucket, prefObj+hash, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

		fmt.Println(prefObj + hash)

		_, err = io.Copy(file, data)
		if err != nil {
			return err
		}
		defer data.Close()
	}

	return nil
}

func getBlocksHashes(db *sql.DB, fromRecovery string, dateOfRecovery string) (hashes []string, err error) {

	// get hashes
	rows, err := db.Query("SELECT b.hash FROM blocks b JOIN files f ON f.id_file = b.id_file JOIN snapshots s ON f.id_snapshot = s.id WHERE s.timestamp = ? AND path_file = ? ORDER BY b.block_index ASC;", dateOfRecovery, fromRecovery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []string

	for rows.Next() {

		var path string
		err = rows.Scan(&path)
		if err != nil {
			return nil, err
		}

		res = append(res, path)
	}

	return res, nil
}

func connectToS3(config config.Config) (*minio.Client, error) {

	// read config
	endpoint := config.Endpoint
	accessKeyID := config.AccessKeyID
	secretKey := config.SecretKey
	useSSL := config.UseSSL

	// create minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
