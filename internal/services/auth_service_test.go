package services

import "testing"

func TestExtractBearerToken(t *testing.T) {
	svc := newAuthService()

	if got := svc.extractBearerToken("Bearer token_123"); got != "token_123" {
		t.Fatalf("expected bearer token to be extracted, got %q", got)
	}

	if got := svc.extractBearerToken("token_123"); got != "" {
		t.Fatalf("expected raw token to be rejected by bearer extractor, got %q", got)
	}
}
