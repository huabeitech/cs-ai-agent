package executor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func resolveCheckPointID(input string, runID string) string {
	checkPointID := strings.TrimSpace(input)
	if checkPointID != "" {
		return checkPointID
	}
	return "eino_cp_" + strings.TrimSpace(runID)
}

func buildResumeDataMessage(resumeData map[string]any) *schema.Message {
	if len(resumeData) == 0 {
		return nil
	}
	data, err := json.Marshal(resumeData)
	if err != nil {
		return schema.UserMessage(fmt.Sprint(resumeData))
	}
	return schema.UserMessage(string(data))
}

func buildRunOptions(checkPointID string) []adk.AgentRunOption {
	options := make([]adk.AgentRunOption, 0, 1)
	if strings.TrimSpace(checkPointID) != "" {
		options = append(options, adk.WithCheckPointID(checkPointID))
	}
	return options
}

func buildResumeOptions(checkPointID string, resumeData *schema.Message) []adk.AgentRunOption {
	options := make([]adk.AgentRunOption, 0, 1)
	if strings.TrimSpace(checkPointID) != "" {
		options = append(options, adk.WithCheckPointID(checkPointID))
	}
	_ = resumeData
	return options
}

func previewInterruptInfo(info any) string {
	if info == nil {
		return ""
	}
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Sprint(info)
	}
	return string(data)
}
