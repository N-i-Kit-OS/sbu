package config

type Config struct {
	Source         string `yaml:"source"`
	Bucket         string `yaml:"bucket"`
	Endpoint       string `yaml:"endpoint"`
	AccessKeyID    string `yaml:"accessKeyId"`
	SecretKey      string `yaml:"secretKey"`
	UseSSL         bool   `yaml:"useSSL"`
	FromRecovery   string `yaml:"fromRecovery"`
	DateRecovery   string `yaml:"dateRecovery"`
	PathToRecovery string `yaml:"pathToRecovery"`
}
