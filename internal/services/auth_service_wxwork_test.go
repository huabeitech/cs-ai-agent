package services

import (
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/wxwork"
	"testing"
)

func TestWxWorkStateRoundTrip(t *testing.T) {
	wxwork.Init(&config.Config{
		WxWork: config.WxWorkConfig{
			Enabled:    true,
			CorpID:     "ww_test",
			CorpSecret: "secret",
		},
	})

	next, err := AuthService.createWxWorkState("/workspace")
	if err != nil {
		t.Fatalf("createWxWorkState failed: %v", err)
	}

	got, err := AuthService.parseWxWorkState(next)
	if err != nil {
		t.Fatalf("parseWxWorkState failed: %v", err)
	}
	if got != "/workspace" {
		t.Fatalf("expected /workspace, got %s", got)
	}
}

func TestWxWorkLoginTicketSingleUse(t *testing.T) {
	ticket, err := AuthService.issueWxWorkLoginTicket(nil)
	if err == nil {
		t.Fatal("expected nil login response to fail")
	}

	ticket, err = AuthService.issueWxWorkLoginTicket(&response.LoginResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
	})
	if err != nil {
		t.Fatalf("issueWxWorkLoginTicket failed: %v", err)
	}

	resp, err := AuthService.consumeWxWorkLoginTicket(ticket)
	if err != nil {
		t.Fatalf("consumeWxWorkLoginTicket failed: %v", err)
	}
	if resp.AccessToken != "access" {
		t.Fatalf("expected access token access, got %s", resp.AccessToken)
	}

	if _, err = AuthService.consumeWxWorkLoginTicket(ticket); err == nil {
		t.Fatal("expected consumed ticket to become invalid")
	}
}
