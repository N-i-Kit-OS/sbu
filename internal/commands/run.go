package commands

import (
	"diplom/internal/app"
	"diplom/internal/config"
	"diplom/internal/storage"
)

func runRunCommand(path string) error {
	// read config
	cfg, err := config.ReadConfigF(path)
	if err != nil {
		return err
	}

	client, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return err
	}

	return app.RunBackupOrRestore(cfg, client)

}
