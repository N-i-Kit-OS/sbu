package commands

import (
	"diplom/internal/app"
	"diplom/internal/config"
	"diplom/internal/storage"
	"fmt"
)

func handleRun(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	client, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}

	return app.RunBackupOrRestore(cfg, client)
}
