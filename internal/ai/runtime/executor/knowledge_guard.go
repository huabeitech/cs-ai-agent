package executor

import (
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	"github.com/cloudwego/eino/schema"
)

type knowledgeGuardDecision struct {
	FallbackReply string
	Instructions  []*schema.Message
}

func buildKnowledgeGuardDecision(aiAgent *models.AIAgent, retrieveResult *retrievers.KnowledgeRetrieveResult) knowledgeGuardDecision {
	if aiAgent == nil || retrieveResult == nil || len(retrieveResult.KnowledgeBaseIDs) == 0 {
		return knowledgeGuardDecision{}
	}
	fallbackReply := resolveKnowledgeFallbackReply(aiAgent, retrieveResult.FallbackMode)
	if len(retrieveResult.Hits) == 0 {
		return knowledgeGuardDecision{FallbackReply: fallbackReply}
	}
	instruction := buildKnowledgeRuntimeInstruction(retrieveResult.AnswerMode, fallbackReply)
	if instruction == "" {
		return knowledgeGuardDecision{}
	}
	return knowledgeGuardDecision{
		Instructions: []*schema.Message{schema.SystemMessage(instruction)},
	}
}

func resolveKnowledgeFallbackReply(aiAgent *models.AIAgent, fallbackMode enums.KnowledgeFallbackMode) string {
	if aiAgent != nil {
		if reply := strings.TrimSpace(aiAgent.FallbackMessage); reply != "" {
			return reply
		}
	}
	switch fallbackMode {
	case enums.KnowledgeFallbackModeSuggestRetry:
		return "当前知识库里没有找到足够明确的信息，你可以换个更具体的问法再试一次。"
	case enums.KnowledgeFallbackModeTransferHuman:
		return "当前知识库里没有找到足够明确的信息，建议转人工进一步处理。"
	default:
		return "当前知识库暂无明确信息。"
	}
}

func buildKnowledgeRuntimeInstruction(answerMode enums.KnowledgeAnswerMode, fallbackReply string) string {
	fallbackReply = strings.TrimSpace(fallbackReply)
	if fallbackReply == "" {
		fallbackReply = "当前知识库暂无明确信息。"
	}
	if answerMode == enums.KnowledgeAnswerModeAssist {
		return "知识库回答约束：优先依据后续提供的知识片段回答，可以做轻度归纳，但不要编造片段中未提供的事实。若知识片段不足以直接支持答案，必须明确回复：" + fallbackReply
	}
	return "知识库回答约束：本轮只能依据后续提供的知识片段回答，不得使用模型常识补充未提供的事实。若知识片段不足以支持回答，必须明确回复：" + fallbackReply
}
