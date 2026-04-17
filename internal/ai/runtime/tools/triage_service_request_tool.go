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

type TriageServiceRequestTool struct {
	conversation models.Conversation
}

func NewTriageServiceRequestTool() *TriageServiceRequestTool {
	return &TriageServiceRequestTool{}
}

func (t *TriageServiceRequestTool) Spec() toolx.ToolSpec {
	return toolx.GraphTriageServiceRequest
}

func (t *TriageServiceRequestTool) Name() string {
	return toolx.GraphTriageServiceRequest.Name
}

func (t *TriageServiceRequestTool) Code() string {
	return toolx.GraphTriageServiceRequest.Code
}

func (t *TriageServiceRequestTool) Enabled(ctx registry.Context) bool {
	return true
}

func (t *TriageServiceRequestTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &TriageServiceRequestTool{conversation: ctx.Conversation}, nil
}

func (t *TriageServiceRequestTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: toolx.GraphTriageServiceRequest.Name,
		Desc: "Graph Tool。用于综合分析当前对话，判断应该继续解答、整理工单草稿还是转人工；当判断为建单时，会一并返回结构化工单草稿建议。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "goal",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "当前分析目标，例如判断是否需要升级、是否要建单或转人工。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "observedIssue",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "当前观察到的主要问题或争议点。",
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
					Key: "additionalContext",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "补充上下文，例如你已发现的风险点或限制条件。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode":   toolx.GraphTriageServiceRequest.Code,
			"sourceType": "graph",
		},
	}, nil
}

func (t *TriageServiceRequestTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	return graphs.NewTriageServiceRequestGraph(t.conversation).Run(ctx, argumentsInJSON)
}
