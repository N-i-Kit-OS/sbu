package config

type S3Config struct {
	Endpoint    string `yaml:"endpoint"`
	AccessKeyID string `yaml:"accessKeyId"`
	SecretKey   string `yaml:"secretKey"`
	UseSSL      bool   `yaml:"useSSL"`
	Region      string `yaml:"region"`
}

type RestoreConfig struct {
	Source       string   `yaml:"source"`
	Date         string   `yaml:"date"`
	Target       string   `yaml:"target"`
	SnapshotName string   `yaml:"name"`
	Bucket       string   `yaml:"bucket"`
	Ignore       []string `yaml:"ignore"`
}

type BackupConfig struct {
	SnapshotName string   `yaml:"name"`
	Source       string   `yaml:"source"`
	Bucket       string   `yaml:"bucket"`
	Ignore       []string `yaml:"ignore"`
}

type Config struct {
	Mode          string        `yaml:"mode"`
	S3Config      S3Config      `yaml:"s3"`
	RestoreConfig RestoreConfig `yaml:"restore"`
	BackupConfig  BackupConfig  `yaml:"backup"`
}
