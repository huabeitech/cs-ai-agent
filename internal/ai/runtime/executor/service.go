package executor

import (
	"context"
	"fmt"
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/internal/impl/factory"

	"github.com/cloudwego/eino/adk"
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
	tooling := prepareTooling(toolDefs, req.SelectedSkill, req.ToolSet, req.SelectedSkill != nil)
	summary.ToolCodes = append(summary.ToolCodes, tooling.toolCodes...)
	collector.Data.Input.ToolCodes = append(collector.Data.Input.ToolCodes, summary.ToolCodes...)
	collector.SetTooling(tooling.staticToolCodes, definitionToolCodes(tooling.definitions), len(tooling.definitions) > 0)

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
		InstructionToolDefinitions: tooling.definitions,
		DynamicMCPToolDefinitions:  tooling.definitions,
		StaticTools:                tooling.staticTools,
		StaticToolCodes:            tooling.staticToolCodeMap,
		StaticToolMetadata:         tooling.staticToolMetadata,
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

	checkPointID := resolveCheckPointID(req.CheckPointID, summary.RunID)
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
	messages := buildRunMessages(ctx, req, summary, collector)
	if strings.TrimSpace(summary.ReplyText) != "" {
		summary.Status = "completed"
		summary.ModelName = req.AIConfig.ModelName
		collector.Data.Status = summary.Status
		collector.Data.Output.ReplyText = summary.ReplyText
		collector.Data.Output.FinishReason = summary.Status
		summary.TraceData = collector.Marshal()
		return summary, nil
	}
	collector.Data.Interrupt.CheckPointID = checkPointID
	consumeAgentEvents(runner.Run(ctx, messages, buildRunOptions(checkPointID)...), summary, collector, tooling.toolDefsByModelName)
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
	tooling := prepareTooling(toolDefs, nil, req.ToolSet, false)
	summary.ToolCodes = append(summary.ToolCodes, tooling.toolCodes...)
	collector.Data.Input.ToolCodes = append(collector.Data.Input.ToolCodes, summary.ToolCodes...)
	collector.SetTooling(tooling.staticToolCodes, definitionToolCodes(tooling.definitions), len(tooling.definitions) > 0)
	collector.Data.Model.Provider = string(req.AIConfig.Provider)
	collector.Data.Model.Name = req.AIConfig.ModelName

	agent, err := s.agentFactory.BuildCustomerServiceAgent(ctx, factory.BuildCustomerServiceAgentInput{
		AIAgent:                    req.AIAgent,
		AIConfig:                   req.AIConfig,
		InstructionToolDefinitions: tooling.definitions,
		DynamicMCPToolDefinitions:  tooling.definitions,
		StaticTools:                tooling.staticTools,
		StaticToolCodes:            tooling.staticToolCodeMap,
		StaticToolMetadata:         tooling.staticToolMetadata,
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
	resumeTargets := buildResumeTargets(req.ResumeData)
	var (
		iter *adk.AsyncIterator[*adk.AgentEvent]
	)
	if len(resumeTargets) > 0 {
		iter, err = runner.ResumeWithParams(ctx, summary.CheckPointID, &adk.ResumeParams{
			Targets: resumeTargets,
		}, buildResumeOptions(summary.CheckPointID, resumeData)...)
	} else {
		iter, err = runner.Resume(ctx, summary.CheckPointID, buildResumeOptions(summary.CheckPointID, resumeData)...)
	}
	if err != nil {
		summary.Status = "error"
		summary.ErrorMessage = err.Error()
		collector.Data.Status = summary.Status
		collector.Data.Error.Message = err.Error()
		collector.Data.Error.Stage = "resume_execute"
		summary.TraceData = collector.Marshal()
		return summary, err
	}
	consumeAgentEvents(iter, summary, collector, tooling.toolDefsByModelName)
	summary.ModelName = req.AIConfig.ModelName
	collector.Data.Status = summary.Status
	collector.Data.Output.ReplyText = summary.ReplyText
	collector.Data.Output.FinishReason = summary.Status
	summary.TraceData = collector.Marshal()
	return summary, nil
}
