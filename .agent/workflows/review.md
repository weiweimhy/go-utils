---
description: 如何执行代码审查 (Code Review)
---

# Code Review Workflow

在使用 `/review` 命令或请求代码审查时，请遵循以下步骤：

1. **调用通用审查能力**：
   使用 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 评估代码的基础质量。重点关注：
   - 业务逻辑正确性
   - 代码可读性与命名
   - 基础安全性与潜在 Bug

2. **分析编程语言**：
   识别代码段或文件所属的编程语言。

3. **执行专项深度审查**：
   根据识别出的语言，调用相应的专项 Skill：
   - 如果是 **Go**：必须调用 [go-review](file:///e:/go/go-utils/.agent/skills/core/go-review/SKILL.md) 检查地道性 (Idiomatic Go)、并发安全及性能。
   - 如果是 **Python**：寻找并使用 `python-review`。
   - 如果是 **Lua**：必须调用 [lua-review](file:///e:/go/go-utils/.agent/skills/core/lua-review/SKILL.md) 检查全局污染、作用域及内存效率。
   - 如果是 **C#**：必须调用 [csharp-review](file:///e:/go/go-utils/.agent/skills/core/csharp-review/SKILL.md) 检查异步安全性、.NET 现代特性及资源管理。
   - 如果是 **TypeScript**：必须调用 [typescript-review](file:///e:/go/go-utils/.agent/skills/core/typescript-review/SKILL.md) 检查类型严密性、泛型应用及 TS 5.x 现代特性。
   - 其他语言以此类推。

4. **汇总反馈**：
   将通用建议与专项建议整合，产出一份结构清晰、包含优先级（Blocker/Suggestion/Nitpick）的完整报告。
