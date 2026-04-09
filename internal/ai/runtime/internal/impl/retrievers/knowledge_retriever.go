package retrievers

import (
	"context"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/utils"
)

type KnowledgeRetriever struct {
	AIAgent *models.AIAgent
}

func NewKnowledgeRetriever(aiAgent *models.AIAgent) *KnowledgeRetriever {
	return &KnowledgeRetriever{AIAgent: aiAgent}
}

func (r *KnowledgeRetriever) KnowledgeBaseIDs() []int64 {
	if r == nil || r.AIAgent == nil {
		return nil
	}
	return utils.SplitInt64s(r.AIAgent.KnowledgeIDs)
}

func (r *KnowledgeRetriever) Retrieve(ctx context.Context, query string) ([]rag.RetrieveResult, *rag.RetrieveTrace, error) {
	ids := r.KnowledgeBaseIDs()
	return rag.Retrieve.RetrieveWithTrace(ctx, rag.RetrieveRequest{
		Query:            query,
		KnowledgeBaseIDs: ids,
	})
}
