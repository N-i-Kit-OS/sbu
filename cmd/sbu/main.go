package main

import (
	"context"
	"diplom/internal/config"
	"diplom/internal/restore"
	"diplom/internal/storage"
	"encoding/json"
	"fmt"
	"net/http"
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

	case "ui":
		http.HandleFunc("/api/connect", handleConnect)
		http.HandleFunc("/api/backup", handleBackup)
		http.HandleFunc("/api/snapshots", handleSnapshots)
		http.HandleFunc("/api/restore", handleRestore)
		http.HandleFunc("/api/snapshot-files", handleSnapshotFiles)

		http.Handle("/", http.FileServer(http.Dir("./ui")))

		fmt.Println("Starting UI server at http://localhost:8010")
		err := http.ListenAndServe(":8010", nil)
		if err != nil {
			return err
		}

	default:
		printUsage()
	}

	return nil
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint    string `json:"endpoint"`
		AccessKeyID string `json:"accessKeyId"`
		SecretKey   string `json:"secretKey"`
		UseSSL      bool   `json:"useSSL"`
		Region      string `json:"region"`
	}

	JSONDec := json.NewDecoder(r.Body)
	err := JSONDec.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := config.S3Config{
		Endpoint:    req.Endpoint,
		AccessKeyID: req.AccessKeyID,
		SecretKey:   req.SecretKey,
		UseSSL:      req.UseSSL,
		Region:      req.Region,
	}

	minioClient, err := storage.ConnectToS3(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
func handleBackup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source string          `json:"source"`
		Name   string          `json:"name"`
		Bucket string          `json:"bucket"`
		S3     config.S3Config `json:"s3"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connect to S3
	minioClient, err := storage.ConnectToS3(req.S3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create backup config
	backupCfg := config.BackupConfig{
		Source:       req.Source,
		SnapshotName: req.Name,
		Bucket:       req.Bucket,
	}

	// Run backup
	if err := storage.Backup(backupCfg, minioClient); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Backup completed"})
}
func handleSnapshots(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bucket string          `json:"bucket"`
		S3     config.S3Config `json:"s3"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connect to S3
	minioClient, err := storage.ConnectToS3(req.S3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := storage.SetupDB(req.Bucket, minioClient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT name, timestamp FROM snapshots")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var snapshots []map[string]string
	for rows.Next() {
		var name, ts string
		rows.Scan(&name, &ts)
		snapshots = append(snapshots, map[string]string{"name": name, "timestamp": ts})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)

}
func handleRestore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Snapshot string          `json:"snapshot"`
		Target   string          `json:"target"`
		Bucket   string          `json:"bucket"`
		Source   string          `json:"source"`
		Date     string          `json:"date"`
		S3       config.S3Config `json:"s3"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client, err := storage.ConnectToS3(req.S3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	restoreCfg := config.RestoreConfig{
		Source:       req.Source,
		SnapshotName: req.Snapshot,
		Target:       req.Target,
		Bucket:       req.Bucket,
		Date:         req.Date,
	}

	fmt.Println(restoreCfg)
	if err := restore.Restore(restoreCfg, client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Restore completed"})
}
func handleSnapshotFiles(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Snapshot string          `json:"snapshot"`
		Bucket   string          `json:"bucket"`
		S3       config.S3Config `json:"s3"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	minioClient, err := storage.ConnectToS3(req.S3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := storage.SetupDB(req.Bucket, minioClient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем id снапшота по имени
	var snapshotID int
	err = db.QueryRow("SELECT id FROM snapshots WHERE name = ?", req.Snapshot).Scan(&snapshotID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем все файлы для этого снапшота
	rows, err := db.Query("SELECT path_file FROM files WHERE id_snapshot = ?", snapshotID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var path string
		rows.Scan(&path)
		files = append(files, path)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
func printUsage() {
	fmt.Println("Usage: sbu <command> [config_path]")
	fmt.Println("Commands:")
	fmt.Println("  init [config_path]       create example config (default: config.yaml)")
	fmt.Println("  run  [config_path]       run backup or restore (default: config.yaml)")
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
  date: "2026-04-24_22:48:47"
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
