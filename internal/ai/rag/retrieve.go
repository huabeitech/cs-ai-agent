package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mlogclub/simple/sqls"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
)

type retrieve struct {
}

var Retrieve = &retrieve{}

func (s *retrieve) Retrieve(ctx context.Context, req RetrieveRequest) ([]RetrieveResult, error) {
	if req.Query == "" {
		return nil, nil
	}
	knowledgeBaseIDs := normalizeKnowledgeBaseIDs(req.KnowledgeBaseIDs)
	if len(knowledgeBaseIDs) == 0 {
		return nil, nil
	}

	retrievableKnowledgeBaseIDs := s.filterRetrievableKnowledgeBaseIDs(knowledgeBaseIDs)
	if len(retrievableKnowledgeBaseIDs) == 0 {
		slog.Info("Skip retrieve for non-enabled knowledge bases",
			"knowledge_base_ids", fmt.Sprint(knowledgeBaseIDs))
		return nil, nil
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 8
	}

	scoreThreshold := float32(req.ScoreThreshold)
	if scoreThreshold <= 0 {
		scoreThreshold = 0.3
	}

	embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	collectionName := knowledgeCollectionName
	provider := vectordb.GetProvider()
	if provider == nil {
		return nil, fmt.Errorf("vectordb provider not initialized")
	}

	searchResults, err := provider.Search(ctx, &vectordb.SearchRequest{
		CollectionName: collectionName,
		Vector:         embeddingResult.Vector,
		TopK:           topK,
		ScoreThreshold: scoreThreshold,
		Filter: &vectordb.SearchFilter{
			KnowledgeBaseIDs: retrievableKnowledgeBaseIDs,
		},
	})
	if err != nil {
		slog.Error("Failed to search vectors", "error", err)
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	if len(searchResults) == 0 && scoreThreshold > 0 {
		s.logEmptySearchDiagnostics(ctx, provider, collectionName, embeddingResult.Vector, topK, scoreThreshold, retrievableKnowledgeBaseIDs, req)
	}

	if len(searchResults) == 0 {
		return nil, nil
	}

	results := make([]RetrieveResult, 0, len(searchResults))
	for _, sr := range searchResults {
		chunk := repositories.KnowledgeChunkRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("vector_id", sr.ID))
		if chunk == nil || chunk.Status != enums.StatusOk {
			continue
		}

		document := repositories.KnowledgeDocumentRepository.Get(sqls.DB(), chunk.DocumentID)
		if document == nil || document.Status != enums.StatusOk {
			continue
		}
		documentTitle := document.Title

		results = append(results, RetrieveResult{
			KnowledgeBaseID: chunk.KnowledgeBaseID,
			ChunkID:         chunk.ID,
			DocumentID:      chunk.DocumentID,
			DocumentTitle:   documentTitle,
			ChunkNo:         chunk.ChunkNo,
			Title:           chunk.Title,
			SectionPath:     sr.Payload.SectionPath,
			Content:         chunk.Content,
			Score:           sr.Score,
			ChunkType:       extractChunkType(sr.Payload),
		})
	}

	return results, nil
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

func (s *retrieve) filterRetrievableKnowledgeBaseIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	items := repositories.KnowledgeBaseRepository.Find(sqls.DB(), sqls.NewCnd().In("id", ids))
	if len(items) == 0 {
		return nil
	}
	allowed := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item.Status == enums.StatusOk {
			allowed[item.ID] = struct{}{}
		}
	}
	filtered := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
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
