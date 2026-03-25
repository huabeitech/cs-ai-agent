package services

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"

	"github.com/mlogclub/simple/sqls"
)

func TestCreateAIAgentNormalizesKnowledgeIDs(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}
	seedEnabledAIConfig(t, 101)
	seedEnabledAgentTeam(t, 201)
	seedEnabledKnowledgeBase(t, 301, "知识库A")
	seedEnabledKnowledgeBase(t, 302, "知识库B")

	item, err := AIAgentService.CreateAIAgent(request.CreateAIAgentRequest{
		Name:             "官网客服",
		AIConfigID:       101,
		ServiceMode:      enums.IMConversationServiceModeAIFirst,
		TeamIDs:          []int64{201},
		HandoffMode:      enums.AIAgentHandoffModeDefaultTeamPool,
		MaxAIReplyRounds: 2,
		FallbackMode:     enums.AIAgentFallbackModeGuideRephrase,
		KnowledgeIDs:     []int64{301, 302, 301},
	}, operator)
	if err != nil {
		t.Fatalf("create ai agent failed: %v", err)
	}
	if item.KnowledgeIDs != "301,302" {
		t.Fatalf("expected normalized knowledge ids 301,302, got %s", item.KnowledgeIDs)
	}
	if item.TeamIDs != "201" {
		t.Fatalf("expected normalized team ids 201, got %s", item.TeamIDs)
	}
}

func TestCreateWidgetSiteRejectsDisabledAIAgent(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}
	if err := sqls.DB().Create(&models.AIAgent{
		Name:   "停用客服",
		Status: enums.StatusDisabled,
		AuditFields: models.AuditFields{
			CreateUserID:   operator.UserID,
			CreateUserName: operator.Username,
			UpdateUserID:   operator.UserID,
			UpdateUserName: operator.Username,
		},
	}).Error; err != nil {
		t.Fatalf("seed disabled ai agent failed: %v", err)
	}

	_, err := WidgetSiteService.CreateSite(request.CreateWidgetSiteRequest{
		AIAgentID: 1,
		Name:      "官网站点",
	}, operator)
	if err == nil {
		t.Fatal("expected disabled ai agent rejected")
	}
}

func seedEnabledAIConfig(t *testing.T, id int64) {
	t.Helper()
	item := &models.AIConfig{
		ID:        id,
		Name:      "主配置",
		Provider:  enums.AIProviderOpenAI,
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "test-key",
		ModelType: enums.AIModelTypeLLM,
		ModelName: "gpt-4o-mini",
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("seed ai config failed: %v", err)
	}
	if err := sqls.DB().Model(&models.AIConfig{}).Where("id = ?", id).Update("status", int(enums.StatusOk)).Error; err != nil {
		t.Fatalf("enable ai config failed: %v", err)
	}
}

func seedEnabledAgentTeam(t *testing.T, id int64) {
	t.Helper()
	if err := sqls.DB().Create(&models.AgentTeam{
		ID:     id,
		Name:   "默认客服组",
		Status: enums.StatusOk,
	}).Error; err != nil {
		t.Fatalf("seed agent team failed: %v", err)
	}
}

func seedEnabledKnowledgeBase(t *testing.T, id int64, name string) {
	t.Helper()
	if err := sqls.DB().Create(&models.KnowledgeBase{
		ID:     id,
		Name:   name,
		Status: enums.StatusOk,
	}).Error; err != nil {
		t.Fatalf("seed knowledge base failed: %v", err)
	}
}
