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
