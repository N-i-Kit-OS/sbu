package commands

import (
	"diplom/internal/app"
	"diplom/internal/config"
	"diplom/internal/storage"
	"fmt"
)

func handleSLS(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return fmt.Errorf("failed to connect to S3: %w", err)
	}

	return app.GetAllSnapshots(client, cfg.RestoreConfig.Bucket)
}
