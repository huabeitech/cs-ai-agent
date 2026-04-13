package executor

import internalexecutor "cs-agent/internal/ai/runtime/internal/executor"

// TODO 为什么要定义类型别名？
type RunInput = internalexecutor.RunInput
type ResumeInput = internalexecutor.ResumeInput
type InterruptContextSummary = internalexecutor.InterruptContextSummary
type RunResult = internalexecutor.RunResult
