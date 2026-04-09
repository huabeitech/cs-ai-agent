package request

type CreateAssetRequest struct {
	Prefix string `form:"prefix"`
}

type DeleteAssetRequest struct {
	ID int64 `json:"id"`
}
