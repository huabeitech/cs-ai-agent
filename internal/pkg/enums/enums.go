package enums

type Status int

const (
	StatusOk       Status = 0
	StatusDisabled Status = 1
	StatusDeleted  Status = 2
)

var StatusValues = []Status{StatusOk, StatusDisabled, StatusDeleted}

var statusLabelMap = map[Status]string{
	StatusOk:       "启用",
	StatusDisabled: "禁用",
	StatusDeleted:  "已删除",
}

func GetStatusLabel(status Status) string {
	return statusLabelMap[status]
}

func IsValidStatus(status int) bool {
	for _, s := range StatusValues {
		if int(s) == status {
			return true
		}
	}
	return false
}

type ThirdProvider string

const (
	ThirdProviderWxWork   ThirdProvider = "wxwork"
	ThirdProviderDingtalk ThirdProvider = "dingtalk"
)

var ThirdProviderValues = []ThirdProvider{ThirdProviderWxWork, ThirdProviderDingtalk}

var thirdProviderLabelMap = map[ThirdProvider]string{
	ThirdProviderWxWork:   "企业微信",
	ThirdProviderDingtalk: "钉钉",
}

func GetThirdProviderLabel(provider ThirdProvider) string {
	return thirdProviderLabelMap[provider]
}

func IsValidThirdProvider(provider string) bool {
	for _, p := range ThirdProviderValues {
		if string(p) == provider {
			return true
		}
	}
	return false
}
