package enums

// ExternalSource 外部身份来源。
//
// 与 ExternalID 组合即可唯一标识某渠道下的访客身份。
type ExternalSource string

const (
	ExternalSourceGuest    ExternalSource = "guest"     // 访客
	ExternalSourceWxWorkKF ExternalSource = "wxwork_kf" // 企业微信客服
	ExternalSourceUser     ExternalSource = "user"      // 用户信息
)

var externalSourceLabelMap = map[ExternalSource]string{
	ExternalSourceGuest:    "访客",
	ExternalSourceWxWorkKF: "企业微信客服",
	ExternalSourceUser:     "用户",
}

func GetExternalSourceLabel(v ExternalSource) string {
	if s, ok := externalSourceLabelMap[v]; ok {
		return s
	}
	return string(v)
}

// IsAllowedOpenImExternalSource 开放 IM 入口允许的外部来源（闭集校验）。
func IsAllowedOpenImExternalSource(s ExternalSource) bool {
	switch s {
	case ExternalSourceGuest, ExternalSourceUser:
		return true
	default:
		return false
	}
}
