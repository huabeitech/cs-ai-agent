package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"
)

type UserBuildOptions struct {
	Roles       bool
	Permissions bool
}

func BuildUserList(items []models.User, options UserBuildOptions) []response.UserResponse {
	results := make([]response.UserResponse, 0, len(items))
	for _, item := range items {
		results = append(results, *BuildUserResponse(&item, options))
	}
	return results
}

func BuildUserResponse(item *models.User, options UserBuildOptions) *response.UserResponse {
	if item == nil {
		return nil
	}
	ret := &response.UserResponse{
		ID:          item.ID,
		Username:    item.Username,
		Nickname:    item.Nickname,
		Avatar:      item.Avatar,
		Status:      item.Status,
		LastLoginAt: utils.FormatTimePtr(item.LastLoginAt),
		LastLoginIP: item.LastLoginIP,
	}

	if item.Mobile != nil {
		ret.Mobile = *item.Mobile
	}
	if item.Email != nil {
		ret.Email = *item.Email
	}

	if options.Roles {
		roleCodes, _ := services.AuthService.GetUserRoles(item.ID)
		ret.Roles = roleCodes
	}
	if options.Permissions {
		permissionCodes, _ := services.AuthService.GetUserPermissions(item.ID)
		ret.Permissions = permissionCodes
	}
	return ret
}
