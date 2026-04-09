package aiagent

import (
	"cs-agent/cmd/testdata/skill"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"fmt"
	"time"

	"github.com/mlogclub/simple/sqls"
)

type InitResult struct {
	Created int
	Updated int
}

// Init 初始化 AI Agent 测试数据
// 依赖于 AI Config 和 Knowledge Base 已初始化
func Init() (*InitResult, error) {
	result := &InitResult{}

	aiConfigID, err := getDefaultAIConfigID()
	if err != nil {
		return result, fmt.Errorf("get default ai config id failed: %w", err)
	}
	if aiConfigID == 0 {
		return result, fmt.Errorf("no default ai config found, please init ai config first")
	}

	knowledgeIDs, err := getDefaultKnowledgeIDs()
	if err != nil {
		return result, fmt.Errorf("get default knowledge ids failed: %w", err)
	}

	defaultTeamIDs := getDefaultTeamIDs()
	defaultSkillIDs, err := getDefaultSkillIDs()
	if err != nil {
		return result, fmt.Errorf("get default skill ids failed: %w", err)
	}

	seedItems := buildSeedItems(aiConfigID, knowledgeIDs, defaultTeamIDs, defaultSkillIDs)
	for _, item := range seedItems {
		itemCopy := item
		if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			existing := repositories.AIAgentRepository.Take(ctx.Tx, "name = ?", itemCopy.Name)
			if existing != nil {
				// 更新
				if err := ctx.Tx.Model(existing).Updates(&itemCopy).Error; err != nil {
					return err
				}
				result.Updated++
			} else {
				// 创建
				if err := ctx.Tx.Create(&itemCopy).Error; err != nil {
					return err
				}
				result.Created++
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("upsert ai agent failed: %w", err)
		}
	}

	return result, nil
}

func buildSeedItems(aiConfigID int64, knowledgeIDs []int64, defaultTeamIDs string, defaultSkillIDs string) []models.AIAgent {
	now := time.Now()
	return []models.AIAgent{
		{
			Name:                "测试AI客服",
			Description:         "本地测试 AI 客服 Agent",
			Status:              enums.StatusOk,
			AIConfigID:          aiConfigID,
			ServiceMode:         enums.IMConversationServiceModeAIFirst,
			SystemPrompt:        "你是一个友好的客服助手，请用中文回答用户的问题。",
			WelcomeMessage:      "您好，欢迎咨询！有什么可以帮助您的？",
			ReplyTimeoutSeconds: 180,
			TeamIDs:             defaultTeamIDs,
			HandoffMode:         enums.AIAgentHandoffModeWaitPool,
			MaxAIReplyRounds:    0,
			FallbackMode:        enums.AIAgentFallbackModeGuideRephrase,
			FallbackMessage:     "我暂时没有找到足够准确的信息。你可以补充订单号、产品名或更具体的问题，我再继续帮你查。",
			KnowledgeIDs:        utils.JoinInt64s(knowledgeIDs),
			SkillIDs:            defaultSkillIDs,
			SortNo:              10,
			Remark:              "Local testdata seed",
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   0,
				CreateUserName: "System",
				UpdatedAt:      now,
				UpdateUserID:   0,
				UpdateUserName: "System",
			},
		},
	}
}

func getDefaultAIConfigID() (int64, error) {
	aiConfig := repositories.AIConfigRepository.Take(
		sqls.DB(),
		"model_type = ? AND status = ?",
		string(enums.AIModelTypeLLM),
		enums.StatusOk,
	)
	if aiConfig == nil {
		return 0, nil
	}
	return aiConfig.ID, nil
}

func getDefaultKnowledgeIDs() ([]int64, error) {
	knowledges := repositories.KnowledgeBaseRepository.Find(
		sqls.DB(),
		sqls.NewCnd().Where("status = ?", enums.StatusOk),
	)
	ids := make([]int64, 0, len(knowledges))
	for _, knowledge := range knowledges {
		ids = append(ids, knowledge.ID)
	}
	return ids, nil
}

func getDefaultTeamIDs() string {
	teams := repositories.AgentTeamRepository.Find(
		sqls.DB(),
		sqls.NewCnd().Where("status = ?", enums.StatusOk),
	)
	teamIDs := make([]int64, 0, len(teams))
	for _, team := range teams {
		teamIDs = append(teamIDs, team.ID)
	}
	return utils.JoinInt64s(teamIDs)
}

func getDefaultSkillIDs() (string, error) {
	skillItem := repositories.SkillDefinitionRepository.Take(
		sqls.DB(),
		"code = ? AND status = ?",
		skill.TestGreetingSkillCode,
		enums.StatusOk,
	)
	if skillItem == nil {
		return "", fmt.Errorf("default test skill not found: %s", skill.TestGreetingSkillCode)
	}
	return utils.JoinInt64s([]int64{skillItem.ID}), nil
}
