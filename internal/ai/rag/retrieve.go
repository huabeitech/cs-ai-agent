package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

type retrieve struct {
}

var Retrieve = &retrieve{}

func (s *retrieve) Retrieve(ctx context.Context, req RetrieveRequest) ([]RetrieveResult, error) {
	results, _, err := s.RetrieveWithTrace(ctx, req)
	return results, err
}

type RetrieveTrace struct {
	EmbeddingMs    int64
	VectorSearchMs int64
	HydrateMs      int64
}

func (s *retrieve) RetrieveWithTrace(ctx context.Context, req RetrieveRequest) ([]RetrieveResult, *RetrieveTrace, error) {
	trace := &RetrieveTrace{}
	if req.Query == "" {
		return nil, trace, nil
	}
	knowledgeBaseIDs := normalizeKnowledgeBaseIDs(req.KnowledgeBaseIDs)
	if len(knowledgeBaseIDs) == 0 {
		return nil, trace, nil
	}

	retrievableKnowledgeBases := s.loadRetrievableKnowledgeBases(knowledgeBaseIDs)
	if len(retrievableKnowledgeBases) == 0 {
		slog.Info("Skip retrieve for non-enabled knowledge bases",
			"knowledge_base_ids", fmt.Sprint(knowledgeBaseIDs))
		return nil, trace, nil
	}

	searchResults, searchTrace, err := s.searchKnowledgeBaseVectors(ctx, req, retrievableKnowledgeBases)
	if err != nil {
		if searchTrace != nil {
			trace.EmbeddingMs = searchTrace.EmbeddingMs
			trace.VectorSearchMs = searchTrace.VectorSearchMs
		}
		return nil, trace, err
	}
	if searchTrace != nil {
		trace.EmbeddingMs = searchTrace.EmbeddingMs
		trace.VectorSearchMs = searchTrace.VectorSearchMs
	}

	if len(searchResults) == 0 {
		return nil, trace, nil
	}
	results, hydrateMs := s.hydrateRetrieveResults(searchResults)
	trace.HydrateMs = hydrateMs

	return results, trace, nil
}

func extractChunkType(payload vectordb.ChunkPayload) string {
	if payload.ChunkType != "" {
		return payload.ChunkType
	}
	return string(enums.KnowledgeChunkTypeText)
}

func (s *retrieve) logEmptySearchDiagnostics(ctx context.Context, provider vectordb.Provider, collectionName string, vector []float32, topK int, scoreThreshold float32, knowledgeBaseIDs []int64, req RetrieveRequest) {
	rawResults, err := provider.Search(ctx, &vectordb.SearchRequest{
		CollectionName: collectionName,
		Vector:         vector,
		TopK:           topK,
		ScoreThreshold: 0,
		Filter: &vectordb.SearchFilter{
			KnowledgeBaseIDs: knowledgeBaseIDs,
		},
	})
	if err != nil {
		slog.Warn("Knowledge retrieve diagnostics failed",
			"knowledge_base_ids", fmt.Sprint(knowledgeBaseIDs),
			"collection", collectionName,
			"query", truncateForLog(req.Query, 80),
			"score_threshold", scoreThreshold,
			"error", err)
		return
	}
	if len(rawResults) == 0 {
		slog.Info("Knowledge retrieve returned no candidates even without threshold",
			"knowledge_base_ids", fmt.Sprint(knowledgeBaseIDs),
			"collection", collectionName,
			"query", truncateForLog(req.Query, 80),
			"score_threshold", scoreThreshold)
		return
	}

	candidates := make([]string, 0, len(rawResults))
	for _, item := range rawResults {
		candidates = append(candidates, fmt.Sprintf("%s:%.4f", item.ID, item.Score))
	}

	slog.Info("Knowledge retrieve filtered all candidates by score threshold",
		"knowledge_base_ids", fmt.Sprint(knowledgeBaseIDs),
		"collection", collectionName,
		"query", truncateForLog(req.Query, 80),
		"score_threshold", scoreThreshold,
		"top_candidates", strings.Join(candidates, ","))
}

func truncateForLog(text string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}

func (s *retrieve) RetrieveWithRerank(ctx context.Context, req RetrieveRequest, rerankLimit int) ([]RetrieveResult, error) {
	results, err := s.Retrieve(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(results) <= rerankLimit {
		return results, nil
	}

	rerankedResults, err := s.rerank(ctx, req.Query, results, rerankLimit)
	if err != nil {
		slog.Warn("Rerank failed, returning original results", "error", err)
		if len(results) > rerankLimit {
			return results[:rerankLimit], nil
		}
		return results, nil
	}

	return rerankedResults, nil
}

func (s *retrieve) rerank(ctx context.Context, query string, results []RetrieveResult, limit int) ([]RetrieveResult, error) {
	return Rerank.RerankResults(ctx, query, results, limit)
}

func (s *retrieve) SelectContextResults(results []RetrieveResult, maxTokens int) []RetrieveResult {
	if len(results) == 0 {
		return nil
	}

	normalizedResults := normalizeContextResults(results)
	selected := make([]RetrieveResult, 0, len(normalizedResults))
	totalTokens := 0
	documentUsage := make(map[int64]int)

	for _, item := range normalizedResults {
		if documentUsage[item.DocumentID] >= 2 {
			continue
		}
		chunkText := buildContextChunkText(item)
		estimatedTokens := len(chunkText) / 2
		if totalTokens+estimatedTokens > maxTokens {
			break
		}
		selected = append(selected, item)
		totalTokens += estimatedTokens
		documentUsage[item.DocumentID]++
	}
	return selected
}

func (s *retrieve) BuildContext(ctx context.Context, results []RetrieveResult, maxTokens int) string {
	if len(results) == 0 {
		return ""
	}

	normalizedResults := s.SelectContextResults(results, maxTokens)
	context := ""
	for _, r := range normalizedResults {
		chunkText := buildContextChunkText(r)
		context += chunkText
	}

	return context
}

func normalizeContextResults(results []RetrieveResult) []RetrieveResult {
	if len(results) == 0 {
		return nil
	}

	merged := mergeAdjacentResults(results)
	return dedupeSectionResults(merged)
}

func dedupeSectionResults(results []RetrieveResult) []RetrieveResult {
	seen := make(map[string]struct{})
	deduped := make([]RetrieveResult, 0, len(results))
	for _, item := range results {
		key := buildSectionKey(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, item)
	}
	return deduped
}

func mergeAdjacentResults(results []RetrieveResult) []RetrieveResult {
	if len(results) == 0 {
		return nil
	}

	merged := make([]RetrieveResult, 0, len(results))
	for _, item := range results {
		if len(merged) == 0 {
			merged = append(merged, item)
			continue
		}

		last := &merged[len(merged)-1]
		if canMergeContextResult(*last, item) {
			last.Content = strings.TrimSpace(last.Content + "\n" + item.Content)
			if item.Score > last.Score {
				last.Score = item.Score
			}
			continue
		}
		merged = append(merged, item)
	}
	return merged
}

func canMergeContextResult(left, right RetrieveResult) bool {
	if left.FaqID > 0 || right.FaqID > 0 {
		return false
	}
	if left.DocumentID != right.DocumentID {
		return false
	}
	if left.SectionPath == "" || right.SectionPath == "" {
		return false
	}
	if left.SectionPath != right.SectionPath {
		return false
	}
	return right.ChunkNo == left.ChunkNo+1
}

func buildSectionKey(item RetrieveResult) string {
	if item.FaqID > 0 {
		return fmt.Sprintf("faq:%d", item.FaqID)
	}
	sectionPath := strings.TrimSpace(item.SectionPath)
	if sectionPath != "" {
		return fmt.Sprintf("%d|%s", item.DocumentID, sectionPath)
	}
	title := strings.TrimSpace(item.Title)
	if title != "" {
		return fmt.Sprintf("%d|%s", item.DocumentID, title)
	}
	return fmt.Sprintf("%d|chunk:%d", item.DocumentID, item.ChunkNo)
}

func buildContextChunkText(item RetrieveResult) string {
	if item.FaqID > 0 {
		title := strings.TrimSpace(item.FaqQuestion)
		if title == "" {
			title = strings.TrimSpace(item.Title)
		}
		if title == "" {
			title = fmt.Sprintf("FAQ#%d", item.FaqID)
		}
		return fmt.Sprintf("【FAQ：%s】\n%s\n\n", title, item.Content)
	}
	title := strings.TrimSpace(item.DocumentTitle)
	if title == "" {
		title = fmt.Sprintf("文档#%d", item.DocumentID)
	}
	if item.SectionPath != "" {
		return fmt.Sprintf("【文档：%s｜章节：%s】\n%s\n\n", title, item.SectionPath, item.Content)
	}
	if item.Title != "" {
		return fmt.Sprintf("【文档：%s｜标题：%s】\n%s\n\n", title, item.Title, item.Content)
	}
	return fmt.Sprintf("【文档：%s】\n%s\n\n", title, item.Content)
}

func (s *retrieve) GetKnowledgeBaseStats(ctx context.Context, knowledgeBaseID int64) (*KnowledgeBaseStats, error) {
	knowledgeBase := repositories.KnowledgeBaseRepository.Get(sqls.DB(), knowledgeBaseID)
	if knowledgeBase == nil {
		return nil, fmt.Errorf("knowledge base not found")
	}

	documentCount := repositories.KnowledgeDocumentRepository.CountByKnowledgeBaseID(sqls.DB(), knowledgeBaseID)
	chunkCount := repositories.KnowledgeChunkRepository.CountByKnowledgeBaseID(sqls.DB(), knowledgeBaseID)

	publishedCount := repositories.KnowledgeDocumentRepository.Count(sqls.DB(), sqls.NewCnd().
		Eq("knowledge_base_id", knowledgeBaseID).
		Eq("status", enums.StatusOk))

	return &KnowledgeBaseStats{
		KnowledgeBaseID: knowledgeBaseID,
		DocumentCount:   documentCount,
		PublishedCount:  publishedCount,
		ChunkCount:      chunkCount,
		VectorCount:     int(chunkCount),
	}, nil
}

func normalizeKnowledgeBaseIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	normalized := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized
}

func resolveKnowledgeBaseSearchOptions(req RetrieveRequest, knowledgeBase *models.KnowledgeBase) (int, float32) {
	topK := req.TopK
	if topK <= 0 && knowledgeBase != nil && knowledgeBase.DefaultTopK > 0 {
		topK = knowledgeBase.DefaultTopK
	}
	if topK <= 0 {
		topK = 8
	}

	scoreThreshold := float32(req.ScoreThreshold)
	if scoreThreshold <= 0 && knowledgeBase != nil && knowledgeBase.DefaultScoreThreshold > 0 {
		scoreThreshold = float32(knowledgeBase.DefaultScoreThreshold)
	}
	if scoreThreshold <= 0 {
		scoreThreshold = 0.3
	}
	return topK, scoreThreshold
}

func (s *retrieve) loadRetrievableKnowledgeBases(ids []int64) []models.KnowledgeBase {
	if len(ids) == 0 {
		return nil
	}
	items := repositories.KnowledgeBaseRepository.Find(sqls.DB(), sqls.NewCnd().In("id", ids))
	if len(items) == 0 {
		return nil
	}
	allowed := make(map[int64]models.KnowledgeBase, len(items))
	for _, item := range items {
		if item.Status == enums.StatusOk {
			allowed[item.ID] = item
		}
	}
	filtered := make([]models.KnowledgeBase, 0, len(ids))
	for _, id := range ids {
		if item, ok := allowed[id]; ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

type KnowledgeBaseStats struct {
	KnowledgeBaseID int64 `json:"knowledgeBaseId"`
	DocumentCount   int64 `json:"documentCount"`
	PublishedCount  int64 `json:"publishedCount"`
	ChunkCount      int64 `json:"chunkCount"`
	VectorCount     int   `json:"vectorCount"`
}
