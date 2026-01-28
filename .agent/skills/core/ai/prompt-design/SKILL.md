---
name: Prompt 设计与优化
description: 帮助设计和优化 LLM Prompts，涵盖结构设计、Few-Shot Learning、Chain-of-Thought 等核心技术
---

# Skill: Prompt Design

帮助设计和优化 LLM Prompts，使模型产出更准确、一致且符合预期的结果。

## 🎯 Purpose

将用户意图转化为 LLM 能准确执行的指令。Prompt 设计需要如同代码一样严谨——小改动可能带来大影响，直觉判断往往不可靠，需要系统化评估。

## 🎯 触发条件

当用户说：

- "帮我设计一个 prompt"
- "优化这个 prompt"
- "写一个 system prompt"
- "这个 prompt 怎么改进"
- `/optimize-prompt` (workflow 触发)

👉 自动启用本 Skill

## 🧩 Capabilities

- **Prompt 结构设计**：构建清晰的 Role/Context/Instructions/Constraints/Output 架构
- **Few-Shot Learning**：通过示例教会模型预期行为
- **Chain-of-Thought**：引导模型进行逐步推理
- **Template Systems**：构建可复用的 Prompt 模板
- **System Prompt 架构**：设计持久化的全局行为约束
- **Progressive Disclosure**：渐进式增加 Prompt 复杂度

## 🔄 Core Patterns

### 1. Structured System Prompt

结构化的系统提示词设计：

```markdown
[Role]       谁 - 模型扮演的角色
[Context]    什么 - 相关背景信息
[Instructions] 怎么做 - 具体任务指令
[Constraints]  不做什么 - 明确限制
[Output]     输出格式 - 期望的结构
[Examples]   示例 - 正确行为的演示
```

### 2. Few-Shot Learning

通过示例教会模型：

```markdown
提取支持工单中的关键信息：

Input: "登录失败，一直报 403 错误"
Output: {"issue": "authentication", "error_code": "403", "priority": "high"}

Input: "希望能在设置中添加深色模式"
Output: {"issue": "feature_request", "error_code": null, "priority": "low"}

现在处理: "上传超过 10MB 的文件时超时"
```

> [!TIP]
> 包含 2-5 个多样化示例，涵盖边缘情况，保持格式一致

### 3. Chain-of-Thought

请求逐步推理：

```markdown
分析这个 Bug 报告并确定根本原因。

请逐步思考：
1. 预期行为是什么？
2. 实际行为是什么？
3. 最近有什么变更可能导致此问题？
4. 涉及哪些组件？
5. 最可能的根本原因是什么？

Bug: "昨天部署缓存更新后，用户无法保存草稿"
```

### 4. Progressive Disclosure

从简单开始，按需增加复杂度：

| 级别 | 策略 | 示例 |
| :--- | :--- | :--- |
| L1 | 直接指令 | "总结这篇文章" |
| L2 | 添加约束 | "用 3 个要点总结这篇文章，聚焦核心发现" |
| L3 | 添加推理 | "阅读文章，识别主要发现，然后用 3 个要点总结" |
| L4 | 添加示例 | 包含 2-3 个输入输出示例 |

### 5. Instruction Hierarchy

标准指令层次结构：

```text
[System Context] → [Task Instruction] → [Examples] → [Input Data] → [Output Format]
```

## ✅ Best Practices

1. **具体明确**：模糊的 Prompt 产生不一致的结果
2. **示例优于描述**：Show, Don't Tell
3. **充分测试**：在多样化、有代表性的输入上评估
4. **快速迭代**：小改动可能有大影响
5. **监控性能**：在生产环境中跟踪指标
6. **版本控制**：像代码一样管理 Prompt 版本
7. **记录意图**：解释为什么这样设计 Prompt

## ❌ Anti-Patterns

| 问题 | 严重性 | 解决方案 |
| :--- | :--- | :--- |
| 使用不精确的语言 | 🔴 高 | 明确具体："列出 3 个要点" vs "列出一些要点" |
| 期望特定格式但未指定 | 🔴 高 | 明确指定输出格式 |
| 只说做什么，不说不做什么 | 🟡 中 | 包含明确的限制条件 |
| 修改 Prompt 但不测量影响 | 🟡 中 | 系统化评估对比 |
| "以防万一"包含无关上下文 | 🟡 中 | 精心筛选上下文内容 |
| 示例有偏差或不具代表性 | 🟡 中 | 使用多样化示例 |
| 所有任务使用默认 temperature | 🟡 中 | 根据任务调整 temperature |
| 未考虑用户输入的 Prompt 注入 | 🔴 高 | 防御注入攻击 |

## ⚠️ Common Pitfalls

- **过度工程化**：在尝试简单方案前就使用复杂 Prompt
- **示例污染**：使用与目标任务不匹配的示例
- **上下文溢出**：过多示例超出 Token 限制
- **指令歧义**：留有多种解释的空间
- **忽视边缘情况**：未在异常或边界输入上测试

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| Prompt 结构设计 | `prompt-engineer` (外部) | ✅ 整合 Role/Context/Instructions 架构 |
| Few-Shot Learning | `prompt-engineering` (外部) | ✅ 整合核心技术与示例 |
| Chain-of-Thought | `prompt-engineering` (外部) | ✅ 整合推理引导技术 |
| Template Systems | `prompt-engineering` (外部) | ✅ 整合模板系统设计 |
| Progressive Disclosure | `prompt-engineering` (外部) | ✅ 整合渐进式复杂度模式 |
| Best Practices | `Prompt Engineering Guide` | ✅ 参考行业最佳实践 |

## 📚 参考资料 (References)

- [Prompt Engineering Guide](https://www.promptingguide.ai/) - 全面的 Prompt 工程指南
- [antigravity-awesome-skills/prompt-engineer](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/prompt-engineer) - 外部 Prompt 工程师 Skill
- [antigravity-awesome-skills/prompt-engineering](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/prompt-engineering) - 外部 Prompt 工程模式 Skill
- [OpenAI Prompt Engineering Guide](https://platform.openai.com/docs/guides/prompt-engineering) - OpenAI 官方指南
- [Anthropic Prompt Design Guide](https://docs.anthropic.com/claude/docs/prompt-design) - Anthropic 官方指南

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 可用 Prompt 设计优化审查反馈格式
- [skill-builder](file:///e:/go/go-utils/.agent/skills/skill-builder/SKILL.md): 创建新 Skill 时可参考 Prompt 设计原则
