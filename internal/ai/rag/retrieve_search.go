package rag

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func (s *retrieve) searchKnowledgeBaseVectors(ctx context.Context, req RetrieveRequest, knowledgeBases []models.KnowledgeBase) ([]vectordb.SearchResult, *RetrieveTrace, error) {
	trace := &RetrieveTrace{}

	embeddingStartedAt := time.Now()
	embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, req.Query)
	trace.EmbeddingMs = time.Since(embeddingStartedAt).Milliseconds()
	if err != nil {
		return nil, trace, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	collectionName := knowledgeCollectionName
	provider := vectordb.GetProvider()
	if provider == nil {
		return nil, trace, fmt.Errorf("vectordb provider not initialized")
	}

	searchResults := make([]vectordb.SearchResult, 0)
	vectorSearchStartedAt := time.Now()
	for _, knowledgeBase := range knowledgeBases {
		topK, scoreThreshold := resolveKnowledgeBaseSearchOptions(req, &knowledgeBase)
		kbResults, searchErr := provider.Search(ctx, &vectordb.SearchRequest{
			CollectionName: collectionName,
			Vector:         embeddingResult.Vector,
			TopK:           topK,
			ScoreThreshold: scoreThreshold,
			Filter: &vectordb.SearchFilter{
				KnowledgeBaseIDs: []int64{knowledgeBase.ID},
			},
		})
		if searchErr != nil {
			slog.Error("Failed to search vectors",
				"knowledge_base_id", knowledgeBase.ID,
				"error", searchErr)
			trace.VectorSearchMs = time.Since(vectorSearchStartedAt).Milliseconds()
			return nil, trace, fmt.Errorf("failed to search vectors: %w", searchErr)
		}
		if len(kbResults) == 0 && scoreThreshold > 0 {
			s.logEmptySearchDiagnostics(ctx, provider, collectionName, embeddingResult.Vector, topK, scoreThreshold, []int64{knowledgeBase.ID}, req)
		}
		searchResults = append(searchResults, kbResults...)
	}
	trace.VectorSearchMs = time.Since(vectorSearchStartedAt).Milliseconds()

	if len(searchResults) > 0 {
		sort.SliceStable(searchResults, func(i, j int) bool {
			if searchResults[i].Score == searchResults[j].Score {
				return searchResults[i].ID < searchResults[j].ID
			}
			return searchResults[i].Score > searchResults[j].Score
		})
	}

	return searchResults, trace, nil
}

func (s *retrieve) hydrateRetrieveResults(searchResults []vectordb.SearchResult) ([]RetrieveResult, int64) {
	if len(searchResults) == 0 {
		return nil, 0
	}

	hydrateStartedAt := time.Now()
	results := make([]RetrieveResult, 0, len(searchResults))
	vectorIDs := make([]string, 0, len(searchResults))
	for _, sr := range searchResults {
		if strings.TrimSpace(sr.ID) == "" {
			continue
		}
		vectorIDs = append(vectorIDs, sr.ID)
	}
	chunks := repositories.KnowledgeChunkRepository.FindByVectorIDs(sqls.DB(), vectorIDs)
	chunkByVectorID := make(map[string]*models.KnowledgeChunk, len(chunks))
	documentIDs := make([]int64, 0)
	faqIDs := make([]int64, 0)
	documentSeen := make(map[int64]struct{})
	faqSeen := make(map[int64]struct{})
	for i := range chunks {
		chunk := &chunks[i]
		chunkByVectorID[chunk.VectorID] = chunk
		if chunk.DocumentID > 0 {
			if _, ok := documentSeen[chunk.DocumentID]; !ok {
				documentSeen[chunk.DocumentID] = struct{}{}
				documentIDs = append(documentIDs, chunk.DocumentID)
			}
		}
		if chunk.FaqID > 0 {
			if _, ok := faqSeen[chunk.FaqID]; !ok {
				faqSeen[chunk.FaqID] = struct{}{}
				faqIDs = append(faqIDs, chunk.FaqID)
			}
		}
	}
	documents := repositories.KnowledgeDocumentRepository.FindByIDs(sqls.DB(), documentIDs)
	documentByID := make(map[int64]*models.KnowledgeDocument, len(documents))
	for i := range documents {
		document := &documents[i]
		documentByID[document.ID] = document
	}
	faqs := repositories.KnowledgeFAQRepository.FindByIDs(sqls.DB(), faqIDs)
	faqByID := make(map[int64]*models.KnowledgeFAQ, len(faqs))
	for i := range faqs {
		faq := &faqs[i]
		faqByID[faq.ID] = faq
	}
	for _, sr := range searchResults {
		chunk := chunkByVectorID[sr.ID]
		if chunk == nil || chunk.Status != enums.StatusOk {
			continue
		}

		documentTitle := ""
		faqQuestion := ""
		if chunk.DocumentID > 0 {
			document := documentByID[chunk.DocumentID]
			if document == nil || document.Status != enums.StatusOk {
				continue
			}
			documentTitle = document.Title
		}
		if chunk.FaqID > 0 {
			faq := faqByID[chunk.FaqID]
			if faq == nil || faq.Status != enums.StatusOk {
				continue
			}
			faqQuestion = faq.Question
		}

		results = append(results, RetrieveResult{
			KnowledgeBaseID: chunk.KnowledgeBaseID,
			ChunkID:         chunk.ID,
			DocumentID:      chunk.DocumentID,
			DocumentTitle:   documentTitle,
			FaqID:           chunk.FaqID,
			FaqQuestion:     faqQuestion,
			ChunkNo:         chunk.ChunkNo,
			Title:           chunk.Title,
			SectionPath:     chunk.SectionPath,
			Content:         chunk.Content,
			Score:           sr.Score,
			ChunkType:       extractChunkType(sr.Payload),
		})
	}

	return results, time.Since(hydrateStartedAt).Milliseconds()
}
