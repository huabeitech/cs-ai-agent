package enums

// ExternalSource 外部身份来源。
//
// 与 ExternalID 组合即可唯一标识某渠道下的访客身份。
type ExternalSource string

const (
	ExternalSourceWebChat  ExternalSource = "web_chat"
	ExternalSourceWxWorkKF ExternalSource = "wxwork_kf"
)

var externalSourceLabelMap = map[ExternalSource]string{
	ExternalSourceWebChat:  "网页客服",
	ExternalSourceWxWorkKF: "企业微信客服",
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
	case ExternalSourceWebChat:
		return true
	default:
		return false
	}
}
