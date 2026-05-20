package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

func ReadConfigF(cfgPath string) (Config, error) {

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return Config{}, err
	}

	// read config
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, err
	}

	// parse config to struct
	var conf Config

	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return Config{}, err
	}

	return conf, nil
}

func CreateExampleConfig(path string) error {

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
	return os.WriteFile(path, []byte(example), 0644)
}
