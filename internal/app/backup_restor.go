package app

import (
	"diplom/internal/config"
	"diplom/internal/restore"
	"diplom/internal/storage"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func RunBackupOrRestore(cfg config.Config, minioClient *minio.Client) error {

	switch cfg.Mode {
	case "backup":

		// upload directory from config to s3
		err := storage.Backup(cfg.BackupConfig, minioClient)
		if err != nil {
			return err
		}

	case "restore":

		// download files from s3 to config path
		err := restore.Restore(cfg.RestoreConfig, minioClient)
		if err != nil {
			return err
		}

	default:
		fmt.Println("The operating mode is not specified or is specified incorrectly")
		return nil
	}

	return nil
}
