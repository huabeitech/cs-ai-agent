package agent

import (
	"context"
	"fmt"
	"strings"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"

	"github.com/mlogclub/simple/common/strs"
)

type ragExecutor struct{}

type knowledgeCandidate struct {
	knowledgeBase *models.KnowledgeBase
	results       []rag.RetrieveResult
	topScore      float32
}

func newRAGExecutor() *ragExecutor {
	return &ragExecutor{}
}

func (e *ragExecutor) Execute(ctx context.Context, turnCtx TurnContext, question string) (*TurnResult, error) {
	candidate, err := e.pickKnowledgeCandidate(ctx, turnCtx, question)
	if err != nil {
		return nil, err
	}
	if candidate == nil || len(candidate.results) == 0 {
		return &TurnResult{
			Action:        ActionFallback,
			Question:      question,
			Reason:        "知识库未命中",
			PlannedAction: ActionRAG,
			PlanReason:    "rag",
		}, nil
	}
	replyText, err := e.generateReply(ctx, turnCtx, candidate, question)
	if err != nil {
		return &TurnResult{
			Action:        ActionFallback,
			Question:      question,
			Reason:        "AI生成失败",
			PlannedAction: ActionRAG,
			PlanReason:    "rag",
			KnowledgeBase: candidate.knowledgeBase,
			RetrieveHits:  candidate.results,
		}, err
	}
	if strings.TrimSpace(replyText) == "" {
		return &TurnResult{
			Action:        ActionFallback,
			Question:      question,
			Reason:        "AI回复为空",
			PlannedAction: ActionRAG,
			PlanReason:    "rag",
			KnowledgeBase: candidate.knowledgeBase,
			RetrieveHits:  candidate.results,
		}, nil
	}
	return &TurnResult{
		Action:        ActionReply,
		Question:      question,
		ReplyText:     replyText,
		PlannedAction: ActionRAG,
		PlanReason:    "rag",
		KnowledgeBase: candidate.knowledgeBase,
		RetrieveHits:  candidate.results,
	}, nil
}

func (e *ragExecutor) pickKnowledgeCandidate(ctx context.Context, turnCtx TurnContext, question string) (*knowledgeCandidate, error) {
	aiAgent := turnCtx.AIAgent
	if aiAgent == nil {
		return nil, nil
	}
	knowledgeIDs := utils.SplitInt64s(aiAgent.KnowledgeIDs)
	if len(knowledgeIDs) == 0 {
		return nil, nil
	}
	results, err := rag.Retrieve.Retrieve(ctx, rag.RetrieveRequest{
		KnowledgeBaseIDs: knowledgeIDs,
		Query:            question,
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	var knowledgeBase *models.KnowledgeBase
	if results[0].KnowledgeBaseID > 0 {
		knowledgeBase = findKnowledgeBase(turnCtx, results[0].KnowledgeBaseID)
	}

	return &knowledgeCandidate{
		knowledgeBase: knowledgeBase,
		results:       results,
		topScore:      results[0].Score,
	}, nil
}

func (e *ragExecutor) generateReply(ctx context.Context, turnCtx TurnContext, candidate *knowledgeCandidate, question string) (string, error) {
	if turnCtx.AIConfig == nil || turnCtx.AIAgent == nil || candidate == nil || candidate.knowledgeBase == nil {
		return "", nil
	}
	contextText := rag.Retrieve.BuildContext(ctx, candidate.results, 4000)
	if strings.TrimSpace(contextText) == "" {
		return "", nil
	}

	systemPrompt := e.buildSystemPrompt(turnCtx.AIAgent, candidate.knowledgeBase)
	userPrompt := fmt.Sprintf("用户问题：%s\n\n参考资料：\n%s", question, contextText)
	result, err := ai.LLM.ChatWithConfig(ctx, turnCtx.AIConfig, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Content), nil
}

func (e *ragExecutor) buildSystemPrompt(aiAgent *models.AIAgent, knowledgeBase *models.KnowledgeBase) string {
	if aiAgent != nil && strs.IsNotBlank(aiAgent.SystemPrompt) {
		return strings.TrimSpace(aiAgent.SystemPrompt)
	}

	basePrompt := "你是严格的客服知识库助手。只能依据提供的知识片段回答；如果资料不足，请明确说明知识库暂无明确信息。"
	if knowledgeBase != nil && enums.KnowledgeAnswerMode(knowledgeBase.AnswerMode) == enums.KnowledgeAnswerModeAssist {
		basePrompt = "你是客服知识库助手。请优先依据提供的知识片段回答，可以做轻度归纳，但不要编造未提供的事实。"
	}
	return basePrompt
}
