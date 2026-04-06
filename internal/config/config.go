package config

type Config struct {
	Source      string `yaml:"source"`
	Bucket      string `yaml:"bucket"`
	Endpoint    string `yaml:"endpoint"`    //"localhost:9000"
	AccessKeyID string `yaml:"accessKeyId"` //"admin"
	SecretKey   string `yaml:"secretKey"`   //"admin123"
	UseSSL      bool   `yaml:"useSSL"`      //false
}
