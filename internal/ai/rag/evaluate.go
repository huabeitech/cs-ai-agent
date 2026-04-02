package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cs-agent/internal/ai"
	ragchunk "cs-agent/internal/ai/rag/chunk"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/google/uuid"
	"github.com/mlogclub/simple/sqls"
)

type evaluate struct {
	registry *ragchunk.Registry
}

var Evaluate = &evaluate{
	registry: ragchunk.NewDefaultRegistry(),
}

func (s *evaluate) DebugCompareProviders(ctx context.Context, req request.KnowledgeCompareRequest) (*response.KnowledgeCompareResponse, error) {
	if strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("问题不能为空")
	}

	knowledgeBase := repositories.KnowledgeBaseRepository.Get(sqls.DB(), req.KnowledgeBaseID)
	if knowledgeBase == nil {
		return nil, fmt.Errorf("知识库不存在")
	}

	queryEmbedding, err := ai.Embedding.GenerateEmbedding(ctx, req.Question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	providers := normalizeCompareProviders(req.Providers)
	items := make([]response.KnowledgeCompareProviderResult, 0, len(providers))
	startedAt := time.Now()
	for _, providerName := range providers {
		item, err := s.compareProvider(ctx, knowledgeBase, providerName, req, queryEmbedding.Vector, queryEmbedding.Dimension)
		if err != nil {
			slog.Warn("Knowledge provider compare failed", "provider", providerName, "error", err)
			continue
		}
		items = append(items, *item)
	}

	return &response.KnowledgeCompareResponse{
		Question:  req.Question,
		Providers: items,
		LatencyMs: time.Since(startedAt).Milliseconds(),
	}, nil
}

func (s *evaluate) compareProvider(ctx context.Context, knowledgeBase *models.KnowledgeBase, providerName string, req request.KnowledgeCompareRequest, queryVector []float32, dimension int) (*response.KnowledgeCompareProviderResult, error) {
	if s.registry.Get(providerName) == nil {
		return nil, fmt.Errorf("chunk provider not found: %s", providerName)
	}

	vectorProvider := vectordb.GetProvider()
	if vectorProvider == nil {
		return nil, fmt.Errorf("vectordb provider not initialized")
	}
	if dimension <= 0 {
		return nil, fmt.Errorf("invalid embedding dimension")
	}

	buildStartedAt := time.Now()
	collectionName := buildKnowledgeEvaluateCollectionName(knowledgeBase.ID, providerName)
	if err := vectorProvider.CreateCollection(ctx, collectionName, dimension); err != nil {
		return nil, fmt.Errorf("failed to create compare collection: %w", err)
	}
	defer func() {
		if err := vectorProvider.DeleteCollection(ctx, collectionName); err != nil {
			slog.Warn("Failed to delete compare collection", "collection", collectionName, "error", err)
		}
	}()

	docs := repositories.KnowledgeDocumentRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("knowledge_base_id", knowledgeBase.ID).
		Eq("status", enums.StatusOk).
		Asc("id"))
	if len(docs) == 0 {
		return &response.KnowledgeCompareProviderResult{
			Provider:           providerName,
			HitCount:           0,
			BuildMs:            time.Since(buildStartedAt).Milliseconds(),
			RetrieveMs:         0,
			Top1Matched:        false,
			Top3Matched:        false,
			MatchedDocumentIDs: nil,
			Results:            nil,
		}, nil
	}

	vectors := make([]vectordb.Vector, 0)
	for _, doc := range docs {
		chunks, err := s.registry.Chunk(ctx, &ragchunk.ChunkRequest{
			KnowledgeBaseID: knowledgeBase.ID,
			DocumentID:      doc.ID,
			DocumentTitle:   doc.Title,
			ContentType:     doc.ContentType,
			Content:         doc.Content,
			PlainText:       ExtractPlainText(doc.Content, doc.ContentType),
			Options: ragchunk.ChunkOptions{
				Provider:       providerName,
				TargetTokens:   knowledgeBase.ChunkTargetTokens,
				MaxTokens:      knowledgeBase.ChunkMaxTokens,
				OverlapTokens:  knowledgeBase.ChunkOverlapTokens,
				EnableFallback: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to chunk document %d: %w", doc.ID, err)
		}
		for _, item := range chunks {
			embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, item.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to generate embedding for compare chunk: %w", err)
			}
			vectors = append(vectors, vectordb.Vector{
				ID:     uuid.NewString(),
				Vector: embeddingResult.Vector,
				Payload: vectordb.ChunkPayload{
					KnowledgeBaseID: knowledgeBase.ID,
					DocumentID:      doc.ID,
					DocumentTitle:   doc.Title,
					ChunkNo:         item.ChunkNo,
					ChunkType:       string(item.ChunkType),
					SectionPath:     item.SectionPath,
					Title:           item.Title,
					Content:         item.Content,
					Provider:        providerName,
				},
			})
		}
	}

	if len(vectors) > 0 {
		if err := vectorProvider.UpsertVectors(ctx, collectionName, vectors); err != nil {
			return nil, fmt.Errorf("failed to upsert compare vectors: %w", err)
		}
	}
	buildMs := time.Since(buildStartedAt).Milliseconds()

	retrieveStartedAt := time.Now()
	topK := req.TopK
	if topK <= 0 {
		topK = knowledgeBase.DefaultTopK
	}
	threshold := float32(req.ScoreThreshold)
	if threshold <= 0 {
		threshold = float32(knowledgeBase.DefaultScoreThreshold)
	}
	searchResults, err := vectorProvider.Search(ctx, &vectordb.SearchRequest{
		CollectionName: collectionName,
		Vector:         queryVector,
		TopK:           topK,
		ScoreThreshold: threshold,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search compare collection: %w", err)
	}
	retrieveMs := time.Since(retrieveStartedAt).Milliseconds()

	results := make([]response.KnowledgeSearchResult, 0, len(searchResults))
	for _, item := range searchResults {
		results = append(results, response.KnowledgeSearchResult{
			KnowledgeBaseID: item.Payload.KnowledgeBaseID,
			ChunkID:         0,
			DocumentID:      item.Payload.DocumentID,
			DocumentTitle:   item.Payload.DocumentTitle,
			FaqID:           item.Payload.FaqID,
			FaqQuestion:     item.Payload.FaqQuestion,
			ChunkNo:         item.Payload.ChunkNo,
			Title:           item.Payload.Title,
			SectionPath:     item.Payload.SectionPath,
			Content:         item.Payload.Content,
			Score:           float64(item.Score),
		})
	}
	top1Matched, top3Matched, matchedDocumentIDs := evaluateExpectedDocumentHits(results, req.ExpectedDocIDs)

	return &response.KnowledgeCompareProviderResult{
		Provider:           providerName,
		HitCount:           len(results),
		BuildMs:            buildMs,
		RetrieveMs:         retrieveMs,
		Top1Matched:        top1Matched,
		Top3Matched:        top3Matched,
		MatchedDocumentIDs: matchedDocumentIDs,
		Results:            results,
	}, nil
}

func normalizeCompareProviders(providers []string) []string {
	if len(providers) == 0 {
		return []string{
			string(enums.KnowledgeChunkProviderFixed),
			string(enums.KnowledgeChunkProviderStructured),
			string(enums.KnowledgeChunkProviderFAQ),
		}
	}
	result := make([]string, 0, len(providers))
	seen := make(map[string]struct{})
	for _, provider := range providers {
		provider = strings.TrimSpace(provider)
		if provider == "" {
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		result = append(result, provider)
	}
	return result
}

func buildKnowledgeEvaluateCollectionName(knowledgeBaseID int64, provider string) string {
	return fmt.Sprintf("kb_eval_%d_%s_%s", knowledgeBaseID, provider, uuid.NewString())
}

func evaluateExpectedDocumentHits(results []response.KnowledgeSearchResult, expectedDocIDs []int64) (bool, bool, []int64) {
	if len(expectedDocIDs) == 0 || len(results) == 0 {
		return false, false, nil
	}
	expected := make(map[int64]struct{}, len(expectedDocIDs))
	for _, id := range expectedDocIDs {
		if id > 0 {
			expected[id] = struct{}{}
		}
	}

	matchedSet := make(map[int64]struct{})
	for _, item := range results {
		if _, ok := expected[item.DocumentID]; ok {
			matchedSet[item.DocumentID] = struct{}{}
		}
	}
	matchedDocumentIDs := make([]int64, 0, len(matchedSet))
	for id := range matchedSet {
		matchedDocumentIDs = append(matchedDocumentIDs, id)
	}

	top1Matched := false
	if len(results) > 0 {
		_, top1Matched = expected[results[0].DocumentID]
	}
	top3Matched := false
	limit := 3
	if len(results) < limit {
		limit = len(results)
	}
	for i := 0; i < limit; i++ {
		if _, ok := expected[results[i].DocumentID]; ok {
			top3Matched = true
			break
		}
	}
	return top1Matched, top3Matched, matchedDocumentIDs
}
