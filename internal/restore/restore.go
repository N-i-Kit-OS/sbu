package restore

import (
	"context"
	"diplom/internal/config"
	"diplom/internal/constants"
	"diplom/internal/sbudb"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func Restore(conf config.RestoreConfig, client *minio.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	snapshotDate := strings.ReplaceAll(conf.Date, "_", "T") + "Z"

	if err := sbudb.DownloadFromS3(ctx, conf.Bucket, client); err != nil {
		return fmt.Errorf(" DB exist but not downloaded: %w", err)
	}

	db, err := sbudb.OpenLocal()
	if err != nil {
		return fmt.Errorf("failed to open local DB: %w", err)
	}
	defer db.Close()

	files, err := sbudb.GetFilesByPrefix(ctx, db, snapshotDate, conf.Source)
	if err != nil {
		return fmt.Errorf("failed to get files by source: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("files not found by source: %s", conf.Source)
	}

	for _, f := range files {
		hashes, err := sbudb.GetBlockHashesByPath(ctx, db, snapshotDate, f)
		if err != nil {
			return fmt.Errorf("failed to get block hashes by path: %w", err)
		}

		if err := downloadFile(ctx, client, conf.Bucket, conf.Target, f, hashes); err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
	}
	fmt.Println(" Restore completed")
	return nil
}

func createFile(pathFile string) (file *os.File, err error) {
	pathSlice := strings.Split((pathFile), string(os.PathSeparator))

	err = os.MkdirAll(strings.Join(pathSlice[:len(pathSlice)-1], string(os.PathSeparator)), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create dir:%s, %w", pathSlice[:len(pathSlice)-1], err)
	}

	file, err = os.Create(filepath.Join(pathSlice...))
	if err != nil {
		return nil, fmt.Errorf("failed to create file:%s, %w", pathFile, err)
	}

	return file, nil
}

func downloadFile(ctx context.Context, client *minio.Client, bucket, targetPath, sourcePath string, hashes []string) error {
	fullPath := filepath.Join(targetPath, sourcePath)
	file, err := createFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, hash := range hashes {
		obj, err := client.GetObject(ctx, bucket, constants.ObjectPrefix+hash, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get object from S3: %w", err)
		}

		_, err = io.Copy(file, obj)
		obj.Close()

		if err != nil {
			return fmt.Errorf("failed to copy object from S3: %w", err)
		}
	}

	return nil
}

/*
func Restore(conf config.RestoreConfig, client *minio.Client) error {

		fmt.Println(conf)
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
*/
/*
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
*/
