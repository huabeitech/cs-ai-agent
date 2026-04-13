package skills

import "cs-agent/internal/models"

// BuildRunLog 根据执行计划与运行结果构建 Skill 运行日志。
func BuildRunLog(ctx RuntimeContext, plan *ExecutionPlan, trace *ExecutionTrace, err error) *models.SkillRunLog {
	return RuntimeService.runlog.Build(ctx, plan, trace, err)
}

// WriteRunLog 写入 Skill 路由日志。
func WriteRunLog(log *models.SkillRunLog) error {
	return RuntimeService.runlog.Write(log)
}
