package tools

import "testing"

func TestParseToolSearchRequest(t *testing.T) {
	req, err := parseToolSearchRequest(`{"query":"  search docs  ","toolCode":" mcp_server/search ","arguments":{"q":"hello"}}`)
	if err != nil {
		t.Fatalf("parseToolSearchRequest returned error: %v", err)
	}
	if req.Query != "search docs" {
		t.Fatalf("unexpected query: %q", req.Query)
	}
	if req.ToolCode != "mcp_server/search" {
		t.Fatalf("unexpected toolCode: %q", req.ToolCode)
	}
	if req.Arguments["q"] != "hello" {
		t.Fatalf("unexpected arguments: %#v", req.Arguments)
	}
}

func TestParseToolSearchRequestDefaultsArguments(t *testing.T) {
	req, err := parseToolSearchRequest(`{"query":"list"}`)
	if err != nil {
		t.Fatalf("parseToolSearchRequest returned error: %v", err)
	}
	if req.Arguments == nil {
		t.Fatalf("expected non-nil arguments map")
	}
	if len(req.Arguments) != 0 {
		t.Fatalf("expected empty arguments map, got %#v", req.Arguments)
	}
}
