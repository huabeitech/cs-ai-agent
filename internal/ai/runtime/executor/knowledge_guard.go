package executor

import (
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	"github.com/cloudwego/eino/schema"
	"github.com/mlogclub/simple/common/strs"
)

type knowledgeGuardDecision struct {
	FallbackReply string
	Instructions  []*schema.Message
}

func buildKnowledgeUnavailableDecision(aiAgent models.AIAgent, knowledgeBaseIDs []int64) knowledgeGuardDecision {
	if len(knowledgeBaseIDs) == 0 {
		return knowledgeGuardDecision{}
	}
	return knowledgeGuardDecision{FallbackReply: resolveKnowledgeFallbackReply(aiAgent)}
}

func buildKnowledgeGuardDecision(aiAgent models.AIAgent, retrieveResult *retrievers.KnowledgeRetrieveResult) knowledgeGuardDecision {
	if retrieveResult == nil || len(retrieveResult.KnowledgeBaseIDs) == 0 {
		return knowledgeGuardDecision{}
	}
	fallbackReply := resolveKnowledgeFallbackReply(aiAgent)
	if len(retrieveResult.Hits) == 0 || strings.TrimSpace(retrieveResult.ContextText) == "" {
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

func resolveKnowledgeFallbackReply(aiAgent models.AIAgent) string {
	if reply := strings.TrimSpace(aiAgent.FallbackMessage); reply != "" {
		return reply
	}
	switch aiAgent.FallbackMode {
	case enums.AIAgentFallbackModeSuggestRetry:
		return "当前知识库里没有找到足够明确的信息，你可以换个更具体的问法再试一次。"
	default:
		return "当前知识库暂无明确信息。"
	}
}

func resolveKnowledgeHumanSupportFallback(aiAgent models.AIAgent) string {
	base := strings.TrimSpace(resolveKnowledgeFallbackReply(aiAgent))
	if strs.IsBlank(base) {
		base = "当前知识库暂无明确信息。"
	}
	return base
}

func buildKnowledgeRuntimeInstruction(answerMode enums.KnowledgeAnswerMode, fallbackReply string) string {
	fallbackReply = strings.TrimSpace(fallbackReply)
	if fallbackReply == "" {
		fallbackReply = "当前知识库暂无明确信息。"
	}
	if answerMode == enums.KnowledgeAnswerModeAssist {
		return "知识库回答约束：优先依据后续提供的知识片段回答，可以做轻度归纳，但不要编造片段中未提供的事实。回答中的具体事实、步骤、承诺必须能被知识片段直接支持；若知识片段不足以直接支持答案，必须明确回复：" + fallbackReply
	}
	return "知识库回答约束：本轮只能依据后续提供的知识片段回答，不得使用模型常识补充未提供的事实，不得输出知识片段外的具体事实、步骤、承诺或建议。若知识片段不足以支持回答，必须明确回复：" + fallbackReply
}
