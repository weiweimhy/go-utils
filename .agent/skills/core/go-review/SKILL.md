---
name: Go 代码审查
description: 针对 Go 语言代码的专家级审查，遵循 Effective Go 及官方 Code Review Comments 准则
---

# Skill: Go Review

针对 Go 语言代码进行深度审查，确保其符合 [Effective Go](https://go.dev/doc/effective_go) 哲学及 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) 规范。

## 🎯 触发条件

当以下情况发生时启用：

- "审查这段 Go 代码"
- "这段代码够不够 Idiomatic Go？"
- "帮我看看这个并发处理有没有死锁风险"
- "优化这个 Go 函数的性能"

👉 自动启用本 Skill

## 🎯 Purpose

提供地道的 (Idiomatic) Go 代码审查，识别潜在的 Bug、性能隐患、并发不安全模式及可读性问题。

## 🧩 Capabilities

- **Idiomatic Go 校验**：
  - **命名规范**：审查变量、函数、包名是否符合 Go 惯例（简洁、避免 Stuttering）。
  - **接口设计**：遵循“接收接口，返回结构”原则；只在必要时定义接口。
  - **错误处理**：确保错误被恰当包装（Wrap）而非仅返回；严禁静默忽略错误。
- **并发安全审查**：
  - **Goroutine 生命周期**：每个 Goroutine 必须有明确的生命周期管理（Context 取消或退出信号）。
  - **竞态检测**：识别对共享资源的未同步访问。
  - **Channel 使用**：审查 Channel 是否有缓冲需求，避免阻塞导致的 Goroutine 泄漏。
- **性能与资源评估**：
  - **内存分配**：识别高频路径下的对象逃逸、切片容量预分配不足等问题。
  - **标准库应用**：优先使用标准库提供的同步原语（Sync 库）及原子操作。
- **测试质量检查**：
  - **表格驱动测试**：推荐使用 Table-Driven Tests 模式以增强测试覆盖面及清晰度。
  - **竞态测试**：确认是否执行了 `-race` 检测。

## 🧠 Usage

当以下情况发生时调用：

- 作为 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 流程的第二阶段，针对 Go 代码进行深度评估。
- 提交 Go 代码 PR 前进行的自我审查或交叉审查。
- 自动化流水线捕获到 Lint 或 Staticcheck 异常时。
- 针对高性能或高并发模块的代码审阅。

## 📥 Input

- 待审查的 `.go` 文件或 Diff。
- 如果涉及业务逻辑，需提供背景说明。

## 📤 Output

结构化的 Go 审查报告（参见 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 的 Output 格式），并增加：

- **Idiomatic Examples**: 展示如何用更地道的方式重写非 Go 风格的代码。
- **Optimization Suggestions**: 具体的代码性能优化建议。

## ⚠️ Constraints

- ❌ **不得强制要求特定第三方库**（如强制要求使用 sonic 等），应优先保证对标准库的兼容性，除非项目有明确约定。
- ❌ **不重复 Lint 任务**：不指正 `gofmt` 或 `goimports` 能够自动解决的格式问题。
- ✅ **优先考虑简洁性**：Go 的设计哲学是 Simple & Explicit。
- ✅ **关注 Panic 风险**：严禁在未恢复的情况下产生非预期的 Panic。

## 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| Idiomatic Go 校验 | `effective-go` | ✅ 语言官方地道指南 |
| 并发安全审计 | `go-code-review-comments` | ✅ 核心团队审查准则 |
| 性能与内存评估 | `uber-go-style-guide` | ✅ 工业界高性能实践 |
| 接口设计准则 | `effective-go` | ✅ 语言官方架构建议 |

## 📚 参考资料 (References)

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 基础审查哲学与流程。
- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md): 涉及 API 层面的设计审查。
- [backend-patterns](file:///e:/go/go-utils/.agent/skills/core/backend/backend-patterns/SKILL.md): 涉及后端架构层面的设计审查。
