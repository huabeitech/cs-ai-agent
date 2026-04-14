package skill

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"time"

	"github.com/mlogclub/simple/sqls"
)

const AfterSalesEscalationSkillCode = "after_sales_escalation_skill"

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
			Code:        AfterSalesEscalationSkillCode,
			Name:        "售后升级处理",
			Description: "处理报障、投诉、售后跟进、建单、转人工等升级诉求。只在用户明确需要售后介入或问题升级处理时命中，不处理普通问候、产品介绍或泛咨询。",
			Instruction: `你是“售后升级处理”专项 Skill，负责承接需要升级处理的客服诉求。

你的职责边界：
1. 仅处理以下场景：报障、投诉、问题久未解决、明确要求建单、明确要求转人工、要求售后继续跟进。
2. 如果用户只是普通咨询、产品使用提问、寒暄、问候、闲聊，说明当前不属于本 Skill 的职责，避免误判为升级处理。

你的处理规则：
1. 先判断用户是否已经明确表达升级诉求；如果还不明确，先用简洁中文追问关键事实，例如订单号、设备/产品名称、故障现象、已尝试过的操作、期望处理方式。
2. 如果用户要投诉、报障或售后跟进，但信息比较散乱，优先调用 graph/prepare_ticket_draft 整理工单草稿，再根据缺失字段继续追问。
3. 只有在用户明确希望提交工单、投诉单、报障单，且标题与问题描述已经足够清晰时，才调用 graph/create_ticket_with_confirmation。
4. 只有在用户明确要求人工客服，或你已经判断必须人工继续处理时，才调用 graph/handoff_to_human。
5. 禁止只在文本里声称“已经建单”或“已经转人工”，相关动作必须通过对应工具执行。
6. 如果信息不足以建单或转人工，先澄清，不要直接升级动作。

回复要求：
1. 全程使用中文，语气专业、简洁、像真实客服。
2. 优先围绕问题定位和升级处理推进，不要输出与当前诉求无关的自我介绍。
3. 如果进入确认流程，明确告知用户你将协助提交或转接，并等待确认结果。`,
			Examples: `[
  "设备今天开始一直离线，重启也没用，帮我提个工单",
  "这个问题三天了还没解决，我要投诉一下",
  "麻烦转人工，你这边解决不了",
  "售后什么时候联系我？这个故障还没有人跟进",
  "帮我登记一下报障，产品型号是AX300，无法联网",
  "我要申请售后处理，这个问题反复出现",
  "你好，你是谁",
  "这个产品怎么安装"
]`,
			ToolWhitelist: `[
  "graph/prepare_ticket_draft",
  "graph/create_ticket_with_confirmation",
  "graph/handoff_to_human"
]`,
			Status: enums.StatusOk,
			Remark: "after-sales escalation skill",
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
