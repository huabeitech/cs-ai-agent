package enums

type AssetStatus int

const (
	AssetStatusPending AssetStatus = 1
	AssetStatusSuccess AssetStatus = 2
	AssetStatusFailed  AssetStatus = 3
	AssetStatusDeleted AssetStatus = 4
)

var assetStatusLabelMap = map[AssetStatus]string{
	AssetStatusPending: "处理中",
	AssetStatusSuccess: "成功",
	AssetStatusFailed:  "失败",
	AssetStatusDeleted: "已删除",
}

func GetAssetStatusLabel(status AssetStatus) string {
	return assetStatusLabelMap[status]
}

type AssetProvider string

const (
	AssetProviderLocal AssetProvider = "local"
	AssetProviderOSS   AssetProvider = "oss"
)

var assetProviderLabelMap = map[AssetProvider]string{
	AssetProviderLocal: "本地存储",
	AssetProviderOSS:   "对象存储",
}

func GetAssetProviderLabel(provider AssetProvider) string {
	return assetProviderLabelMap[provider]
}
