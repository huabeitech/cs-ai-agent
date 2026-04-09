package request

type TagListRequest struct {
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
}

type CreateTagRequest struct {
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Remark   string `json:"remark"`
}

type UpdateTagRequest struct {
	ID int64 `json:"id"`
	CreateTagRequest
}

type DeleteTagRequest struct {
	ID int64 `json:"id"`
}

type UpdateTagStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}
