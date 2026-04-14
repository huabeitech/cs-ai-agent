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

func buildResumeDataMessage(resumeData map[string]string) *schema.Message {
	if len(resumeData) == 0 {
		return nil
	}
	data, err := json.Marshal(resumeData)
	if err != nil {
		return schema.UserMessage(fmt.Sprint(resumeData))
	}
	return schema.UserMessage(string(data))
}

func buildResumeTargets(resumeData map[string]string) map[string]any {
	if len(resumeData) == 0 {
		return nil
	}
	targets := make(map[string]any, len(resumeData))
	for key, value := range resumeData {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		targets[key] = value
	}
	if len(targets) == 0 {
		return nil
	}
	return targets
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
