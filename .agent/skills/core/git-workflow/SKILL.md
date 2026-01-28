---
name: Git 规范工作流
description: 帮助开发者生成规范、清晰、可维护的 Git 提交记录的全流程引导
---

# Skill: git-workflow

## 🎯 Purpose

提供从代码变更分析到生成符合 Conventional Commit 规范的提交信息的全流程引导。

## 🧩 Capabilities

- **总结变更**：分析代码 diff 并总结本次改动内容。
- **判断类型**：识别变更类型（feat / fix / refactor / docs / chore 等）。
- **生成消息**：生成符合规范的 Commit Message。
- **拆分建议**：根据变更规模和逻辑关系提示是否需要拆分提交。

## 🔍 Source Mapping

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| 变更总结 | `codex-review` | 用于深入分析代码语义并生成 CHANGELOG 风格的总结 |
| 类型判断 | `core/git-commit` | 遵循项目定义的提交类型规范 |
| 消息生成 | `core/git-commit` | 遵循项目定义的消息格式及动词开头要求 |
| 流程自动化 | `git-pushing` | 用于暂存、提交及推送的原子操作执行 |
| 任务安排 | `github-workflow-automation` | 用于 PR 关联及整体流程编排 |
| 拆分提示 | **新建** | 检查 diff 行数及文件分布，若过大则提示拆分 |

## 🧠 Usage

- 在准备提交本地变更前。
- 当开发者面对大量杂乱改动不知道如何写提交信息时。
- 在创建 Pull Request 前整理提交历史时。

## 📥 Input

- **代码变更 (diff)**：当前暂存或未暂存的文件差异。
- **意图描述**：开发者对本次改动的简短口头说明。

## 📤 Output

- **改动总结**：清晰的变更列表。
- **推荐 Message**：符合规范的提交信息建议。
- **拆分说明**：如果建议拆分，给出拆分方案。

## ⚠️ Constraints

- 不在未经用户确认的情况下直接执行 `git commit`。
- 不修改源代码逻辑。

## 🔗 Related Skills

- [git-commit](file:///e:/go/go-utils/.agent/skills/core/git-commit/SKILL.md)
- [release](file:///e:/go/go-utils/.agent/skills/core/release/SKILL.md)
- [codex-review](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/codex-review)
