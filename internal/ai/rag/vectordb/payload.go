package vectordb

import (
	"github.com/spf13/cast"
)

type ChunkPayload struct {
	KnowledgeBaseID int64  `json:"knowledgeBaseId"`
	DocumentID      int64  `json:"documentId"`
	DocumentTitle   string `json:"documentTitle"`
	ChunkNo         int    `json:"chunkNo"`
	ChunkType       string `json:"chunkType"`
	SectionPath     string `json:"sectionPath"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Provider        string `json:"provider"`
}

func (p ChunkPayload) ToMap() map[string]any {
	return map[string]any{
		"knowledge_base_id": p.KnowledgeBaseID,
		"document_id":       p.DocumentID,
		"document_title":    p.DocumentTitle,
		"chunk_no":          p.ChunkNo,
		"chunk_type":        p.ChunkType,
		"section_path":      p.SectionPath,
		"title":             p.Title,
		"content":           p.Content,
		"provider":          p.Provider,
	}
}

func ChunkPayloadFromMap(data map[string]any) ChunkPayload {
	if data == nil {
		return ChunkPayload{}
	}
	return ChunkPayload{
		KnowledgeBaseID: cast.ToInt64(data["knowledge_base_id"]),
		DocumentID:      cast.ToInt64(data["document_id"]),
		DocumentTitle:   cast.ToString(data["document_title"]),
		ChunkNo:         cast.ToInt(data["chunk_no"]),
		ChunkType:       cast.ToString(data["chunk_type"]),
		SectionPath:     cast.ToString(data["section_path"]),
		Title:           cast.ToString(data["title"]),
		Content:         cast.ToString(data["content"]),
		Provider:        cast.ToString(data["provider"]),
	}
}
