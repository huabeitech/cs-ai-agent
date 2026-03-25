package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
)

func BuildKnowledgeBaseResponse(item *models.KnowledgeBase) response.KnowledgeBaseResponse {
	return response.KnowledgeBaseResponse{
		ID:                    item.ID,
		Name:                  item.Name,
		Description:           item.Description,
		Status:                item.Status,
		StatusName:            enums.GetStatusLabel(item.Status),
		DefaultTopK:           item.DefaultTopK,
		DefaultScoreThreshold: item.DefaultScoreThreshold,
		DefaultRerankLimit:    item.DefaultRerankLimit,
		ChunkProvider:         item.ChunkProvider,
		ChunkTargetTokens:     item.ChunkTargetTokens,
		ChunkMaxTokens:        item.ChunkMaxTokens,
		ChunkOverlapTokens:    item.ChunkOverlapTokens,
		AnswerMode:            item.AnswerMode,
		AnswerModeName:        enums.GetKnowledgeAnswerModeLabel(enums.KnowledgeAnswerMode(item.AnswerMode)),
		FallbackMode:          item.FallbackMode,
		FallbackModeName:      enums.GetKnowledgeFallbackModeLabel(enums.KnowledgeFallbackMode(item.FallbackMode)),
		Remark:                item.Remark,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
		CreateUserName:        item.CreateUserName,
		UpdateUserName:        item.UpdateUserName,
	}
}

func BuildKnowledgeDocumentResponse(item *models.KnowledgeDocument) response.KnowledgeDocumentResponse {
	return response.KnowledgeDocumentResponse{
		ID:              item.ID,
		KnowledgeBaseID: item.KnowledgeBaseID,
		Title:           item.Title,
		Status:          item.Status,
		StatusName:      enums.GetStatusLabel(item.Status),
		ContentHash:     item.ContentHash,
		ContentType:     item.ContentType,
		Content:         item.Content,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		CreateUserName:  item.CreateUserName,
		UpdateUserName:  item.UpdateUserName,
	}
}

func BuildKnowledgeChunkResponse(item *models.KnowledgeChunk) response.KnowledgeChunkResponse {
	return response.KnowledgeChunkResponse{
		ID:              item.ID,
		KnowledgeBaseID: item.KnowledgeBaseID,
		DocumentID:      item.DocumentID,
		DocumentTitle:   "", // TODO
		ChunkNo:         item.ChunkNo,
		Title:           item.Title,
		Content:         item.Content,
		ContentHash:     item.ContentHash,
		CharCount:       item.CharCount,
		TokenCount:      item.TokenCount,
		Status:          item.Status,
		StatusName:      enums.GetStatusLabel(item.Status),
		VectorID:        item.VectorID,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func BuildKnowledgeRetrieveLogResponse(item *models.KnowledgeRetrieveLog) response.KnowledgeRetrieveLogResponse {
	return response.KnowledgeRetrieveLogResponse{
		ID:                 item.ID,
		KnowledgeBaseID:    item.KnowledgeBaseID,
		Channel:            item.Channel,
		ChannelName:        enums.GetKnowledgeRetrieveChannelLabel(enums.KnowledgeRetrieveChannel(item.Channel)),
		Scene:              item.Scene,
		SceneName:          enums.GetKnowledgeRetrieveSceneLabel(enums.KnowledgeRetrieveScene(item.Scene)),
		SessionID:          item.SessionID,
		ConversationID:     item.ConversationID,
		RequestID:          item.RequestID,
		Question:           item.Question,
		RewriteQuestion:    item.RewriteQuestion,
		Answer:             item.Answer,
		AnswerStatus:       item.AnswerStatus,
		AnswerStatusName:   enums.GetKnowledgeAnswerStatusLabel(enums.KnowledgeAnswerStatus(item.AnswerStatus)),
		HitCount:           item.HitCount,
		TopScore:           item.TopScore,
		ChunkProvider:      item.ChunkProvider,
		ChunkTargetTokens:  item.ChunkTargetTokens,
		ChunkMaxTokens:     item.ChunkMaxTokens,
		ChunkOverlapTokens: item.ChunkOverlapTokens,
		RerankEnabled:      item.RerankEnabled,
		RerankLimit:        item.RerankLimit,
		CitationCount:      item.CitationCount,
		UsedChunkCount:     item.UsedChunkCount,
		LatencyMs:          item.LatencyMs,
		RetrieveMs:         item.RetrieveMs,
		GenerateMs:         item.GenerateMs,
		PromptTokens:       item.PromptTokens,
		CompletionTokens:   item.CompletionTokens,
		ModelName:          item.ModelName,
		TraceData:          item.TraceData,
		CreatedAt:          item.CreatedAt,
	}
}

func BuildKnowledgeRetrieveHitResponse(item *models.KnowledgeRetrieveHit) response.KnowledgeRetrieveHitResponse {
	return response.KnowledgeRetrieveHitResponse{
		ID:            item.ID,
		RetrieveLogID: item.RetrieveLogID,
		ChunkID:       item.ChunkID,
		DocumentID:    item.DocumentID,
		DocumentTitle: item.DocumentTitle,
		ChunkNo:       item.ChunkNo,
		Title:         item.Title,
		SectionPath:   item.SectionPath,
		ChunkType:     item.ChunkType,
		ChunkTypeName: enums.GetKnowledgeChunkTypeLabel(enums.KnowledgeChunkType(item.ChunkType)),
		Provider:      item.Provider,
		RankNo:        item.RankNo,
		Score:         item.Score,
		RerankScore:   item.RerankScore,
		UsedInAnswer:  item.UsedInAnswer,
		IsCitation:    item.IsCitation,
		Snippet:       item.Snippet,
		CreatedAt:     item.CreatedAt,
	}
}
