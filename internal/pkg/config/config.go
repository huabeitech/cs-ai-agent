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

// WxWorkConfig 定义企业微信接入配置。
//
// 当前主要用于后台管理台的企业微信登录流程：
// 1. /api/auth/wxwork/login 生成企业微信授权地址
// 2. 企业微信回调到 OAuthRedirect
// 3. 后端通过 code 换取企业成员身份并完成系统登录
//
// 其中 OAuthRedirect、CorpID、CorpSecret、AgentID 为登录流程核心配置。
type WxWorkConfig struct {
	// Enabled 表示是否启用企业微信登录能力。
	// false 时不会初始化企业微信 SDK，相关登录接口不可用。
	Enabled bool `yaml:"enabled"`
	// CorpID 为企业微信公司 ID，例如 wwxxxxxxxxxxxxxxxx。
	CorpID string `yaml:"corpId"`
	// CorpSecret 为企业微信应用 Secret，用于换取 access_token。
	CorpSecret string `yaml:"corpSecret"`
	// AgentID 为企业微信自建应用 AgentID。
	AgentID string `yaml:"agentId"`
	// OAuthRedirect 为企业微信网页授权回调地址。
	// 必须填写完整 URL，且通常指向后端接口 /api/auth/wxwork/callback。
	OAuthRedirect string `yaml:"oauthRedirect"`
	// StateSecret 为登录 state 的签名密钥，用于防止篡改和重放。
	// 建议填写独立随机字符串；留空时业务代码会退回使用 CorpSecret。
	StateSecret string `yaml:"stateSecret"`
	// RSAPrivateKey 为企业微信回调解密私钥。
	// 当前登录流程未使用，保留给消息回调等场景。
	RSAPrivateKey string `yaml:"rsaPrivateKey"`
	// Token 为企业微信回调 Token。
	// 当前登录流程未使用，保留给消息回调等场景。
	Token string `yaml:"token"`
	// EncodingAESKey 为企业微信消息加解密密钥。
	// 当前登录流程未使用，保留给消息回调等场景。
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
