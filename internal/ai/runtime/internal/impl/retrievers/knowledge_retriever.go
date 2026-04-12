package retrievers

import (
	"context"
	"strings"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/utils"
)

const defaultRuntimeKnowledgeContextTokens = 4000

type KnowledgeRetriever struct {
	AIAgent *models.AIAgent
}

type KnowledgeRetrieveResult struct {
	KnowledgeBaseIDs []int64
	Query            string
	Hits             []rag.RetrieveResult
	ContextResults   []rag.RetrieveResult
	ContextText      string
	Trace            *rag.RetrieveTrace
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

func (r *KnowledgeRetriever) RetrieveContext(ctx context.Context, query string) (*KnowledgeRetrieveResult, error) {
	query = strings.TrimSpace(query)
	knowledgeBaseIDs := r.KnowledgeBaseIDs()
	ret := &KnowledgeRetrieveResult{
		KnowledgeBaseIDs: append([]int64(nil), knowledgeBaseIDs...),
		Query:            query,
	}
	if query == "" || len(knowledgeBaseIDs) == 0 {
		return ret, nil
	}
	results, trace, err := r.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}
	ret.Hits = append([]rag.RetrieveResult(nil), results...)
	ret.Trace = trace
	ret.ContextResults = rag.Retrieve.SelectContextResults(results, defaultRuntimeKnowledgeContextTokens)
	ret.ContextText = strings.TrimSpace(rag.Retrieve.BuildContext(ctx, results, defaultRuntimeKnowledgeContextTokens))
	return ret, nil
}
