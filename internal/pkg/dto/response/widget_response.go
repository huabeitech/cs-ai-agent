package response

type WidgetConfigResponse struct {
	ChannelID  string `json:"channelId"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ThemeColor string `json:"themeColor"`
	Position   string `json:"position"`
	Width      string `json:"width"`
}
