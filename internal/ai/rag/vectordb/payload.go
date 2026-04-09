package vectordb

import (
	"github.com/mlogclub/simple/common/structs"
	"github.com/spf13/cast"
)

type ChunkPayload struct {
	KnowledgeBaseID int64  `json:"knowledge_base_id"`
	DocumentID      int64  `json:"document_id"`
	DocumentTitle   string `json:"document_title"`
	FaqID           int64  `json:"faq_id"`
	FaqQuestion     string `json:"faq_question"`
	ChunkNo         int    `json:"chunk_no"`
	ChunkType       string `json:"chunk_type"`
	SectionPath     string `json:"section_path"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Provider        string `json:"provider"`
}

func (p ChunkPayload) ToMap() map[string]any {
	return structs.StructToMap(p)
}

func ChunkPayloadFromMap(data map[string]any) ChunkPayload {
	if data == nil {
		return ChunkPayload{}
	}
	return ChunkPayload{
		KnowledgeBaseID: cast.ToInt64(data["knowledge_base_id"]),
		DocumentID:      cast.ToInt64(data["document_id"]),
		DocumentTitle:   cast.ToString(data["document_title"]),
		FaqID:           cast.ToInt64(data["faq_id"]),
		FaqQuestion:     cast.ToString(data["faq_question"]),
		ChunkNo:         cast.ToInt(data["chunk_no"]),
		ChunkType:       cast.ToString(data["chunk_type"]),
		SectionPath:     cast.ToString(data["section_path"]),
		Title:           cast.ToString(data["title"]),
		Content:         cast.ToString(data["content"]),
		Provider:        cast.ToString(data["provider"]),
	}
}
