package enums

type AIProvider string

const (
	AIProviderOpenAI AIProvider = "openai"
)

var aiProviderLabelMap = map[AIProvider]string{
	AIProviderOpenAI: "OpenAI",
}

func GetAIProviderLabel(provider AIProvider) string {
	return aiProviderLabelMap[provider]
}

type AIModelType string

const (
	AIModelTypeLLM       AIModelType = "llm"
	AIModelTypeEmbedding AIModelType = "embedding"
	AIModelTypeRerank    AIModelType = "rerank"
)

var aiModelTypeLabelMap = map[AIModelType]string{
	AIModelTypeLLM:       "大语言模型",
	AIModelTypeEmbedding: "向量模型",
	AIModelTypeRerank:    "重排序模型",
}

func GetAIModelTypeLabel(modelType AIModelType) string {
	return aiModelTypeLabelMap[modelType]
}
