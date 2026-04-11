package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/factory"
	"cs-agent/internal/ai/runtime/internal/impl/retrievers"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
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

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	summary := &Summary{
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
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, "builtin/tool_search")
		toolDefsByModelName["tool_search"] = "builtin/tool_search"
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
		summary.SkillAllowedToolCodes = parseJSONArrayList(req.SelectedSkill.AllowedToolCodes)
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
	if results, _, retrieveErr := retriever.Retrieve(ctx, strings.TrimSpace(req.UserMessage.Content)); retrieveErr == nil {
		summary.RetrieverCount = len(results)
		collector.Data.Retriever.Count = len(results)
		for _, item := range results {
			collector.Data.Retriever.Items = append(collector.Data.Retriever.Items, callbacks.RetrieverTraceItem{
				Query:           preview(req.UserMessage.Content, 120),
				KnowledgeBaseID: item.KnowledgeBaseID,
				DocumentID:      item.DocumentID,
				DocumentTitle:   item.DocumentTitle,
				Score:           float64(item.Score),
			})
		}
		if knowledgeContext := buildKnowledgeContext(results); knowledgeContext != "" {
			messages = append(messages, schema.SystemMessage(knowledgeContext))
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

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	summary := &Summary{
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
		summary.ToolCodes = appendIfMissing(summary.ToolCodes, "builtin/tool_search")
		toolDefsByModelName["tool_search"] = "builtin/tool_search"
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
	var iter *adk.AsyncIterator[*adk.AgentEvent]
	if len(req.ResumeData) > 0 {
		iter, err = runner.ResumeWithParams(ctx, summary.CheckPointID, &adk.ResumeParams{Targets: req.ResumeData})
	} else {
		iter, err = runner.Resume(ctx, summary.CheckPointID)
	}
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "resume_prepare"
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

func filterToolDefinitionsBySkill(definitions []adapter.MCPToolDefinition, skill *models.SkillDefinition) []adapter.MCPToolDefinition {
	if len(definitions) == 0 || skill == nil {
		return definitions
	}
	allowed := parseJSONArraySet(skill.AllowedToolCodes)
	if len(allowed) == 0 {
		return definitions
	}
	ret := make([]adapter.MCPToolDefinition, 0, len(definitions))
	for _, item := range definitions {
		if _, ok := allowed[strings.TrimSpace(item.ToolCode)]; ok {
			ret = append(ret, item)
		}
	}
	return ret
}

func parseJSONArraySet(raw string) map[string]struct{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	items := parseJSONArrayList(raw)
	if len(items) == 0 {
		return nil
	}
	ret := make(map[string]struct{}, len(items))
	for _, item := range items {
		ret[item] = struct{}{}
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

func buildRunOptions(checkPointID string) []adk.AgentRunOption {
	if strings.TrimSpace(checkPointID) == "" {
		return nil
	}
	return []adk.AgentRunOption{adk.WithCheckPointID(checkPointID)}
}

func consumeAgentEvents(iter *adk.AsyncIterator[*adk.AgentEvent], summary *Summary, collector *callbacks.RuntimeTraceCollector, toolDefsByModelName map[string]string) {
	if iter == nil || summary == nil || collector == nil {
		return
	}
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Err != nil {
			summary.Status = "error"
			summary.ErrorMessage = event.Err.Error()
			collector.Data.Error.Message = event.Err.Error()
			collector.Data.Error.Stage = "model"
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			summary.Status = "interrupted"
			summary.Interrupted = true
			summary.Interrupts = summarizeInterrupts(event.Action.Interrupted.InterruptContexts)
			collector.Data.Interrupt.Items = convertInterruptTraceItems(summary.Interrupts)
			continue
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		message, getErr := event.Output.MessageOutput.GetMessage()
		if getErr != nil || message == nil {
			continue
		}
		switch event.Output.MessageOutput.Role {
		case schema.Assistant:
			summary.ReplyText = strings.TrimSpace(message.Content)
		case schema.Tool:
			summary.ToolCallCount++
			if toolDefsByModelName != nil {
				toolCode := strings.TrimSpace(toolDefsByModelName[message.ToolName])
				if toolCode != "" {
					summary.InvokedToolCodes = appendIfMissing(summary.InvokedToolCodes, toolCode)
				}
			}
		}
	}
	if summary.Status == "started" {
		if strings.TrimSpace(summary.ReplyText) == "" {
			summary.Status = "fallback"
		} else {
			summary.Status = "completed"
		}
	}
}

func convertInterruptTraceItems(items []InterruptContextSummary) []callbacks.InterruptTraceContext {
	if len(items) == 0 {
		return nil
	}
	ret := make([]callbacks.InterruptTraceContext, 0, len(items))
	for _, item := range items {
		ret = append(ret, callbacks.InterruptTraceContext{
			Type:        item.Type,
			ID:          item.ID,
			InfoPreview: item.InfoPreview,
		})
	}
	return ret
}

func previewInterruptInfo(info any) string {
	if info == nil {
		return ""
	}
	switch v := info.(type) {
	case string:
		return preview(v, 200)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return preview(string(data), 200)
	}
}

func summarizeInterrupts(items []*adk.InterruptCtx) []InterruptContextSummary {
	if len(items) == 0 {
		return nil
	}
	ret := make([]InterruptContextSummary, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ret = append(ret, InterruptContextSummary{
			Type:        extractInterruptType(item.Info),
			ID:          strings.TrimSpace(item.ID),
			InfoPreview: previewInterruptInfo(item.Info),
		})
	}
	return ret
}

func extractInterruptType(info any) string {
	if info == nil {
		return ""
	}
	switch v := info.(type) {
	case map[string]any:
		return strings.TrimSpace(getStringFromAnyMap(v, "type"))
	default:
		return ""
	}
}

func getStringFromAnyMap(data map[string]any, key string) string {
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

func appendIfMissing(items []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return items
	}
	for _, item := range items {
		if strings.TrimSpace(item) == value {
			return items
		}
	}
	return append(items, value)
}

func preview(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func buildKnowledgeContext(items []rag.RetrieveResult) string {
	if len(items) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("以下是可供参考的知识库内容，请优先基于这些内容回答；如果仍不确定，请明确说明并向用户澄清。\n\n")
	for i, item := range items {
		if i >= 5 {
			break
		}
		builder.WriteString("[知识片段")
		builder.WriteString(fmt.Sprintf("%d", i+1))
		builder.WriteString("]\n")
		if strings.TrimSpace(item.DocumentTitle) != "" {
			builder.WriteString("标题: ")
			builder.WriteString(strings.TrimSpace(item.DocumentTitle))
			builder.WriteString("\n")
		}
		if strings.TrimSpace(item.Content) != "" {
			builder.WriteString("内容: ")
			builder.WriteString(strings.TrimSpace(item.Content))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func toolSetStaticTools(toolSet *registry.ToolSet) []einotool.BaseTool {
	if toolSet == nil {
		return nil
	}
	return toolSet.StaticTools
}

func toolSetStaticToolCodes(toolSet *registry.ToolSet) map[string]string {
	if toolSet == nil {
		return nil
	}
	return toolSet.StaticToolCodes
}

func definitionToolCodes(definitions []adapter.MCPToolDefinition) []string {
	if len(definitions) == 0 {
		return nil
	}
	ret := make([]string, 0, len(definitions))
	for _, item := range definitions {
		toolCode := strings.TrimSpace(item.ToolCode)
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}

func staticToolCodeList(toolSet *registry.ToolSet) []string {
	toolCodes := toolSetStaticToolCodes(toolSet)
	if len(toolCodes) == 0 {
		return nil
	}
	ret := make([]string, 0, len(toolCodes))
	for _, toolCode := range toolCodes {
		toolCode = strings.TrimSpace(toolCode)
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}
