package request

type CreateWidgetSiteRequest struct {
	AIAgentID int64  `json:"aiAgentId"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
}

type UpdateWidgetSiteRequest struct {
	ID int64 `json:"id"`
	CreateWidgetSiteRequest
}

type UpdateWidgetSiteStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type DeleteWidgetSiteRequest struct {
	ID int64 `json:"id"`
}
