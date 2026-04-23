package request

import "cs-agent/internal/pkg/enums"

type RevokeSessionRequest struct {
	ID int64 `json:"id"`
}

type RevokeUserSessionsRequest struct {
	UserID int64 `json:"userId"`
}

type CreateUserRequest struct {
	Username string  `json:"username"`
	Nickname string  `json:"nickname"`
	Avatar   string  `json:"avatar"`
	Mobile   *string `json:"mobile"`
	Email    *string `json:"email"`
	Remark   string  `json:"remark"`
	RoleIDs  []int64 `json:"roleIds"`
}

type UpdateUserRequest struct {
	ID       int64   `json:"id"`
	Nickname string  `json:"nickname"`
	Avatar   string  `json:"avatar"`
	Mobile   *string `json:"mobile"`
	Email    *string `json:"email"`
	Remark   string  `json:"remark"`
}

type DeleteUserRequest struct {
	ID int64 `json:"id"`
}

type UpdateUserStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type ChangePasswordRequest struct {
	Password string `json:"password"`
}

type AssignRoleRequest struct {
	UserID  int64   `json:"userId"`
	RoleIDs []int64 `json:"roleIds"`
}

type CreateRoleRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Remark string `json:"remark"`
}

type UpdateRoleRequest struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	SortNo int    `json:"sortNo"`
	Remark string `json:"remark"`
}

type DeleteRoleRequest struct {
	ID int64 `json:"id"`
}

type UpdateRoleStatusRequest struct {
	ID     int64        `json:"id"`
	Status enums.Status `json:"status"`
}

type AssignPermissionRequest struct {
	RoleID        int64   `json:"roleId"`
	PermissionIDs []int64 `json:"permissionIds"`
}

type CreateConversationTagRequest struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	Status int    `json:"status"`
	SortNo int    `json:"sortNo"`
	Remark string `json:"remark"`
}

type UpdateConversationTagRequest struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Status int    `json:"status"`
	SortNo int    `json:"sortNo"`
	Remark string `json:"remark"`
}

type DeleteConversationTagRequest struct {
	ID int64 `json:"id"`
}
