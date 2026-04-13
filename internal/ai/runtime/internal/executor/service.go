package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/factory"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
	"cs-agent/internal/pkg/utils"

	"github.com/cloudwego/eino/adk"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

type Service struct {
	agentFactory  *factory.AgentFactory
	runnerFactory *factory.RunnerFactory
}

func NewService() *Service {
	return &Service{
		agentFactory:  factory.NewAgentFactory(),
		runnerFactory: factory.NewRunnerFactory(),
	}
}

func (s *Service) ExecuteRun(ctx context.Context, req RunInput) (*RunResult, error) {
	summary := &RunResult{
		RunID:            uuid.NewString(),
		Status:           "started",
		ToolCodes:        make([]string, 0),
		InvokedToolCodes: make([]string, 0),
	}
	collector := callbacks.NewRuntimeTraceCollector()
	collector.Data.RunID = summary.RunID
	if req.AIAgent == nil || req.Conversation == nil || req.UserMessage == nil {
		summary.Status = "error"
		summary.ErrorMessage = "invalid runtime request"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	if req.AIConfig == nil {
		summary.Status = "error"
		summary.ErrorMessage = "ai config is nil"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}

	history := adapter.BuildHistoryMessages(req.Conversation.ID, req.UserMessage.ID, 12)
	summary.HistoryMessageCount = len(history.Messages)
	collector.Data.Input.HistoryMessageCount = len(history.Messages)
	collector.Data.Input.KnowledgeBaseIDs = utils.SplitInt64s(req.AIAgent.KnowledgeIDs)
	collector.Data.Input.CurrentUserMessagePreview = preview(req.UserMessage.Content, 120)

	toolDefs, err := factory.NewToolFactory().BuildMCPTools(req.AIAgent)
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "prepare"
		summary.TraceData = collector.Marshal()
		return summary, err
	}
	filteredToolDefs := filterToolDefinitionsBySkill(toolDefs, req.SelectedSkill)
	toolDefsByModelName := make(map[string]string, len(filteredToolDefs))
	for _, item := range filteredToolDefs {
		summary.ToolCodes = append(summary.ToolCodes, item.ToolCode)
		toolDefsByModelName[item.ModelName] = item.ToolCode
	}
	if len(filteredToolDefs) > 0 {
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, toolx.BuiltinToolSearch.Code)
		toolDefsByModelName[toolx.BuiltinToolSearch.Name] = toolx.BuiltinToolSearch.Code
	}
	if req.SelectedSkill != nil {
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, toolx.BuiltinSkill.Code)
		toolDefsByModelName[toolx.BuiltinSkill.Name] = toolx.BuiltinSkill.Code
	}
	for modelName, toolCode := range toolSetStaticToolCodes(req.ToolSet) {
		toolCode = strings.TrimSpace(toolCode)
		modelName = strings.TrimSpace(modelName)
		if toolCode == "" || modelName == "" {
			continue
		}
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, toolCode)
		toolDefsByModelName[modelName] = toolCode
	}
	collector.Data.Input.ToolCodes = append(collector.Data.Input.ToolCodes, summary.ToolCodes...)
	collector.SetTooling(staticToolCodeList(req.ToolSet), definitionToolCodes(filteredToolDefs), len(filteredToolDefs) > 0)

	collector.Data.Model.Provider = string(req.AIConfig.Provider)
	collector.Data.Model.Name = req.AIConfig.ModelName
	summary.SelectedSkillCode = ""
	summary.SelectedSkillName = ""
	summary.SkillRouteReason = strings.TrimSpace(req.SkillRouteReason)
	summary.SkillRouteTrace = strings.TrimSpace(req.SkillRouteTrace)
	if req.SelectedSkill != nil {
		summary.SelectedSkillCode = strings.TrimSpace(req.SelectedSkill.Code)
		summary.SelectedSkillName = strings.TrimSpace(req.SelectedSkill.Name)
		summary.SkillAllowedToolCodes = parseJSONArrayList(req.SelectedSkill.ToolWhitelist)
		collector.Data.Skill.Code = summary.SelectedSkillCode
		collector.Data.Skill.Name = summary.SelectedSkillName
		collector.Data.Skill.AllowedToolCodes = append([]string(nil), summary.SkillAllowedToolCodes...)
	}
	collector.Data.Skill.RouteReason = summary.SkillRouteReason
	collector.Data.Skill.RouteTrace = summary.SkillRouteTrace

	agent, err := s.agentFactory.BuildCustomerServiceAgent(ctx, factory.BuildCustomerServiceAgentInput{
		AIAgent:                    req.AIAgent,
		AIConfig:                   req.AIConfig,
		SelectedSkill:              req.SelectedSkill,
		InstructionToolDefinitions: filteredToolDefs,
		DynamicMCPToolDefinitions:  filteredToolDefs,
		StaticTools:                toolSetStaticTools(req.ToolSet),
		StaticToolCodes:            toolSetStaticToolCodes(req.ToolSet),
		StaticToolMetadata:         toolSetStaticToolMetadata(req.ToolSet),
		Collector:                  collector,
	})
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "prepare"
		summary.TraceData = collector.Marshal()
		return summary, err
	}

	checkPointID := strings.TrimSpace(req.CheckPointID)
	if checkPointID == "" {
		checkPointID = "eino_cp_" + summary.RunID
	}
	summary.CheckPointID = checkPointID
	runner := s.runnerFactory.Build(ctx, agent, false, true)
	if runner == nil {
		summary.Status = "error"
		summary.ErrorMessage = "failed to build runner"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	messages := make([]*schema.Message, 0, len(history.Messages)+3)
	messages = append(messages, history.Messages...)

	retriever := retrievers.NewKnowledgeRetriever(req.AIAgent)
	retrieveOptions := retrievers.DefaultKnowledgeRetrieveOptions()
	retrieveOptions.QueryPreview = preview(req.UserMessage.Content, 120)
	if retrieveResult, retrieveErr := retriever.RetrieveContextByOptions(ctx, retrieveOptions, strings.TrimSpace(req.UserMessage.Content)); retrieveErr == nil && retrieveResult != nil {
		summary.RetrieverCount = len(retrieveResult.Hits)
		collector.SetRetrieverSummary(retrieveResult.TraceSummary)
		collector.Data.Retriever.Items = append(collector.Data.Retriever.Items, retrieveResult.TraceItems...)
		if strings.TrimSpace(retrieveResult.ContextText) != "" {
			messages = append(messages, schema.SystemMessage(retrieveResult.ContextText))
		}
	}

	messages = append(messages, schema.UserMessage(strings.TrimSpace(req.UserMessage.Content)))
	collector.Data.Interrupt.CheckPointID = checkPointID
	consumeAgentEvents(runner.Run(ctx, messages, buildRunOptions(checkPointID)...), summary, collector, toolDefsByModelName)
	summary.ModelName = req.AIConfig.ModelName
	collector.Data.Status = summary.Status
	collector.Data.Output.ReplyText = summary.ReplyText
	collector.Data.Output.FinishReason = summary.Status
	summary.TraceData = collector.Marshal()
	return summary, nil
}

func (s *Service) ExecuteResume(ctx context.Context, req ResumeInput) (*RunResult, error) {
	summary := &RunResult{
		RunID:            uuid.NewString(),
		Status:           "started",
		CheckPointID:     strings.TrimSpace(req.CheckPointID),
		ToolCodes:        make([]string, 0),
		InvokedToolCodes: make([]string, 0),
		Interrupts:       make([]InterruptContextSummary, 0),
	}
	collector := callbacks.NewRuntimeTraceCollector()
	collector.Data.RunID = summary.RunID
	collector.Data.Interrupt.CheckPointID = summary.CheckPointID
	if req.AIAgent == nil {
		summary.Status = "error"
		summary.ErrorMessage = "ai agent is nil"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	if req.AIConfig == nil {
		summary.Status = "error"
		summary.ErrorMessage = "ai config is nil"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	if summary.CheckPointID == "" {
		summary.Status = "error"
		summary.ErrorMessage = "checkpoint id is required"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	toolDefs, err := factory.NewToolFactory().BuildMCPTools(req.AIAgent)
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, err
	}
	toolDefsByModelName := make(map[string]string, len(toolDefs))
	for _, item := range toolDefs {
		summary.ToolCodes = append(summary.ToolCodes, item.ToolCode)
		toolDefsByModelName[item.ModelName] = item.ToolCode
	}
	if len(toolDefs) > 0 {
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, toolx.BuiltinToolSearch.Code)
		toolDefsByModelName[toolx.BuiltinToolSearch.Name] = toolx.BuiltinToolSearch.Code
	}
	for modelName, toolCode := range toolSetStaticToolCodes(req.ToolSet) {
		toolCode = strings.TrimSpace(toolCode)
		modelName = strings.TrimSpace(modelName)
		if toolCode == "" || modelName == "" {
			continue
		}
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, toolCode)
		toolDefsByModelName[modelName] = toolCode
	}
	collector.Data.Input.ToolCodes = append(collector.Data.Input.ToolCodes, summary.ToolCodes...)
	collector.SetTooling(staticToolCodeList(req.ToolSet), definitionToolCodes(toolDefs), len(toolDefs) > 0)
	collector.Data.Model.Provider = string(req.AIConfig.Provider)
	collector.Data.Model.Name = req.AIConfig.ModelName

	agent, err := s.agentFactory.BuildCustomerServiceAgent(ctx, factory.BuildCustomerServiceAgentInput{
		AIAgent:                    req.AIAgent,
		AIConfig:                   req.AIConfig,
		InstructionToolDefinitions: toolDefs,
		DynamicMCPToolDefinitions:  toolDefs,
		StaticTools:                toolSetStaticTools(req.ToolSet),
		StaticToolCodes:            toolSetStaticToolCodes(req.ToolSet),
		StaticToolMetadata:         toolSetStaticToolMetadata(req.ToolSet),
		Collector:                  collector,
	})
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, err
	}
	runner := s.runnerFactory.Build(ctx, agent, false, true)
	if runner == nil {
		summary.Status = "error"
		summary.ErrorMessage = "failed to build runner"
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = summary.ErrorMessage
		collector.Data.Error.Stage = "resume_prepare"
		summary.TraceData = collector.Marshal()
		return summary, fmt.Errorf("%s", summary.ErrorMessage)
	}
	resumeData := buildResumeDataMessage(req.ResumeData)
	iter, err := runner.Resume(ctx, summary.CheckPointID, buildResumeOptions(summary.CheckPointID, resumeData)...)
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "resume_execute"
		summary.TraceData = collector.Marshal()
		return summary, err
	}
	consumeAgentEvents(iter, summary, collector, toolDefsByModelName)
	summary.ModelName = req.AIConfig.ModelName
	collector.Data.Status = summary.Status
	collector.Data.Output.ReplyText = summary.ReplyText
	collector.Data.Output.FinishReason = summary.Status
	summary.TraceData = collector.Marshal()
	return summary, nil
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

func consumeAgentEvents(events *adk.AsyncIterator[*adk.AgentEvent], summary *RunResult, collector *callbacks.RuntimeTraceCollector, toolDefsByModelName map[string]string) {
	if summary == nil {
		return
	}
	if collector == nil {
		collector = callbacks.NewRuntimeTraceCollector()
	}
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			summary.Status = "interrupted"
			summary.Interrupted = true
			summary.Interrupts = buildInterruptSummaries(event)
		}
		if event.Err != nil {
			errMsg := strings.TrimSpace(event.Err.Error())
			if errMsg != "" {
				summary.Status = "error"
				summary.ErrorMessage = errMsg
			}
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		messageOutput := event.Output.MessageOutput
		switch messageOutput.Role {
		case schema.Assistant:
			replyText := strings.TrimSpace(messageOutput.Message.Content)
			if replyText != "" {
				summary.ReplyText = replyText
			}
		case schema.Tool:
			toolName := strings.TrimSpace(messageOutput.ToolName)
			if toolName == "" {
				continue
			}
			toolCode := toolName
			if mappedCode, ok := toolDefsByModelName[toolName]; ok && strings.TrimSpace(mappedCode) != "" {
				toolCode = strings.TrimSpace(mappedCode)
			}
			summary.InvokedToolCodes = appendIfMissing(summary.InvokedToolCodes, toolCode)
		}
	}
	if summary.Status == "started" {
		switch {
		case strings.TrimSpace(summary.ErrorMessage) != "":
			summary.Status = "error"
		case summary.Interrupted:
			summary.Status = "interrupted"
		case strings.TrimSpace(summary.ReplyText) != "":
			summary.Status = "completed"
		default:
			summary.Status = "fallback"
		}
	}
	summary.ToolCallCount = len(summary.InvokedToolCodes)
}

func buildInterruptSummaries(event *adk.AgentEvent) []InterruptContextSummary {
	if event == nil || event.Action == nil || event.Action.Interrupted == nil {
		return nil
	}
	interrupts := event.Action.Interrupted.InterruptContexts
	result := make([]InterruptContextSummary, 0, len(interrupts))
	for _, item := range interrupts {
		if item == nil {
			continue
		}
		result = append(result, InterruptContextSummary{
			ID:          strings.TrimSpace(item.ID),
			InfoPreview: previewInterruptInfo(item.Info),
		})
	}
	return result
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

func appendIfMissing(items []string, item string) []string {
	item = strings.TrimSpace(item)
	if item == "" {
		return items
	}
	for _, existing := range items {
		if strings.TrimSpace(existing) == item {
			return items
		}
	}
	return append(items, item)
}

func staticToolCodeList(toolSet *registry.ToolSet) []string {
	if toolSet == nil {
		return nil
	}
	metadata := toolSetStaticToolMetadata(toolSet)
	ret := make([]string, 0, len(metadata))
	for _, item := range metadata {
		code := strings.TrimSpace(item.ToolCode)
		if code == "" {
			continue
		}
		ret = appendIfMissing(ret, code)
	}
	if len(ret) > 0 {
		return ret
	}
	for _, code := range toolSetStaticToolCodes(toolSet) {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		ret = appendIfMissing(ret, code)
	}
	return ret
}

func toolSetStaticTools(toolSet *registry.ToolSet) []einotool.BaseTool {
	if toolSet == nil {
		return nil
	}
	return append([]einotool.BaseTool(nil), toolSet.StaticTools...)
}

func toolSetStaticToolCodes(toolSet *registry.ToolSet) map[string]string {
	if toolSet == nil || len(toolSet.StaticToolCodes) == 0 {
		return nil
	}
	ret := make(map[string]string, len(toolSet.StaticToolCodes))
	for name, code := range toolSet.StaticToolCodes {
		ret[strings.TrimSpace(name)] = strings.TrimSpace(code)
	}
	return ret
}

func toolSetStaticToolMetadata(toolSet *registry.ToolSet) map[string]registry.ToolMetadata {
	if toolSet == nil || len(toolSet.StaticToolMetadata) == 0 {
		return nil
	}
	ret := make(map[string]registry.ToolMetadata, len(toolSet.StaticToolMetadata))
	for name, item := range toolSet.StaticToolMetadata {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		item.ToolCode = strings.TrimSpace(item.ToolCode)
		item.ServerCode = strings.TrimSpace(item.ServerCode)
		item.ToolName = strings.TrimSpace(item.ToolName)
		item.SourceType = strings.TrimSpace(item.SourceType)
		ret[trimmedName] = item
	}
	return ret
}

func definitionToolCodes(defs []adapter.MCPToolDefinition) []string {
	ret := make([]string, 0, len(defs))
	for _, item := range defs {
		code := strings.TrimSpace(item.ToolCode)
		if code == "" {
			continue
		}
		ret = append(ret, code)
	}
	return ret
}

func filterToolDefinitionsBySkill(defs []adapter.MCPToolDefinition, skill *models.SkillDefinition) []adapter.MCPToolDefinition {
	if skill == nil || strings.TrimSpace(skill.ToolWhitelist) == "" {
		return defs
	}
	var allowed []string
	if err := json.Unmarshal([]byte(skill.ToolWhitelist), &allowed); err != nil {
		return defs
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		item = toolx.NormalizeToolCodeAlias(item)
		if strings.TrimSpace(item) == "" {
			continue
		}
		allowedSet[strings.TrimSpace(item)] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return defs
	}
	ret := make([]adapter.MCPToolDefinition, 0, len(defs))
	for _, item := range defs {
		if _, ok := allowedSet[strings.TrimSpace(item.ToolCode)]; ok {
			ret = append(ret, item)
		}
	}
	return ret
}

func parseJSONArrayList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func preview(text string, limit int) string {
	text = strings.TrimSpace(text)
	if text == "" || limit <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "..."
}
