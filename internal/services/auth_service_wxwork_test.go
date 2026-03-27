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

	next, err := wxwork.CreateState("/workspace")
	if err != nil {
		t.Fatalf("CreateState failed: %v", err)
	}

	got, err := wxwork.ParseState(next)
	if err != nil {
		t.Fatalf("ParseState failed: %v", err)
	}
	if got != "/workspace" {
		t.Fatalf("expected /workspace, got %s", got)
	}
}

func TestWxWorkLoginTicketSingleUse(t *testing.T) {
	ticket, err := wxwork.IssueLoginTicket(nil)
	if err == nil {
		t.Fatal("expected nil login response to fail")
	}

	ticket, err = wxwork.IssueLoginTicket(&response.LoginResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
	})
	if err != nil {
		t.Fatalf("IssueLoginTicket failed: %v", err)
	}

	resp, err := wxwork.ConsumeLoginTicket(ticket)
	if err != nil {
		t.Fatalf("ConsumeLoginTicket failed: %v", err)
	}
	if resp.AccessToken != "access" {
		t.Fatalf("expected access token access, got %s", resp.AccessToken)
	}

	if _, err = wxwork.ConsumeLoginTicket(ticket); err == nil {
		t.Fatal("expected consumed ticket to become invalid")
	}
}
