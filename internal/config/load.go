package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

func Load(cfgPath string) (Config, error) {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file not found: %w", err)
	}

	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var conf Config
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return conf, nil
}

func Init(path string) error {
	example := `# Режим работы: backup или restore
mode: backup

# Настройки подключения к S3
s3:
  endpoint: "s3.cloud.ru"
  accessKeyId: "tenant:access_key"
  secretKey: "your_secret_key"
  useSSL: true
  region: "ru-central-1"

# Настройки для бэкапа (если mode: backup)
backup:
  source: "/home/user/documents"
  name: "my_snapshot_2026"
  bucket: "my-backups"

# Настройки для восстановления (если mode: restore)
restore:
  source: "oldFiles/documents"
  target: "./restored"
  date: "2026-04-24_22:48:47"
  name: "my_snapshot_2026"
  bucket: "my-backups"
`
	if err := os.WriteFile(path, []byte(example), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}
