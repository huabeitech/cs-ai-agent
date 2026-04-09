package rag

type RetrieveRequest struct {
	KnowledgeBaseIDs []int64
	Query            string
	TopK             int
	ScoreThreshold   float64
}

type RetrieveResult struct {
	KnowledgeBaseID int64   `json:"knowledgeBaseId"`
	ChunkID         int64   `json:"chunkId"`
	DocumentID      int64   `json:"documentId"`
	DocumentTitle   string  `json:"documentTitle"`
	FaqID           int64   `json:"faqId"`
	FaqQuestion     string  `json:"faqQuestion"`
	ChunkNo         int     `json:"chunkNo"`
	Title           string  `json:"title"`
	SectionPath     string  `json:"sectionPath"`
	Content         string  `json:"content"`
	Score           float32 `json:"score"`
	ChunkType       string  `json:"chunkType"`
}

type RerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type RerankResponse struct {
	Results []struct {
		Document       string  `json:"document"`
		Index          int     `json:"index"`
		RelevanceScore float64 `json:"relevance_score"`
	} `json:"results"`
	Meta struct {
		APIVersion struct {
			Version string `json:"version"`
		} `json:"api_version"`
	} `json:"meta"`
}

type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevanceScore"`
}
