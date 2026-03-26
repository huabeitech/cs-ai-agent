package config

import (
	"cs-agent/internal/pkg/enums"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	DB       DBConfig       `yaml:"db"`
	Logger   LoggerConfig   `yaml:"logger"`
	Auth     AuthConfig     `yaml:"auth"`
	Storage  StorageConfig  `yaml:"storage"`
	VectorDB VectorDBConfig `yaml:"vectorDB"`
	MCP      MCPConfig      `yaml:"mcp"`
	WxWork   WxWorkConfig   `yaml:"wxWork"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func (s ServerConfig) Address() string {
	if s.Port <= 0 {
		return ":8080"
	}
	return fmt.Sprintf(":%d", s.Port)
}

type DBConfig struct {
	Type                   string `yaml:"type"`
	DSN                    string `yaml:"dsn"`
	MaxIdleConns           int    `yaml:"maxIdleConns"`
	MaxOpenConns           int    `yaml:"maxOpenConns"`
	ConnMaxIdleTimeSeconds int    `yaml:"connMaxIdleTimeSeconds"`
	ConnMaxLifetimeSeconds int    `yaml:"connMaxLifetimeSeconds"`
}

type LoggerConfig struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	AddSource bool   `yaml:"addSource"`
}

type AuthConfig struct {
	AccessTokenTTLHours  int `yaml:"accessTokenTTLHours"`
	RefreshTokenTTLDays  int `yaml:"refreshTokenTTLDays"`
	MaxFailedAttempts    int `yaml:"maxFailedAttempts"`
	CredentialLockMinute int `yaml:"credentialLockMinute"`
}

type StorageConfig struct {
	Default         enums.AssetProvider `yaml:"default"`
	MaxUploadSizeMB int64               `yaml:"maxUploadSizeMB"`
	Local           LocalStorageConfig  `yaml:"local"`
	OSS             OSSStorageConfig    `yaml:"oss"`
}

func (s StorageConfig) MaxUploadSizeBytes() int64 {
	if s.MaxUploadSizeMB <= 0 {
		return 5 << 20
	}
	return s.MaxUploadSizeMB << 20
}

func (s StorageConfig) MaxRequestBodySizeBytes() int64 {
	limit := s.MaxUploadSizeBytes()
	return limit + (1 << 20)
}

type LocalStorageConfig struct {
	Root    string `yaml:"root"`
	BaseURL string `yaml:"baseUrl"`
}

type OSSStorageConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Bucket          string `yaml:"bucket"`
	AccessKeyID     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	BaseURL         string `yaml:"baseUrl"`
	Private         bool   `yaml:"private"`
	SignedURLExpire int    `yaml:"signedUrlExpireSeconds"`
}

type VectorDBConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	APIKey   string `yaml:"apiKey"`
	GrpcPort int    `yaml:"grpcPort"`
	UseTLS   bool   `yaml:"useTls"`
}

type MCPConfig struct {
	Enabled bool                       `yaml:"enabled"`
	Servers map[string]MCPServerConfig `yaml:"servers"`
}

type MCPServerConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Endpoint  string            `yaml:"endpoint"`
	TimeoutMS int               `yaml:"timeoutMs"`
	Headers   map[string]string `yaml:"headers"`
}

type WxWorkConfig struct {
	Enabled        bool   `yaml:"enabled"`
	CorpID         string `yaml:"corpId"`
	CorpSecret     string `yaml:"corpSecret"`
	AgentID        string `yaml:"agentId"`
	RsaPrivateKey  string `yaml:"rsaPrivateKey"`
	Token          string `yaml:"token"`
	EncodingAESKey string `yaml:"encodingAESKey"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
