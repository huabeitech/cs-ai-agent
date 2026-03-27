package response

import "cs-agent/internal/pkg/enums"

type PermissionResponse struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Code      string       `json:"code"`
	Type      string       `json:"type"`
	GroupName string       `json:"groupName"`
	Method    string       `json:"method"`
	ApiPath   string       `json:"apiPath"`
	Status    enums.Status `json:"status"`
	SortNo    int          `json:"sortNo"`
}

type RoleResponse struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Status      enums.Status `json:"status"`
	IsSystem    bool         `json:"isSystem"`
	SortNo      int          `json:"sortNo"`
	Permissions []string     `json:"permissions,omitempty"`
}

type UserResponse struct {
	ID          int64          `json:"id"`
	Username    string         `json:"username"`
	Nickname    string         `json:"nickname"`
	Avatar      string         `json:"avatar"`
	Mobile      string         `json:"mobile,omitempty"`
	Email       string         `json:"email,omitempty"`
	Status      enums.Status   `json:"status"`
	LastLoginAt string         `json:"lastLoginAt,omitempty"`
	LastLoginIP string         `json:"lastLoginIp,omitempty"`
	Roles       []RoleResponse `json:"assignedRoles,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
}

type SessionResponse struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"userId"`
	Username   string `json:"username,omitempty"`
	TokenType  string `json:"tokenType"`
	ClientType string `json:"clientType"`
	ClientIP   string `json:"clientIp"`
	UserAgent  string `json:"userAgent"`
	ExpiredAt  string `json:"expiredAt"`
	RevokedAt  string `json:"revokedAt,omitempty"`
	LastSeenAt string `json:"lastSeenAt,omitempty"`
}
