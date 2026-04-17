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

type CreateTicketGraphTool struct {
	conversation models.Conversation
	aiAgent      models.AIAgent
}

func NewCreateTicketGraphTool() *CreateTicketGraphTool {
	return &CreateTicketGraphTool{}
}

func (t *CreateTicketGraphTool) Spec() toolx.ToolSpec {
	return toolx.GraphCreateTicketConfirm
}

func (t *CreateTicketGraphTool) Name() string {
	return toolx.GraphCreateTicketConfirm.Name
}

func (t *CreateTicketGraphTool) Code() string {
	return toolx.GraphCreateTicketConfirm.Code
}

func (t *CreateTicketGraphTool) Enabled(ctx registry.Context) bool {
	return true
}

func (t *CreateTicketGraphTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &CreateTicketGraphTool{
		conversation: ctx.Conversation,
		aiAgent:      ctx.AIAgent,
	}, nil
}

func (t *CreateTicketGraphTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: toolx.GraphCreateTicketConfirm.Name,
		Desc: "Graph Tool。用于封装建单参数整理、用户确认、真正创建工单和结果返回的确定性流程。仅在用户明确要求建单且标题、描述已整理清楚后调用。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Required: []string{
				"title",
				"description",
			},
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "title",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "工单标题，简洁概括问题。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "description",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "工单描述，清晰整理用户问题、现象和诉求。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "priority",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "工单优先级，可选；未知时可不传。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "severity",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "严重度，可选；1=轻微，2=严重，3=致命。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode":   toolx.GraphCreateTicketConfirm.Code,
			"sourceType": "graph",
		},
	}, nil
}

func (t *CreateTicketGraphTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	return graphs.NewCreateTicketGraph(t.conversation, t.aiAgent).Run(ctx, argumentsInJSON)
}
