package rag

import (
	"encoding/json"
	"fmt"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/response"

	"github.com/google/uuid"
	"github.com/mlogclub/simple/sqls"
)

var RetrieveLog = &retrieveLog{}

type retrieveLog struct {
}

type CreateRetrieveLogRequest struct {
	KnowledgeBaseID    int64
	Channel            string
	Scene              string
	SessionID          string
	ConversationID     int64
	Question           string
	RewriteQuestion    string
	Answer             string
	AnswerStatus       int
	ChunkProvider      string
	ChunkTargetTokens  int
	ChunkMaxTokens     int
	ChunkOverlapTokens int
	RerankEnabled      bool
	RerankLimit        int
	Hits               []response.KnowledgeSearchResult
	UsedHits           []response.KnowledgeSearchResult
	Citations          []response.KnowledgeCitation
	LatencyMs          int64
	RetrieveMs         int64
	GenerateMs         int64
	PromptTokens       int
	CompletionTokens   int
	ModelName          string
}

type retrieveTraceData struct {
	Retrieve    retrieveTraceRetrieve    `json:"retrieve"`
	ChunkConfig retrieveTraceChunkConfig `json:"chunkConfig"`
	Context     retrieveTraceContext     `json:"context"`
	Citations   []retrieveTraceCitation  `json:"citations"`
}

type retrieveTraceRetrieve struct {
	Provider        string `json:"provider"`
	RerankEnabled   bool   `json:"rerankEnabled"`
	RerankLimit     int    `json:"rerankLimit"`
	RawHitCount     int    `json:"rawHitCount"`
	ContextHitCount int    `json:"contextHitCount"`
	CitationCount   int    `json:"citationCount"`
}

type retrieveTraceChunkConfig struct {
	Provider      string `json:"provider"`
	TargetTokens  int    `json:"targetTokens"`
	MaxTokens     int    `json:"maxTokens"`
	OverlapTokens int    `json:"overlapTokens"`
}

type retrieveTraceContext struct {
	KnowledgeBaseIDs []int64  `json:"knowledgeBaseIds"`
	DocumentIDs      []int64  `json:"documentIds"`
	SectionPaths     []string `json:"sectionPaths"`
	UsedChunkKeys    []string `json:"usedChunkKeys"`
}

type retrieveTraceCitation struct {
	DocumentID  int64  `json:"documentId"`
	ChunkNo     int    `json:"chunkNo"`
	SectionPath string `json:"sectionPath"`
}

func (s *retrieveLog) FindHitsByRetrieveLogID(retrieveLogID int64) []models.KnowledgeRetrieveHit {
	if retrieveLogID <= 0 {
		return nil
	}
	var list []models.KnowledgeRetrieveHit
	sqls.DB().Where("retrieve_log_id = ?", retrieveLogID).Order("rank_no asc, id asc").Find(&list)
	return list
}

func (s *retrieveLog) CreateRetrieveLog(req *CreateRetrieveLogRequest, _ *dto.AuthPrincipal) (*models.KnowledgeRetrieveLog, error) {
	if req == nil {
		return nil, fmt.Errorf("retrieve log request is nil")
	}
	now := time.Now()
	topScore := 0.0
	if len(req.Hits) > 0 {
		topScore = req.Hits[0].Score
	}
	traceData := buildRetrieveTraceData(req)

	log := &models.KnowledgeRetrieveLog{
		KnowledgeBaseID:    req.KnowledgeBaseID,
		Channel:            req.Channel,
		Scene:              req.Scene,
		SessionID:          req.SessionID,
		ConversationID:     req.ConversationID,
		RequestID:          uuid.New().String(),
		Question:           req.Question,
		RewriteQuestion:    req.RewriteQuestion,
		Answer:             req.Answer,
		AnswerStatus:       req.AnswerStatus,
		HitCount:           len(req.Hits),
		TopScore:           topScore,
		ChunkProvider:      req.ChunkProvider,
		ChunkTargetTokens:  req.ChunkTargetTokens,
		ChunkMaxTokens:     req.ChunkMaxTokens,
		ChunkOverlapTokens: req.ChunkOverlapTokens,
		RerankEnabled:      req.RerankEnabled,
		RerankLimit:        req.RerankLimit,
		CitationCount:      len(req.Citations),
		UsedChunkCount:     len(req.UsedHits),
		LatencyMs:          req.LatencyMs,
		RetrieveMs:         req.RetrieveMs,
		GenerateMs:         req.GenerateMs,
		PromptTokens:       req.PromptTokens,
		CompletionTokens:   req.CompletionTokens,
		ModelName:          req.ModelName,
		TraceData:          traceData,
		CreatedAt:          now,
	}

	usedHitKeys := make(map[string]struct{}, len(req.UsedHits))
	for _, item := range req.UsedHits {
		usedHitKeys[buildKnowledgeSearchResultKey(item)] = struct{}{}
	}
	citationKeys := make(map[string]struct{}, len(req.Citations))
	for _, item := range req.Citations {
		citationKeys[buildKnowledgeCitationKey(item)] = struct{}{}
	}

	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(log).Error; err != nil {
			return err
		}
		for i, hit := range req.Hits {
			hitKey := buildKnowledgeSearchResultKey(hit)
			hitRecord := &models.KnowledgeRetrieveHit{
				RetrieveLogID:   log.ID,
				KnowledgeBaseID: hit.KnowledgeBaseID,
				ChunkID:         hit.ChunkID,
				DocumentID:      hit.DocumentID,
				DocumentTitle:   hit.DocumentTitle,
				ChunkNo:         hit.ChunkNo,
				Title:           hit.Title,
				SectionPath:     hit.SectionPath,
				ChunkType:       "",
				Provider:        req.ChunkProvider,
				RankNo:          i + 1,
				Score:           hit.Score,
				RerankScore:     hit.RerankScore,
				UsedInAnswer:    hasHitKey(usedHitKeys, hitKey),
				IsCitation:      hasHitKey(citationKeys, hitKey),
				Snippet:         hit.Content,
				CreatedAt:       now,
			}
			if err := ctx.Tx.Create(hitRecord).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return log, nil
}

func buildRetrieveTraceData(req *CreateRetrieveLogRequest) string {
	trace := retrieveTraceData{
		Retrieve: retrieveTraceRetrieve{
			Provider:        req.ChunkProvider,
			RerankEnabled:   req.RerankEnabled,
			RerankLimit:     req.RerankLimit,
			RawHitCount:     len(req.Hits),
			ContextHitCount: len(req.UsedHits),
			CitationCount:   len(req.Citations),
		},
		ChunkConfig: retrieveTraceChunkConfig{
			Provider:      req.ChunkProvider,
			TargetTokens:  req.ChunkTargetTokens,
			MaxTokens:     req.ChunkMaxTokens,
			OverlapTokens: req.ChunkOverlapTokens,
		},
		Context: retrieveTraceContext{
			KnowledgeBaseIDs: distinctKnowledgeBaseIDs(req.UsedHits),
			DocumentIDs:      distinctDocumentIDs(req.UsedHits),
			SectionPaths:     distinctSectionPaths(req.UsedHits),
			UsedChunkKeys:    buildUsedChunkKeys(req.UsedHits),
		},
		Citations: buildTraceCitations(req.Citations),
	}
	data, err := json.Marshal(trace)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildTraceCitations(citations []response.KnowledgeCitation) []retrieveTraceCitation {
	items := make([]retrieveTraceCitation, 0, len(citations))
	for _, item := range citations {
		items = append(items, retrieveTraceCitation{
			DocumentID:  item.DocumentID,
			ChunkNo:     item.ChunkNo,
			SectionPath: item.SectionPath,
		})
	}
	return items
}

func buildUsedChunkKeys(hits []response.KnowledgeSearchResult) []string {
	keys := make([]string, 0, len(hits))
	for _, item := range hits {
		keys = append(keys, buildKnowledgeSearchResultKey(item))
	}
	return keys
}

func distinctKnowledgeBaseIDs(hits []response.KnowledgeSearchResult) []int64 {
	ids := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, item := range hits {
		if item.KnowledgeBaseID <= 0 {
			continue
		}
		if _, ok := seen[item.KnowledgeBaseID]; ok {
			continue
		}
		seen[item.KnowledgeBaseID] = struct{}{}
		ids = append(ids, item.KnowledgeBaseID)
	}
	return ids
}

func distinctDocumentIDs(hits []response.KnowledgeSearchResult) []int64 {
	seen := make(map[int64]struct{})
	items := make([]int64, 0)
	for _, item := range hits {
		if item.DocumentID <= 0 {
			continue
		}
		if _, ok := seen[item.DocumentID]; ok {
			continue
		}
		seen[item.DocumentID] = struct{}{}
		items = append(items, item.DocumentID)
	}
	return items
}

func distinctSectionPaths(hits []response.KnowledgeSearchResult) []string {
	seen := make(map[string]struct{})
	items := make([]string, 0)
	for _, item := range hits {
		sectionPath := item.SectionPath
		if sectionPath == "" {
			continue
		}
		if _, ok := seen[sectionPath]; ok {
			continue
		}
		seen[sectionPath] = struct{}{}
		items = append(items, sectionPath)
	}
	return items
}

func buildKnowledgeSearchResultKey(item response.KnowledgeSearchResult) string {
	return fmt.Sprintf("%d|%s|%d", item.DocumentID, item.SectionPath, item.ChunkNo)
}

func buildKnowledgeCitationKey(item response.KnowledgeCitation) string {
	return fmt.Sprintf("%d|%s|%d", item.DocumentID, item.SectionPath, item.ChunkNo)
}

func hasHitKey(items map[string]struct{}, key string) bool {
	_, ok := items[key]
	return ok
}
