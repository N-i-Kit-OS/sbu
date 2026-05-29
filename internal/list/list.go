package list

import (
	_ "modernc.org/sqlite"
)

const (
	nameDB = "sbu.db"
)

type SnapshotInfo struct {
	Name      string
	Timestamp string
}

/*
func GetAllSnapshots(cfg config.Config) ([]SnapshotInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return nil, err
	}

	// get snapshots
	rows, err := db.Query("SELECT name, timestamp FROM snapshots;")
	if err != nil {
		return []SnapshotInfo{}, err
	}
	defer rows.Close()

	var res []SnapshotInfo

	for rows.Next() {

		var name, timestamp string
		err = rows.Scan(&name, &timestamp)
		if err != nil {
			return []SnapshotInfo{}, err
		}

		res = append(res, SnapshotInfo{Name: name, Timestamp: timestamp})
	}
	return res, nil
}

func GetSnapshotFiles(cfg config.Config, snapshotName string) ([]string, error) {

	// connect to s3
	minioClient, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return nil, err
	}

	// download and open db
	db, err := storage.SetupDB(cfg.BackupConfig.Bucket, minioClient)
	if err != nil {
		return nil, err
	}

	// get snapshots
	rows, err := db.Query("SELECT f.path_file FROM files f JOIN snapshots s ON f.id_snapshot = s.id WHERE s.name = ?;", snapshotName)
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
