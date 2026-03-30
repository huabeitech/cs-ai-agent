package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"

	"cs-agent/internal/ai"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
)

type answer struct {
}

var Answer = &answer{}

func (s *answer) DebugSearch(ctx context.Context, req request.KnowledgeSearchRequest) (*response.KnowledgeSearchResponse, error) {
	if strings.TrimSpace(req.Question) == "" {
		return nil, errorsx.InvalidParam("问题不能为空")
	}
	startedAt := time.Now()
	results, err := s.retrieve(req, ctx)
	if err != nil {
		return nil, err
	}

	respResults := make([]response.KnowledgeSearchResult, 0, len(results))
	for _, item := range results {
		respResults = append(respResults, response.KnowledgeSearchResult{
			KnowledgeBaseID: item.KnowledgeBaseID,
			ChunkID:         item.ChunkID,
			DocumentID:      item.DocumentID,
			DocumentTitle:   item.DocumentTitle,
			ChunkNo:         item.ChunkNo,
			Title:           item.Title,
			SectionPath:     item.SectionPath,
			Content:         item.Content,
			Score:           float64(item.Score),
		})
	}

	return &response.KnowledgeSearchResponse{
		Question:  req.Question,
		Results:   respResults,
		HitCount:  len(respResults),
		LatencyMs: time.Since(startedAt).Milliseconds(),
	}, nil
}

func (s *answer) DebugAnswer(ctx context.Context, req request.KnowledgeAnswerRequest, operator *dto.AuthPrincipal) (*response.KnowledgeAnswerResponse, error) {
	if strings.TrimSpace(req.Question) == "" {
		return nil, errorsx.InvalidParam("问题不能为空")
	}
	startedAt := time.Now()

	retrieveStartedAt := time.Now()
	knowledgeBase := s.resolvePrimaryKnowledgeBase(req.KnowledgeBaseIDs)
	results, err := s.retrieve(request.KnowledgeSearchRequest{
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		Question:         req.Question,
		TopK:             req.TopK,
		ScoreThreshold:   req.ScoreThreshold,
		RerankLimit:      req.RerankLimit,
	}, ctx)
	if err != nil {
		return nil, err
	}
	retrieveMs := time.Since(retrieveStartedAt).Milliseconds()
	contextResults := buildContextHits(Retrieve.SelectContextResults(results, 4000))

	hits := make([]response.KnowledgeSearchResult, 0, len(results))
	topScore := 0.0
	for i, item := range results {
		score := float64(item.Score)
		if i == 0 {
			topScore = score
		}
		hits = append(hits, response.KnowledgeSearchResult{
			KnowledgeBaseID: item.KnowledgeBaseID,
			ChunkID:         item.ChunkID,
			DocumentID:      item.DocumentID,
			DocumentTitle:   item.DocumentTitle,
			ChunkNo:         item.ChunkNo,
			Title:           item.Title,
			SectionPath:     item.SectionPath,
			Content:         item.Content,
			Score:           score,
		})
	}
	citations := buildKnowledgeCitations(hits, 3)

	answerMode := enums.KnowledgeAnswerMode(req.AnswerMode)
	if answerMode == 0 {
		if knowledgeBase != nil {
			answerMode = enums.KnowledgeAnswerMode(knowledgeBase.AnswerMode)
		}
		if answerMode == 0 {
			answerMode = enums.KnowledgeAnswerModeStrict
		}
	}

	fallbackMode := enums.KnowledgeFallbackMode(req.FallbackMode)
	if fallbackMode == 0 {
		if knowledgeBase != nil {
			fallbackMode = enums.KnowledgeFallbackMode(knowledgeBase.FallbackMode)
		}
		if fallbackMode == 0 {
			fallbackMode = enums.KnowledgeFallbackModeNoAnswer
		}
	}

	answerStatus := enums.KnowledgeAnswerStatusNormal
	answer := ""
	modelName := ""
	promptTokens := 0
	completionTokens := 0
	generateStartedAt := time.Now()

	if len(hits) == 0 {
		answerStatus = enums.KnowledgeAnswerStatusNoAnswer
		answer = buildFallbackAnswer(fallbackMode)
	} else {
		contextText := Retrieve.BuildContext(ctx, results, 4000)
		systemPrompt := buildAnswerSystemPrompt(answerMode)
		userPrompt := fmt.Sprintf("用户问题：%s\n\n参考资料：\n%s", req.Question, contextText)
		llmResult, llmErr := ai.LLM.Chat(ctx, systemPrompt, userPrompt)
		if llmErr != nil {
			answerStatus = enums.KnowledgeAnswerStatusFallback
			answer = buildFallbackAnswer(fallbackMode)
		} else {
			answer = llmResult.Content
			modelName = llmResult.ModelName
			promptTokens = llmResult.PromptTokens
			completionTokens = llmResult.CompletionTokens
			if strings.TrimSpace(answer) == "" {
				answerStatus = enums.KnowledgeAnswerStatusFallback
				answer = buildFallbackAnswer(fallbackMode)
			}
		}
	}
	generateMs := time.Since(generateStartedAt).Milliseconds()
	rerankLimit := 0
	chunkProvider := ""
	chunkTargetTokens := 0
	chunkMaxTokens := 0
	chunkOverlapTokens := 0
	if knowledgeBase != nil {
		rerankLimit = resolveRerankLimit(req.RerankLimit, knowledgeBase.DefaultRerankLimit)
		chunkProvider = knowledgeBase.ChunkProvider
		chunkTargetTokens = knowledgeBase.ChunkTargetTokens
		chunkMaxTokens = knowledgeBase.ChunkMaxTokens
		chunkOverlapTokens = knowledgeBase.ChunkOverlapTokens
	}

	logItem, err := RetrieveLog.CreateRetrieveLog(&CreateRetrieveLogRequest{
		KnowledgeBaseID:    firstKnowledgeBaseID(req.KnowledgeBaseIDs),
		Channel:            defaultRetrieveChannel(req.Channel),
		Scene:              defaultRetrieveScene(req.Scene),
		SessionID:          req.SessionID,
		ConversationID:     req.ConversationID,
		Question:           req.Question,
		RewriteQuestion:    "",
		Answer:             answer,
		AnswerStatus:       int(answerStatus),
		ChunkProvider:      chunkProvider,
		ChunkTargetTokens:  chunkTargetTokens,
		ChunkMaxTokens:     chunkMaxTokens,
		ChunkOverlapTokens: chunkOverlapTokens,
		RerankEnabled:      rerankLimit > 0,
		RerankLimit:        rerankLimit,
		Hits:               hits,
		UsedHits:           contextResults,
		Citations:          citations,
		LatencyMs:          time.Since(startedAt).Milliseconds(),
		RetrieveMs:         retrieveMs,
		GenerateMs:         generateMs,
		PromptTokens:       promptTokens,
		CompletionTokens:   completionTokens,
		ModelName:          modelName,
	}, operator)
	if err != nil {
		return nil, err
	}

	return &response.KnowledgeAnswerResponse{
		Question:         req.Question,
		Answer:           answer,
		AnswerStatus:     int(answerStatus),
		AnswerStatusName: getAnswerStatusName(answerStatus),
		Citations:        citations,
		Hits:             hits,
		HitCount:         len(hits),
		TopScore:         topScore,
		LatencyMs:        time.Since(startedAt).Milliseconds(),
		RetrieveMs:       retrieveMs,
		GenerateMs:       generateMs,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		ModelName:        modelName,
		RetrieveLogID:    logItem.ID,
	}, nil
}

func buildContextHits(results []RetrieveResult) []response.KnowledgeSearchResult {
	if len(results) == 0 {
		return nil
	}
	hits := make([]response.KnowledgeSearchResult, 0, len(results))
	for _, item := range results {
		hits = append(hits, response.KnowledgeSearchResult{
			KnowledgeBaseID: item.KnowledgeBaseID,
			ChunkID:         item.ChunkID,
			DocumentID:      item.DocumentID,
			DocumentTitle:   item.DocumentTitle,
			ChunkNo:         item.ChunkNo,
			Title:           item.Title,
			SectionPath:     item.SectionPath,
			Content:         item.Content,
			Score:           float64(item.Score),
		})
	}
	return hits
}

func buildKnowledgeCitations(hits []response.KnowledgeSearchResult, limit int) []response.KnowledgeCitation {
	if len(hits) == 0 || limit <= 0 {
		return nil
	}
	citations := make([]response.KnowledgeCitation, 0, limit)
	seen := make(map[string]struct{})
	for _, item := range hits {
		key := fmt.Sprintf("%d|%s|%d", item.DocumentID, item.SectionPath, item.ChunkNo)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		citations = append(citations, response.KnowledgeCitation{
			DocumentID:    item.DocumentID,
			DocumentTitle: item.DocumentTitle,
			ChunkNo:       item.ChunkNo,
			Title:         item.Title,
			SectionPath:   item.SectionPath,
			Snippet:       truncateCitationSnippet(item.Content, 120),
			Score:         item.Score,
		})
		if len(citations) >= limit {
			break
		}
	}
	return citations
}

func truncateCitationSnippet(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}

func (s *answer) BuildDocumentIndex(ctx context.Context, documentID int64) error {
	return Index.IndexDocumentByID(ctx, documentID)
}

func (s *answer) retrieve(req request.KnowledgeSearchRequest, ctx context.Context) ([]RetrieveResult, error) {
	if len(normalizeKnowledgeBaseIDs(req.KnowledgeBaseIDs)) == 0 {
		return nil, errorsx.InvalidParam("知识库不能为空")
	}
	knowledgeBase := s.resolvePrimaryKnowledgeBase(req.KnowledgeBaseIDs)

	results, err := Retrieve.Retrieve(ctx, RetrieveRequest{
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		Query:            req.Question,
		TopK:             req.TopK,
		ScoreThreshold:   req.ScoreThreshold,
	})
	if err != nil {
		return nil, err
	}

	defaultRerankLimit := 0
	if knowledgeBase != nil {
		defaultRerankLimit = knowledgeBase.DefaultRerankLimit
	}
	rerankLimit := resolveRerankLimit(req.RerankLimit, defaultRerankLimit)
	if rerankLimit > 0 && len(results) > rerankLimit {
		return Retrieve.RetrieveWithRerank(ctx, RetrieveRequest{
			KnowledgeBaseIDs: req.KnowledgeBaseIDs,
			Query:            req.Question,
			TopK:             req.TopK,
			ScoreThreshold:   req.ScoreThreshold,
		}, rerankLimit)
	}
	return results, nil
}

func (s *answer) resolvePrimaryKnowledgeBase(knowledgeBaseIDs []int64) *models.KnowledgeBase {
	normalized := normalizeKnowledgeBaseIDs(knowledgeBaseIDs)
	for _, id := range normalized {
		item := repositories.KnowledgeBaseRepository.Get(sqls.DB(), id)
		if item != nil {
			return item
		}
	}
	return nil
}

func firstKnowledgeBaseID(ids []int64) int64 {
	normalized := normalizeKnowledgeBaseIDs(ids)
	if len(normalized) == 0 {
		return 0
	}
	return normalized[0]
}

func resolveRerankLimit(requestLimit int, defaultLimit int) int {
	if requestLimit > 0 {
		return requestLimit
	}
	if defaultLimit > 0 {
		return defaultLimit
	}
	return 0
}

func buildAnswerSystemPrompt(answerMode enums.KnowledgeAnswerMode) string {
	if answerMode == enums.KnowledgeAnswerModeAssist {
		return "你是客服知识库助手。请优先依据提供的知识片段回答，可以做轻度归纳，但不要编造未提供的事实。"
	}
	return "你是严格的客服知识库助手。只能依据提供的知识片段回答；如果资料不足，请明确说明知识库暂无明确信息。"
}

func buildFallbackAnswer(fallbackMode enums.KnowledgeFallbackMode) string {
	switch fallbackMode {
	case enums.KnowledgeFallbackModeSuggestRetry:
		return "当前知识库里没有找到足够明确的信息，你可以换个更具体的问法再试一次。"
	case enums.KnowledgeFallbackModeTransferHuman:
		return "当前知识库里没有找到足够明确的信息，建议转人工进一步处理。"
	default:
		return "当前知识库暂无明确信息。"
	}
}

func defaultRetrieveChannel(channel string) string {
	if strings.TrimSpace(channel) == "" {
		return string(enums.KnowledgeRetrieveChannelDebug)
	}
	return channel
}

func defaultRetrieveScene(scene string) string {
	if strings.TrimSpace(scene) == "" {
		return string(enums.KnowledgeRetrieveSceneQA)
	}
	return scene
}

func getAnswerStatusName(status enums.KnowledgeAnswerStatus) string {
	switch status {
	case enums.KnowledgeAnswerStatusNormal:
		return "正常"
	case enums.KnowledgeAnswerStatusNoAnswer:
		return "无答案"
	case enums.KnowledgeAnswerStatusFallback:
		return "兜底"
	case enums.KnowledgeAnswerStatusBlocked:
		return "风控拦截"
	default:
		return "未知"
	}
}
