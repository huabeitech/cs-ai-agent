package response

import "cs-agent/internal/pkg/enums"

type CustomerContactResponse struct {
	ID           int64             `json:"id"`
	CustomerID   int64             `json:"customerId"`
	ContactType  enums.ContactType `json:"contactType"`
	ContactValue string            `json:"contactValue"`
	IsPrimary    bool              `json:"isPrimary"`
	IsVerified   bool              `json:"isVerified"`
	VerifiedAt   string            `json:"verifiedAt,omitempty"`
	Source       string            `json:"source"`
	Status       enums.Status      `json:"status"`
	Remark       string            `json:"remark"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}
