package callbacks

import "testing"

func TestRuntimeTraceCollectorAddRetrieverItems(t *testing.T) {
	collector := NewRuntimeTraceCollector()

	collector.AddRetrieverItems([]RetrieverTraceItem{
		{KnowledgeBaseID: 1, DocumentID: 10, DocumentTitle: "doc-1"},
		{KnowledgeBaseID: 2, DocumentID: 20, DocumentTitle: "doc-2"},
	})
	collector.AddRetrieverItems(nil)

	if got := len(collector.Data.Retriever.Items); got != 2 {
		t.Fatalf("expected two retriever items, got %d", got)
	}
	if collector.Data.Retriever.Items[0].DocumentTitle != "doc-1" || collector.Data.Retriever.Items[1].DocumentTitle != "doc-2" {
		t.Fatalf("unexpected retriever items: %#v", collector.Data.Retriever.Items)
	}
}
