package engine

import "cs-agent/internal/ai/runtime/internal/executor"

// TODO 这个地方为什么要定义类型别名，不能直接用吗？
type RunInput = executor.RunInput
type ResumeInput = executor.ResumeInput
type InterruptContextSummary = executor.InterruptContextSummary
type RunResult = executor.RunResult

type Request = RunInput
type ResumeRequest = ResumeInput
type Summary = RunResult
