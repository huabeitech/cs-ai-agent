package request

import "cs-agent/internal/pkg/enums"

type TicketCategoryListRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	ParentID int64  `json:"parentId"`
	Status   int    `json:"status"`
}

type CreateTicketCategoryRequest struct {
	ParentID    int64        `json:"parentId"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Description string       `json:"description"`
	Status      enums.Status `json:"status"`
	Remark      string       `json:"remark"`
}

type UpdateTicketCategoryRequest struct {
	ID int64 `json:"id"`
	CreateTicketCategoryRequest
}

type DeleteTicketCategoryRequest struct {
	ID int64 `json:"id"`
}

type TicketListRequest struct {
	Title             string `json:"title"`
	CategoryID        int64  `json:"categoryId"`
	Status            int    `json:"status"`
	Priority          int    `json:"priority"`
	CurrentAssigneeID int64  `json:"currentAssigneeId"`
	ChannelType       string `json:"channelType"`
	Keyword           string `json:"keyword"`
}

type CreateTicketRequest struct {
	Title              string `json:"title"`
	Content            string `json:"content"`
	ChannelType        string `json:"channelType"`
	ChannelID          string `json:"channelId"`
	CategoryID         int64  `json:"categoryId"`
	Priority           int    `json:"priority"`
	SourceUserID       int64  `json:"sourceUserId"`
	ExternalUserID     string `json:"externalUserId"`
	ExternalUserName   string `json:"externalUserName"`
	ExternalUserEmail  string `json:"externalUserEmail"`
	ExternalUserMobile string `json:"externalUserMobile"`
	ConversationID     int64  `json:"conversationId"`
	Tags               string `json:"tags"`
	Remark             string `json:"remark"`
}

type UpdateTicketRequest struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CategoryID int64  `json:"categoryId"`
	Priority   int    `json:"priority"`
	Status     int    `json:"status"`
	Tags       string `json:"tags"`
	Remark     string `json:"remark"`
}

type DeleteTicketRequest struct {
	ID int64 `json:"id"`
}

type AssignTicketRequest struct {
	ID       int64  `json:"id"`
	ToUserID int64  `json:"toUserId"`
	ToTeamID int64  `json:"toTeamId"`
	Reason   string `json:"reason"`
}

type CloseTicketRequest struct {
	ID int64 `json:"id"`
}

type ReopenTicketRequest struct {
	ID int64 `json:"id"`
}

type TicketReplyListRequest struct {
	TicketID int64 `json:"ticketId"`
}

type CreateTicketReplyRequest struct {
	TicketID      int64  `json:"ticketId"`
	ParentID      int64  `json:"parentId"`
	Content       string `json:"content"`
	IsInternal    bool   `json:"isInternal"`
	AttachmentIDs string `json:"attachmentIds"`
}

type UpdateTicketReplyRequest struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

type DeleteTicketReplyRequest struct {
	ID int64 `json:"id"`
}
