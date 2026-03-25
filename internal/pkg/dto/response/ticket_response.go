package response

type TicketCategoryResponse struct {
	ID          int64                    `json:"id"`
	ParentID    int64                    `json:"parentId"`
	Name        string                   `json:"name"`
	Code        string                   `json:"code"`
	Description string                   `json:"description"`
	SortNo      int                      `json:"sortNo"`
	Status      int                      `json:"status"`
	Remark      string                   `json:"remark"`
	Children    []TicketCategoryResponse `json:"children,omitempty"`
}

type TicketResponse struct {
	ID                 int64  `json:"id"`
	TicketNo           string `json:"ticketNo"`
	Title              string `json:"title"`
	Content            string `json:"content"`
	ChannelType        string `json:"channelType"`
	ChannelID          string `json:"channelId"`
	CategoryID         int64  `json:"categoryId"`
	Priority           int    `json:"priority"`
	Status             int    `json:"status"`
	SourceUserID       int64  `json:"sourceUserId"`
	ExternalUserID     string `json:"externalUserId"`
	ExternalUserName   string `json:"externalUserName"`
	ExternalUserEmail  string `json:"externalUserEmail"`
	ExternalUserMobile string `json:"externalUserMobile"`
	CurrentAssigneeID  int64  `json:"currentAssigneeId"`
	CurrentTeamID      int64  `json:"currentTeamId"`
	ConversationID     int64  `json:"conversationId"`
	ReplyCount         int    `json:"replyCount"`
	Satisfied          *int   `json:"satisfied"`
	SatisfiedRemark    string `json:"satisfiedRemark"`
	Tags               string `json:"tags"`
	Remark             string `json:"remark"`
	CreatedAt          int64  `json:"createdAt"`
	UpdatedAt          int64  `json:"updatedAt"`
	ResolvedAt         *int64 `json:"resolvedAt"`
	ClosedAt           *int64 `json:"closedAt"`
}

type TicketReplyResponse struct {
	ID            int64  `json:"id"`
	TicketID      int64  `json:"ticketId"`
	ParentID      int64  `json:"parentId"`
	Content       string `json:"content"`
	SenderType    string `json:"senderType"`
	SenderID      int64  `json:"senderId"`
	SenderName    string `json:"senderName"`
	IsInternal    bool   `json:"isInternal"`
	SendStatus    int    `json:"sendStatus"`
	AttachmentIDs string `json:"attachmentIds"`
	CreatedAt     int64  `json:"createdAt"`
	SentAt        *int64 `json:"sentAt"`
	ReadAt        *int64 `json:"readAt"`
}
