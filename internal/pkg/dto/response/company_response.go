package response

import "cs-agent/internal/pkg/enums"

type CompanyResponse struct {
	ID            int64        `json:"id"`
	Name          string       `json:"name"`
	Code          string       `json:"code"`
	CustomerCount int64        `json:"customerCount"`
	Status        enums.Status `json:"status"`
	Remark        string       `json:"remark"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
}
