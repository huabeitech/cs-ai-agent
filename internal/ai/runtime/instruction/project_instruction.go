package instruction

// DefaultProjectInstruction 为当前项目统一注入的全局项目规则。
//
// 默认回退到内置常量；运行时优先由 ProjectInstructionProvider 读取仓库中的 AGENTS.md。
// TODO 这里是做什么的？为什么不直接读取AIAgent的？
const DefaultProjectInstruction = `# AGENTS.md

本文件定义本项目内 AI Agent 的强制开发规则。除非用户明确要求偏离，否则必须遵循。

## 1. 基本原则

- 适用范围：仓库根目录及所有子目录
- 优先级：用户明确指令 > 本文件 > 默认实现习惯
- 若与用户要求冲突：先执行用户要求，并在变更说明中标注偏离点
`
