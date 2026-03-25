package rag

import (
	"testing"

	"cs-agent/internal/pkg/dto/response"
)

func TestEvaluateExpectedDocumentHits(t *testing.T) {
	results := []response.KnowledgeSearchResult{
		{DocumentID: 11},
		{DocumentID: 22},
		{DocumentID: 33},
	}

	top1Matched, top3Matched, matched := evaluateExpectedDocumentHits(results, []int64{22, 44})
	if top1Matched {
		t.Fatalf("expected top1 not matched")
	}
	if !top3Matched {
		t.Fatalf("expected top3 matched")
	}
	if len(matched) != 1 || matched[0] != 22 {
		t.Fatalf("unexpected matched document ids: %+v", matched)
	}
}

func TestEvaluateExpectedDocumentHitsTop1(t *testing.T) {
	results := []response.KnowledgeSearchResult{
		{DocumentID: 11},
		{DocumentID: 22},
	}

	top1Matched, top3Matched, matched := evaluateExpectedDocumentHits(results, []int64{11})
	if !top1Matched || !top3Matched {
		t.Fatalf("expected both top1 and top3 matched")
	}
	if len(matched) != 1 || matched[0] != 11 {
		t.Fatalf("unexpected matched document ids: %+v", matched)
	}
}
