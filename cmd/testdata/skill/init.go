package skill

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"time"

	"github.com/mlogclub/simple/sqls"
)

const TestGreetingSkillCode = "testdata_greeting_skill"

type InitResult struct {
	Created int
	Updated int
}

func Init() (*InitResult, error) {
	result := &InitResult{}
	seedItems := buildSeedItems()
	for _, item := range seedItems {
		itemCopy := item
		if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			existing := repositories.SkillDefinitionRepository.Take(ctx.Tx, "code = ?", itemCopy.Code)
			if existing != nil {
				if err := ctx.Tx.Model(existing).Updates(&itemCopy).Error; err != nil {
					return err
				}
				result.Updated++
				return nil
			}
			if err := ctx.Tx.Create(&itemCopy).Error; err != nil {
				return err
			}
			result.Created++
			return nil
		}); err != nil {
			return nil, fmt.Errorf("upsert skill failed: %w", err)
		}
	}
	return result, nil
}

func buildSeedItems() []models.SkillDefinition {
	now := time.Now()
	return []models.SkillDefinition{
		{
			Code:        TestGreetingSkillCode,
			Name:        "测试问候Skill",
			Description: "用于测试客服 Agent 的 skill 路由。适合处理问候、机器人自我介绍、询问你能做什么、明确提到测试 skill 的问题。",
			Instruction: "你是一个用于本地测试的客服 Skill。当用户在打招呼、询问你是谁、问你能做什么，或者明确提到测试 skill 时，请直接用中文回复：这是测试 Skill 的固定回复，说明 Agent 已经成功路由到 Skill 能力。",
			Priority:    100,
			Status:      enums.StatusOk,
			Remark:      "Local testdata seed",
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
