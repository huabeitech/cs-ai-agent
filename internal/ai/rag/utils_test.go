package rag

import (
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

func TestExtractStringPayloadNormalizesNonStringValue(t *testing.T) {
	got := ExtractStringPayload(map[string]interface{}{
		"title": "  hello\nworld  ",
		"count": 12,
	}, "count")
	want := "12"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExtractInt64PayloadSupportsStringValue(t *testing.T) {
	got := ExtractInt64Payload(map[string]interface{}{
		"document_id": "123",
	}, "document_id")
	if got != 123 {
		t.Fatalf("expected %d, got %d", 123, got)
	}
}
