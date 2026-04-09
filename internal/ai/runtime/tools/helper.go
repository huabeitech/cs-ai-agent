package tools

import (
	"fmt"
	"strings"
)

type Decision string

const (
	DecisionConfirm Decision = "confirm"
	DecisionCancel  Decision = "cancel"
)

func ParseConfirmationDecision(value string) Decision {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	confirmWords := []string{"确认", "是", "好的", "可以", "ok", "yes", "继续", "同意"}
	for _, item := range confirmWords {
		if strings.Contains(value, item) {
			return DecisionConfirm
		}
	}
	cancelWords := []string{"取消", "不用", "不需要", "算了", "no"}
	for _, item := range cancelWords {
		if strings.Contains(value, item) {
			return DecisionCancel
		}
	}
	return ""
}

func getStringValue(data map[string]any, key string) string {
	value, ok := data[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getInt64Value(data map[string]any, key string) int64 {
	value, ok := data[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
