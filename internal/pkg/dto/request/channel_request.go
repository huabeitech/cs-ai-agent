package request

type CreateChannelRequest struct {
	ChannelType string `json:"channelType"`
	AIAgentID   int64  `json:"aiAgentId"`
	Name        string `json:"name"`
	AppID       string `json:"appId"`
	ConfigJSON  string `json:"configJson"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
}

type UpdateChannelRequest struct {
	ID int64 `json:"id"`
	CreateChannelRequest
}

type UpdateChannelStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type DeleteChannelRequest struct {
	ID int64 `json:"id"`
}
