package response

import "cs-agent/internal/pkg/enums"

type CompanyResponse struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Code      string       `json:"code"`
	Industry  string       `json:"industry"`
	Website   string       `json:"website"`
	Province  string       `json:"province"`
	City      string       `json:"city"`
	Address   string       `json:"address"`
	Status    enums.Status `json:"status"`
	Remark    string       `json:"remark"`
	CreatedAt string       `json:"createdAt"`
	UpdatedAt string       `json:"updatedAt"`
}
