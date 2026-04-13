package skills

import (
	"context"
)

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	return RuntimeService.BuildExecutionPlan(execCtx, ctx)
}

// Select 执行一次 Skill 路由并记录路由日志。
func Select(ctx context.Context, runtimeCtx RuntimeContext) (*ExecutionResult, error) {
	return RuntimeService.Select(ctx, runtimeCtx)
}
