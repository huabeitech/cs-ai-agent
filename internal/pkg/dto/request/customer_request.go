package request

// CustomerListRequest 客户分页列表查询（POST /customer/list JSON Body）。
type CustomerListRequest struct {
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	Status        *int   `json:"status,omitempty"`
	Gender        *int   `json:"gender,omitempty"`
	CompanyID     *int64 `json:"companyId,omitempty"`
	Name          string `json:"name"`
	PrimaryMobile string `json:"primaryMobile"`
	PrimaryEmail  string `json:"primaryEmail"`
}

// Offset 分页偏移（需先保证 Page、Limit 已规范化为正数）。
func (r CustomerListRequest) Offset() int {
	if r.Page <= 0 || r.Limit <= 0 {
		return 0
	}
	return (r.Page - 1) * r.Limit
}

type CreateCustomerRequest struct {
	Name          string `json:"name"`
	Gender        int    `json:"gender"`
	CompanyID     int64  `json:"companyId"`
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

// CustomerProfileContactItem 保存客户档案时的联系方式行（无 id 表示新建）。
type CustomerProfileContactItem struct {
	ID           *int64 `json:"id,omitempty"`
	ContactType  string `json:"contactType"`
	ContactValue string `json:"contactValue"`
	IsPrimary    bool   `json:"isPrimary"`
	Remark       string `json:"remark"`
}

// SaveCustomerProfileRequest 客户主信息与联系方式一并保存（单事务）；id 为空或 0 表示新建客户。
type SaveCustomerProfileRequest struct {
	ID        *int64                       `json:"id,omitempty"`
	Name      string                       `json:"name"`
	Gender    int                          `json:"gender"`
	CompanyID int64                        `json:"companyId"`
	Remark    string                       `json:"remark"`
	Contacts  []CustomerProfileContactItem `json:"contacts"`
}
