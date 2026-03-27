package response

import "cs-agent/internal/pkg/enums"

type CustomerResponse struct {
	ID            int64        `json:"id"`
	Name          string       `json:"name"`
	Gender        enums.Gender `json:"gender"`
	CompanyID     int64        `json:"companyId"`
	LastActiveAt  string       `json:"lastActiveAt"`
	PrimaryMobile string       `json:"primaryMobile"`
	PrimaryEmail  string       `json:"primaryEmail"`
	Status        enums.Status `json:"status"`
	Remark        string       `json:"remark"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
}
