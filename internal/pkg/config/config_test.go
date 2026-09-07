package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsCORSAllowedOrigins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`server:
  port: 8083
  cors:
    allowedOrigins:
      - https://console.example.com
      - http://localhost:3000
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got := cfg.Server.CORS.AllowedOrigins
	want := []string{"https://console.example.com", "http://localhost:3000"}
	if len(got) != len(want) {
		t.Fatalf("len(AllowedOrigins)=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllowedOrigins[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestLoadOverridesValuesFromEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`server:
  port: 8083
db:
  type: sqlite
  dsn: file:./data/app.db?_busy_timeout=5000
storage:
  local:
    baseUrl: /storage
mcp:
  servers:
    system:
      endpoint: http://127.0.0.1:8083/api/mcp
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("AGENT_DESK_SERVER_PORT", "8090")
	t.Setenv("AGENT_DESK_DB_DSN", "mysql-dsn")
	t.Setenv("AGENT_DESK_STORAGE_LOCAL_BASEURL", "/files")
	t.Setenv("AGENT_DESK_MCP_SERVERS_SYSTEM_ENDPOINT", "http://127.0.0.1:8090/api/mcp")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8090 {
		t.Fatalf("Server.Port=%d want 8090", cfg.Server.Port)
	}
	if cfg.DB.Type != "sqlite" {
		t.Fatalf("DB.Type=%q want sqlite", cfg.DB.Type)
	}
	if cfg.DB.DSN != "mysql-dsn" {
		t.Fatalf("DB.DSN=%q want mysql-dsn", cfg.DB.DSN)
	}
	if cfg.Storage.Local.BaseURL != "/files" {
		t.Fatalf("Storage.Local.BaseURL=%q want /files", cfg.Storage.Local.BaseURL)
	}
	if cfg.MCP.Servers["system"].Endpoint != "http://127.0.0.1:8090/api/mcp" {
		t.Fatalf("MCP system endpoint=%q", cfg.MCP.Servers["system"].Endpoint)
	}
}

func TestLoadFromDotEnvAndStandardEnvAliases(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	envContent := []byte(`PORT=9090
COMPANY_NAME=CustomDesk
COMPANY_LOGO_URL=/custom-logo.svg
DATABASE_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable
PASSWORD_LOGIN_ENABLED=false
JWT_SECRET=super-secret-key-12345
QDRANT_HOST=10.0.0.5
QDRANT_PORT=6334
OPENAI_API_KEY=sk-test-openai-key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_LLM_MODEL=gpt-4o
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_EMBEDDING_DIMENSION=1536
OIDC_ENABLED=true
OIDC_ISSUER=https://auth.example.com
OIDC_CLIENT_ID=client-123
OIDC_CLIENT_SECRET=secret-456
OIDC_REDIRECT_URL=https://desk.example.com/api/auth/oidc_callback
ORG_SYNC_SECRET=webhook-secret-789
EMAIL_PROVIDER=brevo
EMAIL_FROM=help@example.com
EMAIL_FROM_NAME=Helpdesk Team
BREVO_API_KEY=xkeysib-test-123
EMAIL_INBOUND_SECRET=inbound-secret-456
`)
	if err := os.WriteFile(envPath, envContent, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("ENV_FILE", envPath)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf("Server.Port=%d want 9090", cfg.Server.Port)
	}
	if cfg.Server.CompanyName != "CustomDesk" {
		t.Fatalf("Server.CompanyName=%q want CustomDesk", cfg.Server.CompanyName)
	}
	if cfg.Server.CompanyLogoURL != "/custom-logo.svg" {
		t.Fatalf("Server.CompanyLogoURL=%q want /custom-logo.svg", cfg.Server.CompanyLogoURL)
	}
	if cfg.DB.Type != "postgres" {
		t.Fatalf("DB.Type=%q want postgres", cfg.DB.Type)
	}
	if cfg.DB.DSN != "postgres://user:pass@localhost:5432/mydb?sslmode=disable" {
		t.Fatalf("DB.DSN=%q", cfg.DB.DSN)
	}
	if cfg.Auth.IsPasswordLoginEnabled() {
		t.Fatalf("expected PasswordLoginEnabled to be false")
	}
	if cfg.CustomerSession.Secret != "super-secret-key-12345" {
		t.Fatalf("CustomerSession.Secret=%q", cfg.CustomerSession.Secret)
	}
	if cfg.VectorDB.Qdrant.Host != "10.0.0.5" {
		t.Fatalf("Qdrant.Host=%q", cfg.VectorDB.Qdrant.Host)
	}
	if cfg.VectorDB.Qdrant.GrpcPort != 6334 {
		t.Fatalf("Qdrant.GrpcPort=%d", cfg.VectorDB.Qdrant.GrpcPort)
	}
	if cfg.AI.APIKey != "sk-test-openai-key" {
		t.Fatalf("AI.APIKey=%q", cfg.AI.APIKey)
	}
	if cfg.AI.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("AI.BaseURL=%q", cfg.AI.BaseURL)
	}
	if cfg.AI.LLMModel != "gpt-4o" {
		t.Fatalf("AI.LLMModel=%q", cfg.AI.LLMModel)
	}
	if cfg.AI.EmbeddingModel != "text-embedding-3-small" {
		t.Fatalf("AI.EmbeddingModel=%q", cfg.AI.EmbeddingModel)
	}
	if cfg.AI.EmbeddingDimension != 1536 {
		t.Fatalf("AI.EmbeddingDimension=%d", cfg.AI.EmbeddingDimension)
	}
	if !cfg.OIDC.Enabled {
		t.Fatalf("expected OIDC.Enabled=true")
	}
	if cfg.OIDC.Issuer != "https://auth.example.com" {
		t.Fatalf("OIDC.Issuer=%q", cfg.OIDC.Issuer)
	}
	if cfg.OIDC.ClientID != "client-123" {
		t.Fatalf("OIDC.ClientID=%q", cfg.OIDC.ClientID)
	}
	if cfg.OIDC.ClientSecret != "secret-456" {
		t.Fatalf("OIDC.ClientSecret=%q", cfg.OIDC.ClientSecret)
	}
	if cfg.OIDC.RedirectURL != "https://desk.example.com/api/auth/oidc_callback" {
		t.Fatalf("OIDC.RedirectURL=%q", cfg.OIDC.RedirectURL)
	}
	if cfg.Webhook.OrgSyncSecret != "webhook-secret-789" {
		t.Fatalf("Webhook.OrgSyncSecret=%q", cfg.Webhook.OrgSyncSecret)
	}
	if cfg.Email.Provider != "brevo" {
		t.Fatalf("Email.Provider=%q want brevo", cfg.Email.Provider)
	}
	if cfg.Email.FromAddress != "help@example.com" {
		t.Fatalf("Email.FromAddress=%q want help@example.com", cfg.Email.FromAddress)
	}
	if cfg.Email.FromName != "Helpdesk Team" {
		t.Fatalf("Email.FromName=%q want Helpdesk Team", cfg.Email.FromName)
	}
	if cfg.Email.APIKey != "xkeysib-test-123" {
		t.Fatalf("Email.APIKey=%q want xkeysib-test-123", cfg.Email.APIKey)
	}
	if cfg.Email.InboundSecret != "inbound-secret-456" {
		t.Fatalf("Email.InboundSecret=%q want inbound-secret-456", cfg.Email.InboundSecret)
	}
}
