package request

// CustomerListRequest 客户分页列表查询（POST /customer/list JSON Body）。
type CustomerListRequest struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Status    *int   `json:"status,omitempty"`
	Gender    *int   `json:"gender,omitempty"`
	CompanyID *int64 `json:"companyId,omitempty"`
	// Keyword 模糊匹配：客户姓名、主手机号、主邮箱、任意联系方式（t_customer_contact）、公司名称（t_company）。
	Keyword string `json:"keyword"`
}

func (r CustomerListRequest) GetPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

func (r CustomerListRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 20
	}
	return r.Limit
}

func (r CustomerListRequest) Offset() int {
	return (r.GetPage() - 1) * r.GetLimit()
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
