package instruction

import (
	"strings"

	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

const defaultGovernanceInstruction = `
你正在一个有明确工程约束的客服系统中工作。
执行时必须严格遵守当前注入的 Agent 规则和技能规则。
如果存在工具白名单限制，只能调用当前允许的工具；信息不足时优先追问，不要伪造事实或跳过必要确认。
禁止承诺未经系统确认的处理时效、完成时间、回访时间或联系时间。
禁止代表人工团队、技术团队、售后团队承诺后续动作，除非当前上下文已有明确的工具结果、人工确认或知识库事实支持。
当用户只表示已发送资料、邮件、截图或附件时，只能确认已收到当前消息或建议等待人工确认，不能自行补充内部处理流程、SLA 或跟进安排。
`

type ToolAppendixProvider struct{}

func NewToolAppendixProvider() *ToolAppendixProvider {
	return &ToolAppendixProvider{}
}

type GovernanceInstructionProvider struct{}

func NewGovernanceInstructionProvider() *GovernanceInstructionProvider {
	return &GovernanceInstructionProvider{}
}

func (p *GovernanceInstructionProvider) Resolve() string {
	return strings.TrimSpace(defaultGovernanceInstruction)
}

type SkillInstructionProvider struct{}

func NewSkillInstructionProvider() *SkillInstructionProvider {
	return &SkillInstructionProvider{}
}

func (p *SkillInstructionProvider) Resolve(selectedSkill *models.SkillDefinition) string {
	return BuildSelectedSkillActivationInstruction(selectedSkill)
}

func (p *ToolAppendixProvider) Build(toolDefinitions []runtimetooling.MCPToolDefinition, extraToolCodes map[string]string) []string {
	appendixParts := make([]string, 0, 1)
	toolCodes := make([]string, 0, len(toolDefinitions)+len(extraToolCodes))
	for _, item := range toolDefinitions {
		toolCodes = append(toolCodes, item.ToolCode)
	}
	for _, item := range extraToolCodes {
		toolCodes = append(toolCodes, item)
	}
	appendixParts = append(appendixParts, toolx.BuildToolAppendicesForCodes(len(toolDefinitions) > 0, toolCodes)...)
	return appendixParts
}
