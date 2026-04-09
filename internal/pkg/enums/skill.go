package enums

type SkillExecutionMode string

const (
	SkillExecutionModePromptOnly SkillExecutionMode = "prompt_only"
	SkillExecutionModeMCPTool    SkillExecutionMode = "mcp_tool"
)

var skillExecutionModeLabelMap = map[SkillExecutionMode]string{
	SkillExecutionModePromptOnly: "Prompt直出",
	SkillExecutionModeMCPTool:    "MCP工具",
}

func GetSkillExecutionModeLabel(mode SkillExecutionMode) string {
	return skillExecutionModeLabelMap[mode]
}
