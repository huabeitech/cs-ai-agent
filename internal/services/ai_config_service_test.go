package services

import (
	"testing"

	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
)

func TestCreateAIConfigDisablesOtherEnabledConfigsOfSameModelType(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}

	first, err := AIConfigService.CreateAIConfig(request.CreateAIConfigRequest{
		Name:      "LLM A",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		ModelType: enums.AIModelTypeLLM,
		ModelName: "gpt-4o-mini",
	}, operator)
	if err != nil {
		t.Fatalf("create first config failed: %v", err)
	}

	second, err := AIConfigService.CreateAIConfig(request.CreateAIConfigRequest{
		Name:      "LLM B",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		ModelType: enums.AIModelTypeLLM,
		ModelName: "gpt-4.1-mini",
	}, operator)
	if err != nil {
		t.Fatalf("create second config failed: %v", err)
	}

	reloadedFirst := AIConfigService.Get(first.ID)
	reloadedSecond := AIConfigService.Get(second.ID)
	if reloadedFirst == nil || reloadedSecond == nil {
		t.Fatal("expected configs to exist")
	}
	if reloadedFirst.Status != enums.StatusDisabled {
		t.Fatalf("expected first config disabled, got %d", reloadedFirst.Status)
	}
	if reloadedSecond.Status != enums.StatusOk {
		t.Fatalf("expected second config enabled, got %d", reloadedSecond.Status)
	}
}

func TestUpdateAIConfigStatusDisablesOtherEnabledConfigsOfSameModelType(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}

	first, err := AIConfigService.CreateAIConfig(request.CreateAIConfigRequest{
		Name:      "Embedding A",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		ModelType: enums.AIModelTypeEmbedding,
		ModelName: "text-embedding-3-small",
	}, operator)
	if err != nil {
		t.Fatalf("create first config failed: %v", err)
	}

	second, err := AIConfigService.CreateAIConfig(request.CreateAIConfigRequest{
		Name:      "Embedding B",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		ModelType: enums.AIModelTypeEmbedding,
		ModelName: "text-embedding-3-large",
	}, operator)
	if err != nil {
		t.Fatalf("create second config failed: %v", err)
	}

	if err := AIConfigService.UpdateStatus(second.ID, enums.StatusOk, operator); err != nil {
		t.Fatalf("enable second config failed: %v", err)
	}

	reloadedFirst := AIConfigService.Get(first.ID)
	reloadedSecond := AIConfigService.Get(second.ID)
	if reloadedFirst == nil || reloadedSecond == nil {
		t.Fatal("expected configs to exist")
	}
	if reloadedFirst.Status != enums.StatusDisabled {
		t.Fatalf("expected first config disabled, got %d", reloadedFirst.Status)
	}
	if reloadedSecond.Status != enums.StatusOk {
		t.Fatalf("expected second config enabled, got %d", reloadedSecond.Status)
	}
}

func TestDeleteAIConfigRejectsEnabledConfig(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}

	item, err := AIConfigService.CreateAIConfig(request.CreateAIConfigRequest{
		Name:      "Rerank A",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		ModelType: enums.AIModelTypeRerank,
		ModelName: "rerank-v1",
	}, operator)
	if err != nil {
		t.Fatalf("create config failed: %v", err)
	}

	if err := AIConfigService.DeleteAIConfig(item.ID); err == nil {
		t.Fatal("expected enabled config delete rejected")
	}
}
