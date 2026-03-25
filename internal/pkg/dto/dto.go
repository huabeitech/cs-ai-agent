package dto

import "cs-agent/internal/pkg/enums"

type AuthPrincipal struct {
	UserID      int64
	Username    string
	Nickname    string
	Avatar      string
	Status      enums.Status
	IsVisitor   bool
	VisitorID   string
	Roles       []string
	Permissions []string
}
