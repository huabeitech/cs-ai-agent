package enums

// 客户来源
type CustomerSourceType string

const (
	CustomerSourceTypeWebChat CustomerSourceType = "web_chat"
)

var customerSourceTypeLabelMap = map[CustomerSourceType]string{
	CustomerSourceTypeWebChat: "网页客服",
}

func GetCustomerSourceTypeLabel(sourceType CustomerSourceType) string {
	return customerSourceTypeLabelMap[sourceType]
}

// 联系方式类型
type ContactType string

const (
	ContactTypeMobile ContactType = "mobile"
	ContactTypeEmail  ContactType = "email"
	ContactTypeWeChat ContactType = "wechat"
	ContactTypeOther  ContactType = "other"
)

var contactTypeLabelMap = map[ContactType]string{
	ContactTypeMobile: "手机号",
	ContactTypeEmail:  "邮箱",
	ContactTypeWeChat: "微信",
	ContactTypeOther:  "其他",
}

func GetContactTypeLabel(contactType ContactType) string {
	return contactTypeLabelMap[contactType]
}
