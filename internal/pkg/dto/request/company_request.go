package request

type CreateCompanyRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Remark string `json:"remark"`
}

type UpdateCompanyRequest struct {
	ID int64 `json:"id"`
	CreateCompanyRequest
}

type DeleteCompanyRequest struct {
	ID int64 `json:"id"`
}

type UpdateCompanyStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}
