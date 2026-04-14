package tools

import (
	"context"
	"fmt"

	"cs-agent/internal/ai/runtime/graphs"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	einojsonschema "github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type PrepareTicketDraftTool struct {
	conversation *models.Conversation
}

func NewPrepareTicketDraftTool() *PrepareTicketDraftTool {
	return &PrepareTicketDraftTool{}
}

func (t *PrepareTicketDraftTool) Spec() toolx.ToolSpec {
	return toolx.GraphPrepareTicketDraft
}

func (t *PrepareTicketDraftTool) Name() string {
	return toolx.GraphPrepareTicketDraft.Name
}

func (t *PrepareTicketDraftTool) Code() string {
	return toolx.GraphPrepareTicketDraft.Code
}

func (t *PrepareTicketDraftTool) Enabled(ctx registry.Context) bool {
	return ctx.Conversation != nil
}

func (t *PrepareTicketDraftTool) Build(ctx registry.Context) (einotool.BaseTool, error) {
	if !t.Enabled(ctx) {
		return nil, nil
	}
	return &PrepareTicketDraftTool{conversation: ctx.Conversation}, nil
}

func (t *PrepareTicketDraftTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: toolx.GraphPrepareTicketDraft.Name,
		Desc: "Graph Tool。用于根据当前会话和已收集信息整理工单草稿，输出建议标题、建议描述、缺失字段和追问建议。适合在真正调用 create_ticket_with_confirmation 前先整理工单内容。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&einojsonschema.Schema{
			Version: einojsonschema.Version,
			Type:    "object",
			Properties: orderedmap.New[string, *einojsonschema.Schema](orderedmap.WithInitialData(
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "title",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "已整理出的工单标题，可选。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "description",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "已整理出的工单描述，可选。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "issue",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "用户当前遇到的问题现象或报错信息。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "impact",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "问题影响范围，例如无法登录、无法下单、业务中断等。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "expectedOutcome",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "用户期望的处理结果或诉求。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "currentAttempt",
					Value: &einojsonschema.Schema{
						Type:        "string",
						Description: "当前已尝试过的处理步骤，可选。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "priority",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "建议工单优先级，可选。",
					},
				},
				orderedmap.Pair[string, *einojsonschema.Schema]{
					Key: "severity",
					Value: &einojsonschema.Schema{
						Type:        "integer",
						Description: "建议严重度，可选；1=轻微，2=严重，3=致命。",
					},
				},
			)),
		}),
		Extra: map[string]any{
			"toolCode":   toolx.GraphPrepareTicketDraft.Code,
			"sourceType": "graph",
		},
	}, nil
}

func (t *PrepareTicketDraftTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...einotool.Option) (string, error) {
	if t == nil || t.conversation == nil {
		return "", fmt.Errorf("prepare ticket draft tool not initialized")
	}
	return graphs.NewPrepareTicketDraftGraph(t.conversation).Run(ctx, argumentsInJSON)
}
