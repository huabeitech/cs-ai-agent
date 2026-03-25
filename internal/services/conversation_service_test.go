package services

import (
	"cs-agent/internal/pkg/dto"
	"testing"
)

func TestBuildDefaultSubject(t *testing.T) {
	tests := []struct {
		name     string
		operator *dto.AuthPrincipal
		expected string
	}{
		{
			name: "访客-正常ID",
			operator: &dto.AuthPrincipal{
				IsVisitor: true,
				VisitorID: "visitor_550e8400-e29b-41d4-a716-446655440000",
			},
			expected: "访客a3f5b2c1",
		},
		{
			name: "访客-空ID",
			operator: &dto.AuthPrincipal{
				IsVisitor: true,
				VisitorID: "",
			},
			expected: "访客unknown",
		},
		{
			name: "登录用户-有用户名",
			operator: &dto.AuthPrincipal{
				IsVisitor: false,
				UserID:    1001,
				Username:  "张三",
			},
			expected: "张三",
		},
		{
			name: "登录用户-无用户名",
			operator: &dto.AuthPrincipal{
				IsVisitor: false,
				UserID:    1002,
				Username:  "",
			},
			expected: "用户1002",
		},
		{
			name:     "operator为nil",
			operator: nil,
			expected: "访客unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newConversationService()
			result := svc.buildDefaultSubject(tt.operator)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHashUUID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"visitor_550e8400-e29b-41d4-a716-446655440000", "a3f5b2c1"},
		{"", "unknown"},
		{"abc", "90015098"},
	}

	for _, tt := range tests {
		result := hashUUID(tt.input)
		if result != tt.expected {
			t.Errorf("input=%s, expected %s, got %s", tt.input, tt.expected, result)
		}
	}
}
