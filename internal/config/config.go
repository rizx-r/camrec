package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Recorder RecorderConfig `yaml:"recorder"`
	MinIO    MinioConfig    `yaml:"minio"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type ServerConfig struct {
	Addr               string `yaml:"addr"`
	PresignExpireSec   int64  `yaml:"presign_expire_seconds"`
	PublicBucketPolicy bool   `yaml:"public_bucket_policy"`
}

type RecorderConfig struct {
	FfmpegPath     string `yaml:"ffmpeg_path"`
	RTSPURL        string `yaml:"rtsp_url"`
	OutputDir      string `yaml:"output_dir"`
	SegmentSeconds int    `yaml:"segment_seconds"`
}

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
	Region    string `yaml:"region"`
}

type PostgresConfig struct {
	DSN string `yaml:"dsn"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.Recorder.OutputDir == "" {
		c.Recorder.OutputDir = "data"
	}
	if c.Recorder.SegmentSeconds == 0 {
		c.Recorder.SegmentSeconds = 600
	}
	if c.Server.PresignExpireSec == 0 {
		c.Server.PresignExpireSec = 3600
	}
	return &c, nil
}
