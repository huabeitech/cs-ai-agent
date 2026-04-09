package mcps

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cs-agent/internal/pkg/errorsx"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) TestConnection(ctx context.Context, cfg ServerConfig) (*ConnectionResult, error) {
	session, closeFn, err := c.connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	initResult := session.InitializeResult()
	serverName := ""
	version := ""
	protocol := ""
	if initResult != nil {
		serverName = initResult.ServerInfo.Name
		version = initResult.ServerInfo.Version
		protocol = initResult.ProtocolVersion
	}
	return &ConnectionResult{
		ServerCode: cfg.Code,
		Endpoint:   cfg.Endpoint,
		Protocol:   protocol,
		ServerName: serverName,
		Version:    version,
	}, nil
}

func (c *Client) ListTools(ctx context.Context, cfg ServerConfig) ([]ToolInfo, error) {
	session, closeFn, err := c.connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("列出 MCP 工具失败: %w", err)
	}
	ret := make([]ToolInfo, 0, len(result.Tools))
	for _, tool := range result.Tools {
		ret = append(ret, ToolInfo{
			Name:         tool.Name,
			Title:        tool.Title,
			Description:  tool.Description,
			InputSchema:  tool.InputSchema,
			OutputSchema: tool.OutputSchema,
		})
	}
	return ret, nil
}

func (c *Client) CallTool(ctx context.Context, cfg ServerConfig, toolName string, arguments map[string]any) (*ToolCallResult, error) {
	toolName = strings.TrimSpace(toolName)
	if toolName == "" {
		return nil, errorsx.InvalidParam("toolName不能为空")
	}

	session, closeFn, err := c.connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 MCP 工具失败: %w", err)
	}
	return &ToolCallResult{
		ServerCode:        cfg.Code,
		ToolName:          toolName,
		IsError:           result.IsError,
		Content:           convertContents(result.Content),
		StructuredContent: result.StructuredContent,
	}, nil
}

func (c *Client) connect(ctx context.Context, cfg ServerConfig) (*mcp.ClientSession, func(), error) {
	if strings.TrimSpace(cfg.Code) == "" {
		return nil, nil, errorsx.InvalidParam("serverCode不能为空")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, nil, errorsx.InvalidParam("MCP endpoint不能为空")
	}

	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	connCtx, cancel := context.WithTimeout(ctx, timeout)

	httpClient := &http.Client{
		Transport: &headerRoundTripper{
			next:    http.DefaultTransport,
			headers: cfg.Headers,
		},
	}
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "cs-agent-mcp-client",
		Version: "v1",
	}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             cfg.Endpoint,
		HTTPClient:           httpClient,
		MaxRetries:           0,
		DisableStandaloneSSE: true,
	}
	session, err := client.Connect(connCtx, transport, nil)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("连接 MCP Server 失败: %w", err)
	}
	return session, func() {
		_ = session.Close()
		cancel()
	}, nil
}

func convertContents(contents []mcp.Content) []ToolResultContent {
	ret := make([]ToolResultContent, 0, len(contents))
	for _, item := range contents {
		switch v := item.(type) {
		case *mcp.TextContent:
			ret = append(ret, ToolResultContent{
				Type: "text",
				Text: v.Text,
			})
		case *mcp.ImageContent:
			ret = append(ret, ToolResultContent{
				Type: "image",
				Data: map[string]any{
					"mimeType": v.MIMEType,
					"data":     v.Data,
				},
			})
		case *mcp.AudioContent:
			ret = append(ret, ToolResultContent{
				Type: "audio",
				Data: map[string]any{
					"mimeType": v.MIMEType,
					"data":     v.Data,
				},
			})
		case *mcp.EmbeddedResource:
			ret = append(ret, ToolResultContent{
				Type: "resource",
				Data: v.Resource,
			})
		default:
			ret = append(ret, ToolResultContent{
				Type: "unknown",
				Data: v,
			})
		}
	}
	return ret
}

type headerRoundTripper struct {
	next    http.RoundTripper
	headers map[string]string
}

func (r *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	next := r.next
	if next == nil {
		next = http.DefaultTransport
	}
	clone := req.Clone(req.Context())
	for key, value := range r.headers {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		clone.Header.Set(key, value)
	}
	return next.RoundTrip(clone)
}
