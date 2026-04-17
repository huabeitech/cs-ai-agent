package tools

import (
	"context"

	"cs-agent/internal/ai/runtime/graphs"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	einojsonschema "github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type AnalyzeConversationTool struct {
	conversation models.Conversation
}

func NewAnalyzeConversationTool() *AnalyzeConversationTool {
	return &AnalyzeConversationTool{}
}

func (t *AnalyzeConversationTool) Spec() toolx.ToolSpec {
	return toolx.GraphAnalyzeConversation
}

func (t *AnalyzeConversationTool) Name() string {
	return toolx.GraphAnalyzeConversation.Name
}

func (t *AnalyzeConversationTool) Code() string {
	return toolx.GraphAnalyzeConversation.Code
}

func (t *AnalyzeConversationTool) Enabled(ctx registry.Context) bool {
	return true
}

func (t *AnalyzeConversationTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &AnalyzeConversationTool{conversation: ctx.Conversation}, nil
}

func (t *AnalyzeConversationTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: toolx.GraphAnalyzeConversation.Name,
		Desc: "Graph Tool。用于整理当前对话摘要、识别投诉/资金/情绪等风险信号，并给出继续解答、建单或转人工的建议。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "goal",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "当前分析目标，例如判断是否需要转人工、是否需要建单、是否需要做风险质检。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "observedIssue",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "当前观察到的主要问题或诉求。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "needTicket",
					Value: &einojsonschema.Schema{
						Type:        "boolean",
						Description: "是否重点评估建单必要性。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "needHumanHandoff",
					Value: &einojsonschema.Schema{
						Type:        "boolean",
						Description: "是否重点评估转人工必要性。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "needQualityCheck",
					Value: &einojsonschema.Schema{
						Type:        "boolean",
						Description: "是否重点做风险/质检分析。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "additionalContext",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "补充上下文，例如你已发现的争议点、投诉点或业务限制。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode":   toolx.GraphAnalyzeConversation.Code,
			"sourceType": toolx.GraphAnalyzeConversation.SourceType,
		},
	}, nil
}

func (t *AnalyzeConversationTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	return graphs.NewAnalyzeConversationGraph(t.conversation).Run(ctx, argumentsInJSON)
}
