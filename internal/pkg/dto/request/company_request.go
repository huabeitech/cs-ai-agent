package request

type CreateCompanyRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Industry string `json:"industry"`
	Website  string `json:"website"`
	Province string `json:"province"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Remark   string `json:"remark"`
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
