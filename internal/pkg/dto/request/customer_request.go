package request

type CreateCustomerRequest struct {
	Name          string `json:"name"`
	Gender        int    `json:"gender"`
	CompanyID     int64  `json:"companyId"`
	Province      string `json:"province"`
	City          string `json:"city"`
	PrimaryMobile string `json:"primaryMobile"`
	PrimaryEmail  string `json:"primaryEmail"`
	Remark        string `json:"remark"`
}

type UpdateCustomerRequest struct {
	ID int64 `json:"id"`
	CreateCustomerRequest
}

type DeleteCustomerRequest struct {
	ID int64 `json:"id"`
}

type UpdateCustomerStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}
