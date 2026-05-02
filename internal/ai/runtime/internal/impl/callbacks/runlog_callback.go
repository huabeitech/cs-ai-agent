package callbacks

import (
	"encoding/json"
	"sync"
)

type RuntimeTraceCollector struct {
	mu   sync.Mutex
	Data RuntimeTraceData
}

func NewRuntimeTraceCollector() *RuntimeTraceCollector {
	ret := &RuntimeTraceCollector{}
	ret.Data.Version = "v1"
	ret.Data.Status = "started"
	return ret
}

func (c *RuntimeTraceCollector) Marshal() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	buf, err := json.Marshal(c.Data)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (c *RuntimeTraceCollector) SetTooling(staticToolCodes []string, dynamicToolCodes []string, toolSearchEnabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Input.StaticToolCodes = append([]string(nil), staticToolCodes...)
	c.Data.Input.DynamicToolCodes = append([]string(nil), dynamicToolCodes...)
	c.Data.Input.ToolSearchEnabled = toolSearchEnabled
}

func (c *RuntimeTraceCollector) SetInstructionSummary(summary InstructionTraceSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Instruction.SectionTitles = append([]string(nil), summary.SectionTitles...)
	c.Data.Instruction.HasAgentRule = summary.HasAgentRule
	c.Data.Instruction.HasSkillRule = summary.HasSkillRule
	c.Data.Instruction.HasToolRule = summary.HasToolRule
}

func (c *RuntimeTraceCollector) SetSkillMiddleware(enabled bool, toolName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.MiddlewareEnabled = enabled
	c.Data.Skill.MiddlewareToolName = toolName
}

type SkillMetadata struct {
	Code             string
	Name             string
	Description      string
	AllowedToolCodes []string
}

func (c *RuntimeTraceCollector) SetVisibleSkills(skills map[string]SkillMetadata) {
	if len(skills) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	codes := make([]string, 0, len(skills))
	for code := range skills {
		if code == "" {
			continue
		}
		codes = append(codes, code)
	}
	c.Data.Skill.VisibleCodes = append([]string(nil), codes...)
}

func (c *RuntimeTraceCollector) ActivateSkill(skill SkillMetadata, routeReason string, routeTrace string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.Code = skill.Code
	c.Data.Skill.Name = skill.Name
	c.Data.Skill.Description = skill.Description
	c.Data.Skill.AllowedToolCodes = append([]string(nil), skill.AllowedToolCodes...)
	c.Data.Skill.RouteReason = routeReason
	c.Data.Skill.RouteTrace = routeTrace
}

func (c *RuntimeTraceCollector) SetFilteredToolCodes(toolCodes []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Skill.FilteredToolCodes = append([]string(nil), toolCodes...)
}

func (c *RuntimeTraceCollector) SetRetrieverSummary(summary RetrieverTraceSummary) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Retriever.TopK = summary.TopK
	c.Data.Retriever.ScoreThreshold = summary.ScoreThreshold
	c.Data.Retriever.ContextMaxTokens = summary.ContextMaxTokens
	c.Data.Retriever.MaxContextItems = summary.MaxContextItems
	c.Data.Retriever.Count = summary.HitCount
	c.Data.Retriever.ContextCount = summary.ContextCount
	c.Data.Retriever.EmbeddingMs = summary.EmbeddingMs
	c.Data.Retriever.VectorSearchMs = summary.VectorSearchMs
	c.Data.Retriever.HydrateMs = summary.HydrateMs
	c.Data.Retriever.Policies = append([]RetrieverPolicyTraceItem(nil), summary.Policies...)
}

func (c *RuntimeTraceCollector) AddRetrieverItems(items []RetrieverTraceItem) {
	if len(items) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Retriever.Items = append(c.Data.Retriever.Items, items...)
}

func (c *RuntimeTraceCollector) SetAnswerability(data AnswerabilityTraceData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Answerability = AnswerabilityTraceData{
		Status:             data.Status,
		Reason:             data.Reason,
		SupportingChunkIDs: append([]string(nil), data.SupportingChunkIDs...),
		MissingInfo:        append([]string(nil), data.MissingInfo...),
		LatencyMs:          data.LatencyMs,
		ErrorMessage:       data.ErrorMessage,
	}
}

func (c *RuntimeTraceCollector) AddToolItem(item ToolTraceItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.Tools.Count++
	c.Data.Tools.Items = append(c.Data.Tools.Items, item)
}

func (c *RuntimeTraceCollector) AddToolSearchItem(item ToolSearchTraceItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.ToolSearch.Count++
	c.Data.ToolSearch.Items = append(c.Data.ToolSearch.Items, item)
}

func (c *RuntimeTraceCollector) AddGraphToolItem(item GraphToolTraceItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data.GraphTools.Count++
	c.Data.GraphTools.Items = append(c.Data.GraphTools.Items, item)
}
