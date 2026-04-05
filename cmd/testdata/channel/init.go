package channel

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"time"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

type InitResult struct {
	Created int
	Updated int
}

// Init 初始化 Channel 测试数据。
// 依赖于 AI Agent 已初始化。
func Init() (*InitResult, error) {
	result := &InitResult{}

	aiAgentID, err := getDefaultAIAgentID()
	if err != nil {
		return result, fmt.Errorf("get default ai agent id failed: %w", err)
	}
	if aiAgentID == 0 {
		return result, fmt.Errorf("no default ai agent found, please init ai agent first")
	}

	seedItems := buildSeedItems(aiAgentID)
	for _, item := range seedItems {
		itemCopy := item
		if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			existing := repositories.ChannelRepository.Take(ctx.Tx, "name = ? AND channel_type = ?", itemCopy.Name, itemCopy.ChannelType)
			if existing != nil {
				if err := ctx.Tx.Model(existing).Updates(&itemCopy).Error; err != nil {
					return err
				}
				result.Updated++
			} else {
				if err := ctx.Tx.Create(&itemCopy).Error; err != nil {
					return err
				}
				result.Created++
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("upsert channel failed: %w", err)
		}
	}

	return result, nil
}

func buildSeedItems(aiAgentID int64) []models.Channel {
	now := time.Now()
	return []models.Channel{
		{
			Name:        "官网客服",
			ChannelType: enums.ChannelTypeWeb,
			ChannelID:   strs.UUID(),
			AIAgentID:   aiAgentID,
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

func getDefaultAIAgentID() (int64, error) {
	aiAgent := repositories.AIAgentRepository.Take(
		sqls.DB(),
		"status = ?",
		enums.StatusOk,
	)
	if aiAgent == nil {
		return 0, nil
	}
	return aiAgent.ID, nil
}
