---
description: 生成或修改 Markdown 文件时自动应用规范检查
---

# Markdown 规范检查 Workflow

当 Agent 自动生成或修改 Markdown 文件时，自动触发此流程以确保符合规范。

## 触发条件

- 创建新的 `.md` 文件
- 修改现有 `.md` 文件
- 生成文档（README、Changelog、API 文档等）

## 执行流程

### Step 1：应用 Markdown Skill

参考 [markdown](file:///e:/go/go-utils/.agent/skills/core/markdown/SKILL.md) skill，确保生成的内容符合以下规范：

**必检规则：**

| 规则 | 检查项 |
| :--- | :--- |
| MD001 | 标题层级只能递增 1 级 |
| MD018 | `#` 后必须有空格 |
| MD022 | 标题前后需有空行 |
| MD031 | 代码块前后需有空行 |
| MD032 | 列表前后需有空行 |
| MD040 | 代码块需指定语言 |
| MD047 | 文件末尾需有空行 |

### Step 2：格式化检查

在生成内容时自动应用：

1. **表格格式**：使用 `| :--- |` 左对齐格式
2. **列表符号**：统一使用 `-` 作为无序列表符号
3. **代码块**：始终指定语言标识符
4. **空行**：标题、代码块、列表前后各留一空行

### Step 3：自动修复

如发现以下问题，自动修复：

- 行尾多余空格
- 连续多个空行 → 保留一个
- 缺少文件末尾空行 → 添加

## ⚠️ 注意事项

- 不修改代码块内的内容
- 保留用户有意为之的特殊格式
- 对于已存在的文件，只修改新增/变更的部分

## 适用场景

- `doc-generator` skill 生成文档时
- 创建 SKILL.md、README.md、CHANGELOG.md 等
- 编写 implementation_plan.md、walkthrough.md 等 artifacts
