package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/errorsx"
)

var MCPDebugService = newMCPDebugService()

func newMCPDebugService() *mCPDebugService {
	return &mCPDebugService{
		client: mcps.NewClient(),
	}
}

type mCPDebugService struct {
	client *mcps.Client
}

func (s *mCPDebugService) ListServers(cfg *config.Config) []mcps.ServerInfo {
	if cfg == nil || len(cfg.MCP.Servers) == 0 {
		return nil
	}
	keys := make([]string, 0, len(cfg.MCP.Servers))
	for code := range cfg.MCP.Servers {
		keys = append(keys, code)
	}
	slices.Sort(keys)

	ret := make([]mcps.ServerInfo, 0, len(keys))
	for _, code := range keys {
		server := cfg.MCP.Servers[code]
		ret = append(ret, mcps.ServerInfo{
			Code:      code,
			Enabled:   server.Enabled,
			Endpoint:  strings.TrimSpace(server.Endpoint),
			TimeoutMS: server.TimeoutMS,
		})
	}
	return ret
}

func (s *mCPDebugService) TestConnection(ctx context.Context, cfg *config.Config, serverCode string) (*mcps.ConnectionResult, error) {
	server, err := s.resolveServer(cfg, serverCode)
	if err != nil {
		return nil, err
	}
	startedAt := time.Now()
	result, err := s.client.TestConnection(ctx, server)
	s.logResult("test_connection", serverCode, "", time.Since(startedAt), err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *mCPDebugService) ListTools(ctx context.Context, cfg *config.Config, serverCode string) ([]mcps.ToolInfo, error) {
	server, err := s.resolveServer(cfg, serverCode)
	if err != nil {
		return nil, err
	}
	startedAt := time.Now()
	result, err := s.client.ListTools(ctx, server)
	s.logResult("list_tools", serverCode, "", time.Since(startedAt), err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *mCPDebugService) CallTool(ctx context.Context, cfg *config.Config, serverCode string, toolName string, arguments map[string]any) (*mcps.ToolCallResult, error) {
	server, err := s.resolveServer(cfg, serverCode)
	if err != nil {
		return nil, err
	}
	startedAt := time.Now()
	result, err := s.client.CallTool(ctx, server, toolName, arguments)
	s.logResult("call_tool", serverCode, toolName, time.Since(startedAt), err)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *mCPDebugService) resolveServer(cfg *config.Config, serverCode string) (mcps.ServerConfig, error) {
	if cfg == nil {
		return mcps.ServerConfig{}, fmt.Errorf("系统配置不存在")
	}
	if !cfg.MCP.Enabled {
		return mcps.ServerConfig{}, errorsx.InvalidParam("MCP未启用")
	}
	serverCode = strings.TrimSpace(serverCode)
	if serverCode == "" {
		return mcps.ServerConfig{}, errorsx.InvalidParam("serverCode不能为空")
	}
	server, ok := cfg.MCP.Servers[serverCode]
	if !ok {
		return mcps.ServerConfig{}, errorsx.InvalidParam("MCP服务配置不存在")
	}
	if !server.Enabled {
		return mcps.ServerConfig{}, errorsx.InvalidParam("MCP服务未启用")
	}
	return mcps.ServerConfig{
		Code:      serverCode,
		Endpoint:  strings.TrimSpace(server.Endpoint),
		TimeoutMS: server.TimeoutMS,
		Headers:   cloneHeaders(server.Headers),
	}, nil
}

func (s *mCPDebugService) logResult(action string, serverCode string, toolName string, elapsed time.Duration, err error) {
	fields := []any{
		"action", action,
		"server_code", serverCode,
		"tool_name", toolName,
		"elapsed_ms", elapsed.Milliseconds(),
	}
	if err != nil {
		fields = append(fields, "success", false, "error", err.Error())
		slog.Warn("mcp debug request failed", fields...)
		return
	}
	fields = append(fields, "success", true)
	slog.Info("mcp debug request finished", fields...)
}

func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	ret := make(map[string]string, len(headers))
	for key, value := range headers {
		ret[key] = value
	}
	return ret
}

func DumpPayload(value any) string {
	if value == nil {
		return ""
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(buf)
}
