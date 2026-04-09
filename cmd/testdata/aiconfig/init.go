package aiconfig

import (
	"fmt"
	"os"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const (
	defaultSeedFilePath = "cmd/testdata/aiconfig/ai_config.yaml"
	seedFileEnvKey      = "TESTDATA_AI_CONFIG_FILE"
)

type SeedItem struct {
	Name             string `yaml:"name"`
	Provider         string `yaml:"provider"`
	BaseURL          string `yaml:"baseUrl"`
	APIKey           string `yaml:"apiKey"`
	ModelType        string `yaml:"modelType"`
	ModelName        string `yaml:"modelName"`
	Dimension        int    `yaml:"dimension"`
	MaxContextTokens int    `yaml:"maxContextTokens"`
	MaxOutputTokens  int    `yaml:"maxOutputTokens"`
	TimeoutMS        int    `yaml:"timeoutMs"`
	MaxRetryCount    int    `yaml:"maxRetryCount"`
	RPMLimit         int    `yaml:"rpmLimit"`
	TPMLimit         int    `yaml:"tpmLimit"`
	SortNo           int    `yaml:"sortNo"`
	Remark           string `yaml:"remark"`
}

type InitResult struct {
	FilePath string
	Skipped  bool
	Created  int
	Updated  int
}

type seedPayload struct {
	Items []SeedItem `yaml:"items"`
}

func Init() (*InitResult, error) {
	return InitFromFile(resolveSeedFilePath())
}

func InitFromFile(filePath string) (*InitResult, error) {
	result := &InitResult{FilePath: filePath}

	payload, err := loadSeedPayload(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Skipped = true
			return result, nil
		}
		return nil, err
	}
	if len(payload.Items) == 0 {
		result.Skipped = true
		return result, nil
	}

	err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for _, item := range payload.Items {
			created, upsertErr := upsertAIConfig(ctx.Tx, item)
			if upsertErr != nil {
				return upsertErr
			}
			if created {
				result.Created++
			} else {
				result.Updated++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func loadSeedPayload(filePath string) (*seedPayload, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	payload := &seedPayload{}
	if err := yaml.Unmarshal(content, payload); err != nil {
		return nil, fmt.Errorf("parse ai config seed file failed: %w", err)
	}
	return payload, nil
}

func resolveSeedFilePath() string {
	if filePath := strings.TrimSpace(os.Getenv(seedFileEnvKey)); filePath != "" {
		return filePath
	}
	return defaultSeedFilePath
}

func upsertAIConfig(db *gorm.DB, item SeedItem) (bool, error) {
	model, err := buildAIConfigModel(item)
	if err != nil {
		return false, err
	}

	now := time.Now()
	existing := repositories.AIConfigRepository.FindOne(db, sqls.NewCnd().Eq("name", model.Name).Eq("model_type", model.ModelType))
	if existing == nil {
		model.AuditFields = models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   constants.SystemAuditUserID,
			CreateUserName: constants.SystemAuditUserName,
			UpdatedAt:      now,
			UpdateUserID:   constants.SystemAuditUserID,
			UpdateUserName: constants.SystemAuditUserName,
		}
		if err := repositories.AIConfigRepository.Create(db, model); err != nil {
			return false, err
		}
		return true, nil
	}

	err = repositories.AIConfigRepository.Updates(db, existing.ID, map[string]any{
		"provider":           model.Provider,
		"base_url":           model.BaseURL,
		"api_key":            model.APIKey,
		"model_name":         model.ModelName,
		"dimension":          model.Dimension,
		"max_context_tokens": model.MaxContextTokens,
		"max_output_tokens":  model.MaxOutputTokens,
		"timeout_ms":         model.TimeoutMS,
		"max_retry_count":    model.MaxRetryCount,
		"rpm_limit":          model.RPMLimit,
		"tpm_limit":          model.TPMLimit,
		"status":             model.Status,
		"sort_no":            model.SortNo,
		"remark":             model.Remark,
		"update_user_id":     constants.SystemAuditUserID,
		"update_user_name":   constants.SystemAuditUserName,
		"updated_at":         now,
	})
	if err != nil {
		return false, err
	}
	return false, nil
}

func buildAIConfigModel(item SeedItem) (*models.AIConfig, error) {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		return nil, fmt.Errorf("ai config seed name is required")
	}

	provider, err := parseProvider(item.Provider)
	if err != nil {
		return nil, fmt.Errorf("ai config %s: %w", name, err)
	}
	modelType, err := parseModelType(item.ModelType)
	if err != nil {
		return nil, fmt.Errorf("ai config %s: %w", name, err)
	}

	baseURL := strings.TrimSpace(item.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("ai config %s: baseUrl is required", name)
	}
	modelName := strings.TrimSpace(item.ModelName)
	if modelName == "" {
		return nil, fmt.Errorf("ai config %s: modelName is required", name)
	}

	return &models.AIConfig{
		Name:             name,
		Provider:         provider,
		BaseURL:          baseURL,
		APIKey:           strings.TrimSpace(item.APIKey),
		ModelType:        modelType,
		ModelName:        modelName,
		Dimension:        normalizeNonNegative(item.Dimension),
		MaxContextTokens: normalizeNonNegative(item.MaxContextTokens),
		MaxOutputTokens:  normalizeNonNegative(item.MaxOutputTokens),
		TimeoutMS:        defaultIfNotPositive(item.TimeoutMS, 30000),
		MaxRetryCount:    normalizeNonNegative(item.MaxRetryCount),
		RPMLimit:         normalizeNonNegative(item.RPMLimit),
		TPMLimit:         normalizeNonNegative(item.TPMLimit),
		SortNo:           normalizeNonNegative(item.SortNo),
		Status:           enums.StatusOk,
		Remark:           strings.TrimSpace(item.Remark),
	}, nil
}

func parseProvider(raw string) (enums.AIProvider, error) {
	provider := enums.AIProvider(strings.TrimSpace(raw))
	switch provider {
	case enums.AIProviderOpenAI:
		return provider, nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", raw)
	}
}

func parseModelType(raw string) (enums.AIModelType, error) {
	modelType := enums.AIModelType(strings.TrimSpace(raw))
	switch modelType {
	case enums.AIModelTypeLLM, enums.AIModelTypeEmbedding, enums.AIModelTypeRerank:
		return modelType, nil
	default:
		return "", fmt.Errorf("unsupported modelType: %s", raw)
	}
}

func normalizeNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func defaultIfNotPositive(value int, def int) int {
	if value > 0 {
		return value
	}
	return def
}
