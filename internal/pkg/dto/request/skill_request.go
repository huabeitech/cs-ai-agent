package request

type SkillDefinitionListRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}

type CreateSkillDefinitionRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
	Remark      string `json:"remark"`
}

type UpdateSkillDefinitionRequest struct {
	ID int64 `json:"id"`
	CreateSkillDefinitionRequest
}

type DeleteSkillDefinitionRequest struct {
	ID int64 `json:"id"`
}

type UpdateSkillDefinitionStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}
