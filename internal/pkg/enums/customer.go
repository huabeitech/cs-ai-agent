package enums

const CustomerSourceTypeWebChat = ExternalSourceWebChat

func GetCustomerSourceTypeLabel(sourceType ExternalSource) string {
	return GetExternalSourceLabel(ExternalSource(sourceType))
}

// ExternalType 客户外部身份类型。
type ExternalType string

const (
	ExternalTypeGuest ExternalType = "guest"
	ExternalTypeUser  ExternalType = "user"
)

var externalTypeLabelMap = map[ExternalType]string{
	ExternalTypeGuest: "访客",
	ExternalTypeUser:  "用户",
}

func GetExternalTypeLabel(externalType ExternalType) string {
	if s, ok := externalTypeLabelMap[externalType]; ok {
		return s
	}
	return string(externalType)
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
	ContactTypeOther:  "其他",
}

func GetContactTypeLabel(contactType ContactType) string {
	return contactTypeLabelMap[contactType]
}

// IsValidContactType 判断字符串是否为合法联系方式类型。
func IsValidContactType(v string) bool {
	switch ContactType(v) {
	case ContactTypeMobile, ContactTypeEmail, ContactTypeWeChat, ContactTypeOther:
		return true
	default:
		return false
	}
}
