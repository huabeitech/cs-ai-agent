package rag

import (
	"strings"
	"testing"

	"cs-agent/internal/models"
)

func TestBuildFAQChunkContent(t *testing.T) {
	faq := &models.KnowledgeFAQ{
		Question:         "如何退款",
		SimilarQuestions: `["退款怎么申请","申请售后"]`,
		Answer:           "在订单页发起退款。",
	}

	content := buildFAQChunkContent(faq)
	if !strings.Contains(content, "问题：如何退款") {
		t.Fatalf("expected question in content, got %q", content)
	}
	if !strings.Contains(content, "相似问：退款怎么申请；申请售后") {
		t.Fatalf("expected similar questions in content, got %q", content)
	}
	if !strings.Contains(content, "回答：在订单页发起退款。") {
		t.Fatalf("expected answer in content, got %q", content)
	}
}

func TestBuildFAQChunkModel(t *testing.T) {
	knowledgeBase := &models.KnowledgeBase{ID: 11}
	faq := &models.KnowledgeFAQ{ID: 22, Question: "如何退款"}

	chunk, chunkID := buildFAQChunkModel(knowledgeBase, faq, "问题：如何退款\n回答：在订单页发起退款。")
	if chunkID == "" {
		t.Fatalf("expected chunk id")
	}
	if chunk.KnowledgeBaseID != 11 || chunk.FaqID != 22 {
		t.Fatalf("unexpected chunk identity: %#v", chunk)
	}
	if chunk.VectorID != chunkID {
		t.Fatalf("expected vector id to match chunk id")
	}
	if chunk.Title != "如何退款" {
		t.Fatalf("unexpected title: %q", chunk.Title)
	}
}

func TestNormalizeContextResultsMergesAndDedupes(t *testing.T) {
	results := normalizeContextResults([]RetrieveResult{
		{DocumentID: 1, ChunkNo: 1, SectionPath: "A", Content: "第一段", Score: 0.7},
		{DocumentID: 1, ChunkNo: 2, SectionPath: "A", Content: "第二段", Score: 0.9},
		{DocumentID: 1, ChunkNo: 3, SectionPath: "A", Content: "第三段", Score: 0.6},
		{DocumentID: 2, ChunkNo: 1, Title: "标题", Content: "独立段", Score: 0.5},
		{FaqID: 9, FaqQuestion: "FAQ", Content: "FAQ内容", Score: 0.8},
		{FaqID: 9, FaqQuestion: "FAQ", Content: "FAQ重复", Score: 0.7},
	})

	if len(results) != 3 {
		t.Fatalf("expected 3 normalized results, got %d", len(results))
	}
	if !strings.Contains(results[0].Content, "第一段\n第二段") {
		t.Fatalf("expected adjacent document chunks to merge, got %q", results[0].Content)
	}
	if results[0].Score != 0.9 {
		t.Fatalf("expected merged score to keep max score, got %v", results[0].Score)
	}
	if results[1].Content != "独立段" {
		t.Fatalf("expected section duplicates to be removed after merge, got %q", results[1].Content)
	}
}

func TestBuildContextChunkText(t *testing.T) {
	faqText := buildContextChunkText(RetrieveResult{
		FaqID:       1,
		FaqQuestion: "如何退款",
		Content:     "在订单页发起退款。",
	})
	if !strings.Contains(faqText, "【FAQ：如何退款】") {
		t.Fatalf("unexpected faq context text: %q", faqText)
	}

	docText := buildContextChunkText(RetrieveResult{
		DocumentID:    2,
		DocumentTitle: "退款文档",
		SectionPath:   "售后/退款",
		Content:       "文档内容",
	})
	if !strings.Contains(docText, "【文档：退款文档｜章节：售后/退款】") {
		t.Fatalf("unexpected document context text: %q", docText)
	}
}
