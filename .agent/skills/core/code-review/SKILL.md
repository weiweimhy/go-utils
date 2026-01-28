---
name: 通用代码审查
description: 定义跨语言的代码审查哲学、流程及核心检查单，侧重于可维护性、逻辑正确性及团队协作
---

# Skill: Code Review

本 Skill 提供通用的代码审查指导，旨在提升代码质量、共享知识并保持工程一致性。

## 🎯 触发条件

当以下情况发生时启用：

- "帮我做一下代码审查 (Code Review)"
- "看看这段代码改得怎么样？"
- "/review" (workflow 触发)
- "检查一下这段代码的逻辑"

👉 自动启用本 Skill

## 🎯 Purpose

通过建设性的反馈，确保代码逻辑正确、易于维护、安全且符合团队的最佳实践。

## 🧩 Capabilities

- **审查哲学引导**：推崇同理心、建设性反馈及以“问题”代替“命令”的沟通方式。
- **通用 Checklist**：
  - **正确性**：是否解决了目标问题？是否处理了边界情况？
  - **可读性**：代码是否易于理解？命名是否清晰？
  - **简洁性**：是否存在过度设计或重复代码？
  - **测试**：是否有足够的单元测试？测试是否涵盖了核心逻辑和错误路径？
  - **安全与性能**：是否存在明显的安全漏洞或效率极低的实现？
- **审查流程路由**：作为审查的入口点，首先执行通用质量评估，随后识别语言并路由至专项 Skill（如 `go-review`）。
- **流程规范**：建议小步提交（200-400行）、优先自动化检查、明确审查意见的优先级（如 nitpick, blocker）。

## 🧠 Usage

## 🔄 Workflow

在使用本 Skill 时，应遵循以下执行顺序：

1. **通用评估**：使用本 Skill 的 Checklist 对逻辑、可读性、安全性进行初步扫描。
2. **语言识别**：根据输入内容识别编程语言。
3. **专项深挖**：
   - 如果是 **Go**：必须紧接着调用 [go-review](file:///e:/go/go-utils/.agent/skills/core/go-review/SKILL.md) 进行深度地道性审查。
   - 如果是 **Python**：寻找并调用 `python-review`（如果存在），否则维持通用审查标准。
   - 其他语言以此类推。
4. **综述汇总**：整合通用建议与专项建议，给出最终报告。

## 📥 Input

- 待审查的代码片段或文件路径。
- 变更背景、业务需求及相关的设计文档（可选）。
- 变更范围（Diff）。

## 📤 Output

结构化的审查意见，建议包含：

- **Summary**: 对变更的整体评价和理解。
- **Positive Feedback**: 值得肯定的设计或实现点。
- **Actionable Feedback**:
  - **Blockers**: 必须修复的问题（Bug、安全隐患、严重违背架构原则）。
  - **Suggestions**: 建议改进的地方，用于提升可读性或性能。
  - **Nitpicks**: 微小的细节改动（如拼写、格式）。
- **Rationales**: 提供具体的理由或参考文档，帮助开发者理解为何建议修改。

## ⚠️ Constraints

- ❌ **严禁攻击性言论**：审查应针对代码而非人。
- ❌ **避免风格争议**：风格问题应尽量交给 Linter 自动处理。
- ✅ **优先考虑代码的可维护性**：今天能运行的代码不代表明天能维护。
- ✅ **鼓励提问**：如果不确定代码的意图，先提问而非直接下结论。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| 审查哲学与心态 | `google-code-review-guide` | ✅ 行业公认最佳指南 |
| 反馈质量与沟通 | `palantir-code-review` | ✅ 工业界建设性协作规范 |
| 代码质量 Checklist | `airbnb-best-practices` | ✅ 社区主流质量基准 |
| 流程路由与自动化 | `industry-standard-cicd` | ✅ 工程化通用实践 |

## 📚 参考资料 (References)

- [Google Code Review Developer Guide](https://google.github.io/eng-practices/review/)
- [Palantir: How to Review Code](https://github.com/palantir/gradle-revapi/blob/master/docs/how-to-review-code.md)
- [Airbnb Engineering Best Practices](https://github.com/airbnb/javascript#code-review-guide)

## 🔗 Related Skills

- [go-review](file:///e:/go/go-utils/.agent/skills/core/go-review/SKILL.md): 针对 Go 语言的深度代码审查。
- [git-workflow](file:///e:/go/go-utils/.agent/skills/core/git-workflow/SKILL.md): 在提交流程中集成代码审查。
