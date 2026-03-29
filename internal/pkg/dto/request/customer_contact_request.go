package request

type CreateCustomerContactRequest struct {
	CustomerID   int64  `json:"customerId"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
	IsPrimary    bool   `json:"isPrimary"`
	IsVerified   bool   `json:"isVerified"`
	Source       string `json:"source"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
}

type UpdateCustomerContactRequest struct {
	ID           int64  `json:"id"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
	IsPrimary    bool   `json:"isPrimary"`
	IsVerified   bool   `json:"isVerified"`
	Source       string `json:"source"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
}
