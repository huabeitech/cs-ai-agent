package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/pkg/enums"
)

type rerank struct{}

var Rerank = &rerank{}

func (s *rerank) Rerank(ctx context.Context, query string, documents []string, topN int) ([]RerankResult, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	if topN <= 0 {
		topN = len(documents)
	}

	results, err := s.callRerankAPI(ctx, query, documents, topN)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *rerank) callRerankAPI(ctx context.Context, query string, documents []string, topN int) ([]RerankResult, error) {
	config, err := ai.GetEnabledAIConfig(enums.AIModelTypeRerank)
	if err != nil {
		return nil, err
	}

	reqBody := RerankRequest{
		Model:     config.ModelName,
		Query:     query,
		Documents: documents,
		TopN:      topN,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.BaseURL+"/v1/rerank", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{
		Timeout: time.Duration(config.TimeoutMS) * time.Millisecond,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call rerank API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var rerankResp RerankResponse
	if err := json.Unmarshal(body, &rerankResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	results := make([]RerankResult, 0, len(rerankResp.Results))
	for _, r := range rerankResp.Results {
		results = append(results, RerankResult{
			Index:          r.Index,
			RelevanceScore: r.RelevanceScore,
		})
	}

	return results, nil
}

func (s *rerank) RerankResults(ctx context.Context, query string, results []RetrieveResult, topN int) ([]RetrieveResult, error) {
	if len(results) == 0 {
		return nil, nil
	}

	if topN <= 0 {
		topN = len(results)
	}

	documents := make([]string, 0, len(results))
	for _, r := range results {
		documents = append(documents, r.Content)
	}

	rerankResults, err := s.Rerank(ctx, query, documents, topN)
	if err != nil {
		return nil, err
	}

	rerankedResults := make([]RetrieveResult, 0, len(rerankResults))
	for _, rr := range rerankResults {
		if rr.Index < len(results) {
			result := results[rr.Index]
			result.Score = float32(rr.RelevanceScore)
			rerankedResults = append(rerankedResults, result)
		}
	}

	return rerankedResults, nil
}

func (s *rerank) SimpleRerank(query string, results []RetrieveResult, topN int) []RetrieveResult {
	if len(results) == 0 {
		return nil
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topN > 0 && len(results) > topN {
		return results[:topN]
	}

	return results
}
