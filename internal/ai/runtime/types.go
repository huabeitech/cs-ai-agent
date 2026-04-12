package runtime

import runtimeapp "cs-agent/internal/ai/runtime/app"

// TODO 这里为什么再设置一下类型别名，不能直接在外部使用 runtimeapp.Request 之类的类型呢？
type Request = runtimeapp.Request
type ResumeRequest = runtimeapp.ResumeRequest
type InterruptContextSummary = runtimeapp.InterruptContextSummary
type Summary = runtimeapp.Summary
