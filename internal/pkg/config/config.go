package config

import (
	"agent-desk/internal/pkg/enums"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Language        string                `yaml:"language"`
	Server          ServerConfig          `yaml:"server"`
	DB              DBConfig              `yaml:"db"`
	Logger          LoggerConfig          `yaml:"logger"`
	Auth            AuthConfig            `yaml:"auth"`
	Storage         StorageConfig         `yaml:"storage"`
	VectorDB        VectorDBConfig        `yaml:"vectorDB"`
	AI              AIConfig              `yaml:"ai"`
	MCP             MCPConfig             `yaml:"mcp"`
	WxWork          WxWorkConfig          `yaml:"wxWork"`
	OIDC            OIDCConfig            `yaml:"oidc"`
	CustomerSession CustomerSessionConfig `yaml:"customerSession"`
	Webhook         WebhookConfig         `yaml:"webhook"`
	Email           EmailConfig           `yaml:"email"`
}

func (c Config) LanguageOrDefault() string {
	switch strings.ToLower(strings.TrimSpace(c.Language)) {
	case "zh", "zh-cn", "zh_cn", "zh-hans":
		return "zh-CN"
	case "en", "en-us", "en_us":
		return "en-US"
	default:
		return "zh-CN"
	}
}

type WxWorkNotifyConfig struct {
	Enabled                bool    `yaml:"enabled"`
	ToUsers                []int64 `yaml:"toUsers"`
	Safe                   bool    `yaml:"safe"`
	EnableDuplicateCheck   bool    `yaml:"enableDuplicateCheck"`
	DuplicateCheckInterval int     `yaml:"duplicateCheckInterval"`
}

type ServerConfig struct {
	Port              int        `yaml:"port"`
	PublicURL         string     `yaml:"publicUrl"`
	CompanyName       string     `yaml:"companyName"`
	CompanyLogoURL    string     `yaml:"companyLogoUrl"`
	CompanyFaviconURL string     `yaml:"companyFaviconUrl"`
	CORS              CORSConfig `yaml:"cors"`
}

func (s ServerConfig) Address() string {
	if s.Port <= 0 {
		return ":8080"
	}
	return fmt.Sprintf(":%d", s.Port)
}

func (s ServerConfig) GetPublicBaseURL(oidcRedirectURL string) string {
	if strings.TrimSpace(s.PublicURL) != "" {
		return strings.TrimRight(strings.TrimSpace(s.PublicURL), "/")
	}
	if strings.TrimSpace(oidcRedirectURL) != "" {
		if u, err := url.Parse(strings.TrimSpace(oidcRedirectURL)); err == nil && u.Scheme != "" && u.Host != "" {
			return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}
	}
	return ""
}

type CORSConfig struct {
	// AllowedOrigins 是允许浏览器跨域访问的 Origin 白名单，必须包含协议和域名。
	// 留空表示不允许跨域请求；同源请求通常不会携带 Origin，不受影响。
	AllowedOrigins []string `yaml:"allowedOrigins"`
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
	PasswordLoginEnabled *bool `yaml:"passwordLoginEnabled"`
	TokenTTLHours        int   `yaml:"tokenTTLHours"`
	MaxFailedAttempts    int   `yaml:"maxFailedAttempts"`
	CredentialLockMinute int   `yaml:"credentialLockMinute"`
}

func (a AuthConfig) IsPasswordLoginEnabled() bool {
	if a.PasswordLoginEnabled == nil {
		return true
	}
	return *a.PasswordLoginEnabled
}

type CustomerSessionConfig struct {
	Secret                  string `yaml:"secret"`
	TTLMinutes              int    `yaml:"ttlMinutes"`
	RefreshThresholdMinutes int    `yaml:"refreshThresholdMinutes"`
}

func (c CustomerSessionConfig) TTL() int {
	if c.TTLMinutes <= 0 {
		return 120
	}
	return c.TTLMinutes
}

func (c CustomerSessionConfig) RefreshThreshold() int {
	if c.RefreshThresholdMinutes <= 0 {
		return 30
	}
	return c.RefreshThresholdMinutes
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
	Type    string                `yaml:"type"`
	Qdrant  QdrantVectorDBConfig  `yaml:"qdrant"`
	LanceDB LanceDBVectorDBConfig `yaml:"lancedb"`
}

type AIConfig struct {
	Provider           string `yaml:"provider"`
	BaseURL            string `yaml:"baseUrl"`
	APIKey             string `yaml:"apiKey"`
	LLMModel           string `yaml:"llmModel"`
	EmbeddingModel     string `yaml:"embeddingModel"`
	EmbeddingDimension int    `yaml:"embeddingDimension"`
	TimeoutMS          int    `yaml:"timeoutMs"`
	MaxRetryCount      int    `yaml:"maxRetryCount"`
}

type QdrantVectorDBConfig struct {
	Host     string `yaml:"host"`
	GrpcPort int    `yaml:"grpcPort"`
	APIKey   string `yaml:"apiKey"`
	UseTLS   bool   `yaml:"useTls"`
}

type LanceDBVectorDBConfig struct {
	Path string `yaml:"path"`
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

type OIDCConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Issuer       string   `yaml:"issuer"`
	ClientID     string   `yaml:"clientId"`
	ClientSecret string   `yaml:"clientSecret"`
	AuthStyle    string   `yaml:"authStyle"`
	RedirectURL  string   `yaml:"redirectUrl"`
	StateSecret  string   `yaml:"stateSecret"`
	Scopes       []string `yaml:"scopes"`
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
	// Notify 为企业微信应用消息通知配置。
	Notify WxWorkNotifyConfig `yaml:"notify"`
}

type WebhookConfig struct {
	OrgSyncSecret    string `yaml:"orgSyncSecret"`
	DOSOrgSyncSecret string `yaml:"dosOrgSyncSecret"`
	OutboundURL      string `yaml:"outboundUrl"`
}

type EmailConfig struct {
	Provider      string `yaml:"provider"`
	FromAddress   string `yaml:"fromAddress"`
	FromName      string `yaml:"fromName"`
	APIKey        string `yaml:"apiKey"`
	SMTPHost      string `yaml:"smtpHost"`
	SMTPPort      int    `yaml:"smtpPort"`
	SMTPUser      string `yaml:"smtpUser"`
	SMTPPassword  string `yaml:"smtpPassword"`
	SMTPUseTLS    bool   `yaml:"smtpUseTls"`
	InboundSecret string `yaml:"inboundSecret"`
}

func Load(path string) (*Config, error) {
	loadDotEnv(path)

	v := viper.New()
	bindConfigDefaults(v)

	if strings.TrimSpace(path) != "" {
		v.SetConfigFile(path)
		v.SetConfigType("yaml")
	}

	v.SetEnvPrefix("AGENT_DESK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindEnvironmentAliases(v)

	if strings.TrimSpace(path) != "" {
		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !os.IsNotExist(err) && !errors.As(err, &configFileNotFoundError) {
				return nil, err
			}
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	normalizeLoadedConfig(cfg)
	return cfg, nil
}

func loadDotEnv(configPath string) {
	if envFile := os.Getenv("AGENT_DESK_ENV_FILE"); envFile != "" {
		_ = gotenv.Load(envFile)
		return
	}
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		_ = gotenv.Load(envFile)
		return
	}
	_ = gotenv.Load(".env")
	_ = gotenv.Load("../.env")
	_ = gotenv.Load("../../.env")
	if configPath != "" {
		dir := filepath.Dir(configPath)
		if dir != "." && dir != "" {
			_ = gotenv.Load(filepath.Join(dir, ".env"))
			_ = gotenv.Load(filepath.Join(dir, "..", ".env"))
		}
	}
}

func bindConfigDefaults(v *viper.Viper) {
	v.SetDefault("language", "zh-CN")
	v.SetDefault("server.port", 8083)
	v.SetDefault("server.publicUrl", "")
	v.SetDefault("server.companyName", "")
	v.SetDefault("server.companyLogoUrl", "")
	v.SetDefault("server.companyFaviconUrl", "")
	v.SetDefault("server.cors.allowedOrigins", []string{})
	v.SetDefault("db.type", "sqlite")
	v.SetDefault("db.dsn", "file:./data/app.db?_busy_timeout=5000")
	v.SetDefault("db.maxIdleConns", 5)
	v.SetDefault("db.maxOpenConns", 20)
	v.SetDefault("db.connMaxIdleTimeSeconds", 300)
	v.SetDefault("db.connMaxLifetimeSeconds", 1800)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "text")
	v.SetDefault("logger.addSource", false)
	v.SetDefault("auth.tokenTTLHours", 12)
	v.SetDefault("auth.maxFailedAttempts", 5)
	v.SetDefault("auth.credentialLockMinute", 15)
	v.SetDefault("customerSession.ttlMinutes", 120)
	v.SetDefault("customerSession.refreshThresholdMinutes", 30)
	v.SetDefault("storage.default", "local")
	v.SetDefault("storage.maxUploadSizeMB", 20)
	v.SetDefault("storage.local.root", "data/storage")
	v.SetDefault("storage.local.baseUrl", "/storage")
	v.SetDefault("vectorDB.type", "qdrant")
	v.SetDefault("vectorDB.qdrant.host", "127.0.0.1")
	v.SetDefault("vectorDB.qdrant.grpcPort", 6334)
	v.SetDefault("ai.provider", "openai")
	v.SetDefault("ai.baseUrl", "https://api.openai.com/v1")
	v.SetDefault("ai.apiKey", "")
	v.SetDefault("ai.llmModel", "gpt-4o-mini")
	v.SetDefault("ai.embeddingModel", "text-embedding-3-small")
	v.SetDefault("ai.embeddingDimension", 1536)
	v.SetDefault("ai.timeoutMs", 30000)
	v.SetDefault("ai.maxRetryCount", 1)
	v.SetDefault("mcp.enabled", true)
	v.SetDefault("email.provider", "smtp")
	v.SetDefault("email.fromAddress", "")
	v.SetDefault("email.fromName", "")
	v.SetDefault("email.apiKey", "")
	v.SetDefault("email.smtpHost", "")
	v.SetDefault("email.smtpPort", 587)
	v.SetDefault("email.smtpUser", "")
	v.SetDefault("email.smtpPassword", "")
	v.SetDefault("email.smtpUseTls", false)
	v.SetDefault("email.inboundSecret", "")
}

func bindEnvironmentAliases(v *viper.Viper) {
	_ = v.BindEnv("server.port", "PORT", "SERVER_PORT", "AGENT_DESK_SERVER_PORT")
	_ = v.BindEnv("server.publicUrl", "PUBLIC_URL", "APP_URL", "SERVER_PUBLIC_URL", "BASE_URL", "DESK_BASE_URL", "AGENT_DESK_SERVER_PUBLICURL")
	_ = v.BindEnv("server.companyName", "COMPANY_NAME", "NEXT_PUBLIC_COMPANY_NAME", "BRAND_NAME", "BRAND_COMPANY_NAME", "AGENT_DESK_SERVER_COMPANYNAME")
	_ = v.BindEnv("server.companyLogoUrl", "COMPANY_LOGO_URL", "NEXT_PUBLIC_COMPANY_LOGO_URL", "BRAND_LOGO_URL", "AGENT_DESK_SERVER_COMPANYLOGOURL")
	_ = v.BindEnv("server.companyFaviconUrl", "COMPANY_FAVICON_URL", "NEXT_PUBLIC_COMPANY_FAVICON_URL", "BRAND_FAVICON_URL", "FAVICON_URL", "AGENT_DESK_SERVER_COMPANYFAVICONURL")
	_ = v.BindEnv("db.type", "DATABASE_TYPE", "DB_TYPE", "AGENT_DESK_DB_TYPE")
	_ = v.BindEnv("db.dsn", "DATABASE_URL", "DB_DSN", "AGENT_DESK_DB_DSN")
	_ = v.BindEnv("auth.passwordLoginEnabled", "PASSWORD_LOGIN_ENABLED", "AGENT_DESK_AUTH_PASSWORDLOGINENABLED")
	_ = v.BindEnv("auth.tokenTTLHours", "AUTH_TOKEN_TTL_HOURS", "AGENT_DESK_AUTH_TOKENTTLHOURS")
	_ = v.BindEnv("customerSession.secret", "CUSTOMER_SESSION_SECRET", "SESSION_SECRET", "JWT_SECRET", "AGENT_DESK_CUSTOMERSESSION_SECRET")
	_ = v.BindEnv("storage.default", "STORAGE_DEFAULT", "STORAGE_TYPE", "AGENT_DESK_STORAGE_DEFAULT")
	_ = v.BindEnv("storage.local.root", "STORAGE_LOCAL_ROOT", "AGENT_DESK_STORAGE_LOCAL_ROOT")
	_ = v.BindEnv("storage.local.baseUrl", "STORAGE_LOCAL_BASE_URL", "AGENT_DESK_STORAGE_LOCAL_BASEURL")
	_ = v.BindEnv("vectorDB.type", "VECTOR_DB_TYPE", "AGENT_DESK_VECTORDB_TYPE")
	_ = v.BindEnv("vectorDB.qdrant.host", "QDRANT_HOST", "AGENT_DESK_VECTORDB_QDRANT_HOST")
	_ = v.BindEnv("vectorDB.qdrant.grpcPort", "QDRANT_GRPC_PORT", "QDRANT_PORT", "AGENT_DESK_VECTORDB_QDRANT_GRPCPORT")
	_ = v.BindEnv("vectorDB.qdrant.apiKey", "QDRANT_API_KEY", "AGENT_DESK_VECTORDB_QDRANT_APIKEY")
	_ = v.BindEnv("ai.provider", "AI_PROVIDER", "OPENAI_PROVIDER", "AGENT_DESK_AI_PROVIDER")
	_ = v.BindEnv("ai.baseUrl", "AI_BASE_URL", "OPENAI_BASE_URL", "OPENAI_API_BASE", "DOS_AI_BASE_URL", "AGENT_DESK_AI_BASEURL")
	_ = v.BindEnv("ai.apiKey", "AI_API_KEY", "OPENAI_API_KEY", "DOS_AI_API_KEY", "CROVE_OPENAI_API_KEY", "AGENT_DESK_AI_APIKEY")
	_ = v.BindEnv("ai.llmModel", "AI_LLM_MODEL", "OPENAI_LLM_MODEL", "OPENAI_MODEL", "LLM_MODEL", "DOS_AI_LLM_MODEL", "AGENT_DESK_AI_LLMMODEL")
	_ = v.BindEnv("ai.embeddingModel", "AI_EMBEDDING_MODEL", "OPENAI_EMBEDDING_MODEL", "EMBEDDING_MODEL", "DOS_AI_EMBEDDING_MODEL", "AGENT_DESK_AI_EMBEDDINGMODEL")
	_ = v.BindEnv("ai.embeddingDimension", "AI_EMBEDDING_DIMENSION", "OPENAI_EMBEDDING_DIMENSION", "EMBEDDING_DIMENSION", "DOS_AI_EMBEDDING_DIMENSION", "AGENT_DESK_AI_EMBEDDINGDIMENSION")
	_ = v.BindEnv("ai.timeoutMs", "AI_TIMEOUT_MS", "OPENAI_TIMEOUT_MS", "AGENT_DESK_AI_TIMEOUTMS")
	_ = v.BindEnv("ai.maxRetryCount", "AI_MAX_RETRY_COUNT", "AGENT_DESK_AI_MAXRETRYCOUNT")
	_ = v.BindEnv("oidc.enabled", "OIDC_ENABLED", "AGENT_DESK_OIDC_ENABLED")
	_ = v.BindEnv("oidc.issuer", "OIDC_ISSUER", "AGENT_DESK_OIDC_ISSUER")
	_ = v.BindEnv("oidc.clientId", "OIDC_CLIENT_ID", "CUSTOM_OAUTH_CLIENT_ID", "AGENT_DESK_OIDC_CLIENTID")
	_ = v.BindEnv("oidc.clientSecret", "OIDC_CLIENT_SECRET", "CUSTOM_OAUTH_CLIENT_SECRET", "AGENT_DESK_OIDC_CLIENTSECRET")
	_ = v.BindEnv("oidc.authStyle", "OIDC_AUTH_STYLE", "CUSTOM_OAUTH_AUTH_STYLE", "AGENT_DESK_OIDC_AUTHSTYLE")
	_ = v.BindEnv("oidc.redirectUrl", "OIDC_REDIRECT_URL", "CUSTOM_OAUTH_REDIRECT_URI", "AGENT_DESK_OIDC_REDIRECTURL")
	_ = v.BindEnv("webhook.orgSyncSecret", "ORG_SYNC_SECRET", "WEBHOOK_SECRET", "AGENT_DESK_WEBHOOK_ORGSYNCSECRET")
	_ = v.BindEnv("webhook.outboundUrl", "ORG_SYNC_OUTBOUND_URL", "DOS_ORG_SYNC_URL", "WEBHOOK_OUTBOUND_URL", "AGENT_DESK_WEBHOOK_OUTBOUNDURL")
	_ = v.BindEnv("mcp.enabled", "MCP_ENABLED", "AGENT_DESK_MCP_ENABLED")
	_ = v.BindEnv("email.provider", "EMAIL_PROVIDER", "AGENT_DESK_EMAIL_PROVIDER")
	_ = v.BindEnv("email.fromAddress", "EMAIL_FROM", "EMAIL_FROM_ADDRESS", "SUPPORT_EMAIL", "AGENT_DESK_EMAIL_FROMADDRESS")
	_ = v.BindEnv("email.fromName", "EMAIL_FROM_NAME", "EMAIL_SENDER_NAME", "SUPPORT_SENDER_NAME", "AGENT_DESK_EMAIL_FROMNAME")
	_ = v.BindEnv("email.apiKey", "EMAIL_API_KEY", "BREVO_API_KEY", "CROVE_BREVO_API_KEY", "SENDGRID_API_KEY", "RESEND_API_KEY", "POSTMARK_API_KEY", "MAILGUN_API_KEY", "AGENT_DESK_EMAIL_APIKEY")
	_ = v.BindEnv("email.smtpHost", "SMTP_HOST", "EMAIL_SMTP_HOST", "AGENT_DESK_EMAIL_SMTPHOST")
	_ = v.BindEnv("email.smtpPort", "SMTP_PORT", "EMAIL_SMTP_PORT", "AGENT_DESK_EMAIL_SMTPPORT")
	_ = v.BindEnv("email.smtpUser", "SMTP_USER", "EMAIL_SMTP_USER", "CROVE_SMTP_USER", "AGENT_DESK_EMAIL_SMTPUSER")
	_ = v.BindEnv("email.smtpPassword", "SMTP_PASSWORD", "SMTP_PASS", "EMAIL_SMTP_PASSWORD", "CROVE_SMTP_PASSWORD", "AGENT_DESK_EMAIL_SMTPPASSWORD")
	_ = v.BindEnv("email.smtpUseTls", "SMTP_USE_TLS", "SMTP_SSL", "AGENT_DESK_EMAIL_SMTPUSETLS")
	_ = v.BindEnv("email.inboundSecret", "EMAIL_INBOUND_SECRET", "EMAIL_WEBHOOK_SECRET", "AGENT_DESK_EMAIL_INBOUNDSECRET")
}

func normalizeLoadedConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.DB.Type == "sqlite" && (strings.HasPrefix(cfg.DB.DSN, "postgres://") || strings.HasPrefix(cfg.DB.DSN, "postgresql://")) {
		cfg.DB.Type = "postgres"
	} else if cfg.DB.Type == "sqlite" && strings.Contains(cfg.DB.DSN, "@tcp(") {
		cfg.DB.Type = "mysql"
	}

	if cfg.MCP.Servers == nil {
		cfg.MCP.Servers = make(map[string]MCPServerConfig)
	}
	if _, ok := cfg.MCP.Servers["system"]; !ok {
		port := cfg.Server.Port
		if port <= 0 {
			port = 8083
		}
		cfg.MCP.Servers["system"] = MCPServerConfig{
			Enabled:   true,
			Endpoint:  fmt.Sprintf("http://127.0.0.1:%d/api/mcp", port),
			TimeoutMS: 15000,
		}
	}

	crmEndpoint := strings.TrimSpace(os.Getenv("MCP_CRM_ENDPOINT"))
	if crmEndpoint == "" {
		crmEndpoint = strings.TrimSpace(os.Getenv("CROVE_CRM_MCP_ENDPOINT"))
	}
	if crmEndpoint == "" {
		crmEndpoint = strings.TrimSpace(os.Getenv("TWENTY_CRM_MCP_ENDPOINT"))
	}
	if crmEndpoint != "" {
		apiKey := strings.TrimSpace(os.Getenv("MCP_CRM_API_KEY"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(os.Getenv("CROVE_CRM_API_KEY"))
		}
		if apiKey == "" {
			apiKey = strings.TrimSpace(os.Getenv("TWENTY_CRM_API_KEY"))
		}
		headers := map[string]string{}
		if apiKey != "" {
			headers["Authorization"] = "Bearer " + apiKey
		}
		crmServerConfig := MCPServerConfig{
			Enabled:   true,
			Endpoint:  crmEndpoint,
			TimeoutMS: 15000,
			Headers:   headers,
		}
		cfg.MCP.Servers["twenty_crm"] = crmServerConfig
		cfg.MCP.Servers["crove_crm"] = crmServerConfig
	}
}
