package response

import "cs-agent/internal/pkg/enums"

type AuthUserResponse struct {
	ID       int64        `json:"id"`
	Username string       `json:"username"`
	Nickname string       `json:"nickname"`
	Avatar   string       `json:"avatar"`
	Status   enums.Status `json:"status"`
	Roles    []string     `json:"roles"`
}

type LoginResponse struct {
	AccessToken  string            `json:"accessToken"`
	RefreshToken string            `json:"refreshToken"`
	ExpiresAt    string            `json:"expiresAt"`
	User         *AuthUserResponse `json:"user"`
	Permissions  []string          `json:"permissions"`
	Roles        []string          `json:"roles"`
}
