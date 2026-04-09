package enums

type ServiceStatus int

const (
	ServiceStatusIdle ServiceStatus = 0
	ServiceStatusBusy ServiceStatus = 1
)

var ServiceStatusValues = []ServiceStatus{
	ServiceStatusIdle,
	ServiceStatusBusy,
}

var serviceStatusLabelMap = map[ServiceStatus]string{
	ServiceStatusIdle: "空闲",
	ServiceStatusBusy: "忙碌",
}

func IsValidServiceStatus(status ServiceStatus) bool {
	for _, item := range ServiceStatusValues {
		if item == status {
			return true
		}
	}
	return false
}

func GetServiceStatusLabel(status ServiceStatus) string {
	return serviceStatusLabelMap[status]
}
