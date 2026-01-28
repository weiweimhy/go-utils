---
description: 优化用户输入的 Prompt，使其更清晰、结构化、有效
---

# Prompt 优化 Workflow

当用户使用 `/optimize-prompt` 时触发此流程。

## 使用方式

```text
/optimize-prompt <你的原始 prompt>
```

## 执行流程

### Step 1：分析原始 Prompt

识别用户输入的 Prompt 存在的问题：

- 是否缺少明确的角色定义 (Role)？
- 是否缺少上下文背景 (Context)？
- 指令是否模糊或有歧义？
- 是否缺少约束条件 (Constraints)？
- 是否指定了输出格式 (Output Format)？

### Step 2：应用优化模式

参考 [prompt-design](file:///e:/go/go-utils/.agent/skills/core/ai/prompt-design/SKILL.md) skill，应用以下模式：

1. **Structured System Prompt**：重构为 Role/Context/Instructions/Constraints/Output 结构
2. **Progressive Disclosure**：根据任务复杂度选择合适的级别
3. **Few-Shot Examples**：如需要，添加示例（2-5 个）

### Step 3：输出优化结果

输出格式：

```markdown
## 📊 分析结果

| 问题 | 说明 |
| :--- | :--- |
| 问题 1 | 描述 |

## ✨ 优化后的 Prompt

<优化后的完整 prompt>

## 💡 优化说明

- 改动点 1：原因
- 改动点 2：原因
```

## ⚠️ 注意事项

- 不要改变用户的核心意图
- 简单请求（如"帮我看下这个文件"）不需要过度优化
- 保持用户的语言风格（中文/英文）
