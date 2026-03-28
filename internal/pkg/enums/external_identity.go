package enums

// ExternalSource 外部身份来源。
//
// 会话侧 Conversation.ChannelType 与 CRM 侧 CustomerIdentity.SourceType 共用同一套取值，
// 与 ExternalUserID / SourceID 组合即可唯一标识某渠道下的访客身份。
type ExternalSource string

const (
	ExternalSourceWebChat ExternalSource = "web_chat"
)

var externalSourceLabelMap = map[ExternalSource]string{
	ExternalSourceWebChat: "网页客服",
}

func GetExternalSourceLabel(v ExternalSource) string {
	if s, ok := externalSourceLabelMap[v]; ok {
		return s
	}
	return string(v)
}
