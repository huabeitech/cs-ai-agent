package rag

import (
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/pkg/enums"
	"testing"
)

func TestExtractPlainTextFromHTMLSeparatesBlockContent(t *testing.T) {
	got := ExtractPlainTextFromHTML("<div>Hello</div><div>World</div><p>Again</p>")
	want := "Hello World Again"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExtractPlainTextMarkdownUsesGoldmark(t *testing.T) {
	got := ExtractPlainText("# Title\n\n- one\n- two", enums.KnowledgeDocumentContentTypeMarkdown)
	want := "Title one two"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestChunkPayloadFromMapSupportsTypedConversion(t *testing.T) {
	got := vectordb.ChunkPayloadFromMap(map[string]any{
		"knowledge_base_id": "1",
		"document_id":       "123",
		"document_title":    "Doc",
		"chunk_no":          "2",
		"chunk_type":        "text",
		"section_path":      "A > B",
		"title":             "hello",
		"content":           "world",
		"provider":          "structured",
	})
	if got.KnowledgeBaseID != 1 || got.DocumentID != 123 || got.ChunkNo != 2 {
		t.Fatalf("unexpected numeric conversion result: %+v", got)
	}
	if got.DocumentTitle != "Doc" || got.SectionPath != "A > B" || got.Provider != "structured" {
		t.Fatalf("unexpected string conversion result: %+v", got)
	}
}
