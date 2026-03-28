package request

import (
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/web/params"
)

type AgentConversationFilter string

const (
	AgentConversationFilterMine    AgentConversationFilter = "mine"
	AgentConversationFilterActive  AgentConversationFilter = "active"
	AgentConversationFilterPending AgentConversationFilter = "pending"
	AgentConversationFilterClosed  AgentConversationFilter = "closed"
)

type ConversationListRequest struct {
	Status            int    `json:"status"`
	ExternalSource    string `json:"externalSource"`
	ServiceMode       int    `json:"serviceMode"`
	CurrentAssigneeID int64  `json:"currentAssigneeId"`
	SourceUserID      int64  `json:"sourceUserId"`
	Keyword           string `json:"keyword"`
	TagID             int64  `json:"tagId"`
}

type AssignConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	AssigneeID     int64  `json:"assigneeId"`
	Reason         string `json:"reason"`
}

type DispatchConversationRequest struct {
	ConversationID int64 `json:"conversationId"`
}

type TransferConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	ToUserID       int64  `json:"toUserId"`
	Reason         string `json:"reason"`
}

type CloseConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	CloseReason    string `json:"closeReason"`
}

type ReadConversationRequest struct {
	ConversationID int64 `json:"conversationId"`
	MessageID      int64 `json:"messageId"`
}

type AddConversationTagRequest struct {
	ConversationID int64 `json:"conversationId"`
	TagID          int64 `json:"tagId"`
}

type RemoveConversationTagRequest struct {
	ConversationID int64 `json:"conversationId"`
	TagID          int64 `json:"tagId"`
}

type ExternalInfo struct {
	ExternalSource enums.ExternalSource `json:"externalSource"`
	ExternalID     string               `json:"externalId"`
	ExternalName   string               `json:"externalName"`
}

func GetExternalInfo(ctx iris.Context) (*ExternalInfo, error) {
	externalSource, err := getExternalSource(ctx)
	if err != nil {
		return nil, err
	}
	externalID, err := getExternalID(ctx)
	if err != nil {
		return nil, err
	}
	return &ExternalInfo{
		ExternalSource: externalSource,
		ExternalID:     externalID,
		ExternalName:   getExternalName(ctx),
	}, nil
}

func getExternalSource(ctx iris.Context) (enums.ExternalSource, error) {
	externalSource := ctx.GetHeader("X-External-Source")
	if strs.IsBlank(externalSource) {
		externalSource, _ = params.Get(ctx, "externalSource")
	}
	if strs.IsBlank(externalSource) {
		return "", errorsx.Unauthorized("用户来源不能为空")
	}
	return enums.ExternalSource(strings.TrimSpace(externalSource)), nil
}

func getExternalID(ctx iris.Context) (string, error) {
	externalID := ctx.GetHeader("X-External-Id")
	if strs.IsBlank(externalID) {
		externalID, _ = params.Get(ctx, "externalId")
	}
	if strs.IsBlank(externalID) {
		return "", errorsx.Unauthorized("用户标识不能为空")
	}
	return strings.TrimSpace(externalID), nil
}

func getExternalName(ctx iris.Context) string {
	externalName := ctx.GetHeader("X-External-Name")
	if strs.IsBlank(externalName) {
		externalName, _ = params.Get(ctx, "externalName")
	}
	return strings.TrimSpace(externalName)
}
