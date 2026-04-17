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

type HandoffGraphTool struct {
	conversation models.Conversation
	aiAgent      models.AIAgent
}

func NewHandoffGraphTool() *HandoffGraphTool {
	return &HandoffGraphTool{}
}

func (t *HandoffGraphTool) Spec() toolx.ToolSpec {
	return toolx.GraphHandoffConversation
}

func (t *HandoffGraphTool) Name() string {
	return toolx.GraphHandoffConversation.Name
}

func (t *HandoffGraphTool) Code() string {
	return toolx.GraphHandoffConversation.Code
}

func (t *HandoffGraphTool) Enabled(ctx registry.Context) bool {
	return true
}

func (t *HandoffGraphTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &HandoffGraphTool{
		conversation: ctx.Conversation,
		aiAgent:      ctx.AIAgent,
	}, nil
}

func (t *HandoffGraphTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: toolx.GraphHandoffConversation.Name,
		Desc: "Graph Tool。用于封装转人工原因整理、用户确认、真正转人工和结果返回的确定性流程。仅在用户明确要求人工客服，或你已确认必须转人工处理时调用。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "reason",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "转人工原因，简洁说明为何需要人工介入，例如用户明确要求人工、问题需要人工核验、需要人工售后处理等。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode":   toolx.GraphHandoffConversation.Code,
			"sourceType": toolx.GraphHandoffConversation.SourceType,
		},
	}, nil
}

func (t *HandoffGraphTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	return graphs.NewHandoffGraph(t.conversation, t.aiAgent).Run(ctx, argumentsInJSON)
}
