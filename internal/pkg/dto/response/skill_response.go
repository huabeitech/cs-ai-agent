package response

import "time"

type SkillDefinitionResponse struct {
	ID             int64     `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Prompt         string    `json:"prompt"`
	Priority       int       `json:"priority"`
	Status         int       `json:"status"`
	StatusName     string    `json:"statusName"`
	Remark         string    `json:"remark"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	CreateUserName string    `json:"createUserName"`
	UpdateUserName string    `json:"updateUserName"`
}
