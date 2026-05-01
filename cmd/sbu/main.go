package main

import (
	"diplom/internal/config"
	"diplom/internal/restore"
	"diplom/internal/storage"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

func main() {

	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {

	// check flags
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	// check command
	command := os.Args[1]

	// find config
	var configPath string
	if len(os.Args) > 2 {
		configPath = os.Args[2]
	} else {
		configPath = "config.yml"
	}

	switch command {

	case "init":
		return createConfigExample(configPath)

	case "run":

		// read config
		cfg, err := readConfigF(configPath)
		if err != nil {
			return err
		}

		return runBackupOrRestore(cfg)

	default:
		printUsage()
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: sbu <command> [config_path]")
	fmt.Println("Commands:")
	fmt.Println("  init [config_path]    create example config (default: config.yaml)")
	fmt.Println("  run  [config_path]    run backup or restore (default: config.yaml)")
}

func runBackupOrRestore(cfg config.Config) error {

	minioClient, err := storage.ConnectToS3(cfg.S3Config)
	if err != nil {
		return err
	}

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

func createConfigExample(cfgPath string) error {
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
  ignore:
    - "*.tmp"
    - ".git"

# Настройки для восстановления (если mode: restore)
restore:
  source: "oldFiles/documents"
  target: "./restored"
  date: "2026-04-24_22-48-47"
  name: "my_snapshot_2026"
  bucket: "my-backups"
  ignore:
    - "temp"
`
	return os.WriteFile(cfgPath, []byte(example), 0644)
}

func readConfigF(cfgPath string) (config.Config, error) {

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return config.Config{}, err
	}

	// read config
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return config.Config{}, err
	}

	// parse config to struct
	var conf config.Config

	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return config.Config{}, err
	}

	return conf, nil
}
