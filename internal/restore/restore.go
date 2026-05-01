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
)

const (
	nameDB  = "sbu.db"
	prefObj = "object/"
)

func Restore(conf config.RestoreConfig, client *minio.Client) error {

	// download db
	err := client.FGetObject(context.Background(), conf.Bucket, nameDB, nameDB, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	// restore file
	err = restoreFiles(conf, client)
	if err != nil {
		return err
	}

	return nil
}

func restoreFiles(conf config.RestoreConfig, minioClient *minio.Client) error {

	fmt.Println("Start restore files...")

	// read and parse config
	snapshotDate := strings.ReplaceAll(conf.Date, "_", "T") + "Z"
	source := conf.Source
	bucket := conf.Bucket
	target := conf.Target

	// connect to db
	db, err := sql.Open("sqlite", nameDB)
	if err != nil {
		return err
	}

	// get hashes of file's blocks
	hashes, err := getFileBlocks(db, source, snapshotDate)
	if err != nil {
		return err
	}

	// if file not exist

	if len(hashes) == 0 {

		// get files from dir
		files, err := getFilesFromDir(db, source, snapshotDate)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("file not found")
		}

		// download files
		for _, f := range files {

			// get hashes
			fileHashes, err := getFileBlocks(db, f, snapshotDate)
			if err != nil {
				return err
			}

			// download file
			err = downloadFile(minioClient, bucket, f, target, fileHashes)
			if err != nil {
				return err
			}
		}

	} else {

		// download file
		err = downloadFile(minioClient, bucket, source, target, hashes)
		if err != nil {
			return err
		}

	}
	return nil
}

func getFilesFromDir(db *sql.DB, src string, snapDate string) (files []string, err error) {

	// get hashes
	rows, err := db.Query("SELECT f.path_file FROM files f JOIN snapshots s ON f.id_snapshot = s.id WHERE s.timestamp = ? AND f.path_file LIKE ?", snapDate, src+"%")
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

func createFile(pathFile string) (file *os.File, err error) {

	if filepath.IsAbs(pathFile) {

	}
	pathSlice := strings.Split((pathFile), string(os.PathSeparator))

	// create dir
	err = os.MkdirAll(strings.Join(pathSlice[:len(pathSlice)-1], string(os.PathSeparator)), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// create file
	file, err = os.Create(filepath.Join(pathSlice...))
	if err != nil {
		return nil, err
	}

	return file, nil

}

func downloadFile(minioClient *minio.Client, bucket, src, target string, hashes []string) error {

	pathFile, err := createFile(filepath.Join(target, src))
	if err != nil {
		return err
	}

	defer pathFile.Close()

	// download file
	for _, hash := range hashes {

		data, err := minioClient.GetObject(context.Background(), bucket, prefObj+hash, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

		_, err = io.Copy(pathFile, data)
		if err != nil {
			return err
		}
		defer data.Close()
	}

	return nil
}

func getFileBlocks(db *sql.DB, src string, snapDate string) (hashes []string, err error) {

	// get hashes
	rows, err := db.Query("SELECT b.hash FROM blocks b JOIN files f ON f.id_file = b.id_file JOIN snapshots s ON f.id_snapshot = s.id WHERE s.timestamp = ? AND path_file = ? ORDER BY b.block_index ASC;", snapDate, src)
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
