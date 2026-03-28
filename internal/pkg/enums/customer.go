package enums

// CustomerSourceType 已合并为 ExternalIdentitySource，保留类型别名以兼容旧代码。
type CustomerSourceType = ExternalSource

const CustomerSourceTypeWebChat = ExternalSourceWebChat

func GetCustomerSourceTypeLabel(sourceType CustomerSourceType) string {
	return GetExternalSourceLabel(ExternalSource(sourceType))
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
