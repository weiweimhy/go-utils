---
name: 产品经理翻译官
description: 将"人话"转译为技术语言，把模糊需求转化为清晰的 User Stories 或 PRD 文档
---

# Skill: Product Manager

将"人话"转译为"技术语言"，帮助梳理模糊的需求，将其转化为清晰的 User Stories 或 PRD 文档。

## 🎯 Purpose

当老板只给了一句话需求时，帮助拆解为开发任务清单、User Stories 或结构化的 PRD 文档，解决"需求传递失真"问题。

## 🎯 触发条件

当用户说：

- "帮我把这个需求写成 User Story"
- "老板说要做 xxx，帮我拆解一下"
- "把这个想法转成 PRD"
- "这个需求怎么拆分成开发任务"
- "用户说 xxx，技术上怎么实现"

👉 自动启用本 Skill

## 🧩 Capabilities

- **需求澄清**：通过结构化提问挖掘隐藏需求
- **User Story 生成**：按标准格式生成可执行的用户故事
- **PRD 生成**：生成结构化的产品需求文档
- **任务拆解**：将大需求拆分为开发任务清单
- **验收标准定义**：明确每个需求的 Done Criteria

## 🔄 Core Patterns

### 1. 需求澄清框架 (5W1H+)

在转译前，先通过问题挖掘完整需求：

| 维度 | 问题 | 示例 |
| :--- | :--- | :--- |
| **Who** | 目标用户是谁？ | 普通用户 / 管理员 / VIP？ |
| **What** | 具体要做什么？ | 功能边界在哪里？ |
| **Why** | 为什么要做这个？ | 解决什么痛点？ |
| **When** | 什么时候用？ | 使用场景 / 触发条件？ |
| **Where** | 在哪里使用？ | 入口 / 平台 / 环境？ |
| **How** | 期望怎么操作？ | 交互流程 / 预期行为？ |
| **Scope** | 不做什么？ | 明确排除的范围 |

> [!TIP]
> 不需要问完所有问题，根据需求复杂度选择关键问题

### 2. User Story 格式

标准 User Story 格式：

```markdown
## User Story: [简短标题]

**As a** [角色/用户类型]
**I want** [功能/行为]
**So that** [价值/目标]

### Acceptance Criteria (验收标准)

- [ ] Given [前置条件], When [操作], Then [预期结果]
- [ ] Given [前置条件], When [操作], Then [预期结果]

### Notes (补充说明)

- 技术约束 / 依赖 / 风险点
```

### 3. PRD 文档结构

```markdown
# [产品/功能名称] PRD

## 1. 背景与目标

- **背景**：为什么要做这个功能
- **目标**：期望达成什么效果
- **成功指标**：如何衡量成功 (KPI/OKR)

## 2. 用户与场景

- **目标用户**：主要用户群体
- **使用场景**：典型使用场景描述

## 3. 功能需求

### 3.1 [功能模块 1]

- 功能描述
- 用户流程
- 界面要求

## 4. 非功能需求

- 性能要求
- 安全要求
- 兼容性要求

## 5. 验收标准

- [ ] 功能验收项 1
- [ ] 功能验收项 2

## 6. 排期与里程碑

| 阶段 | 内容 | 预估时间 |
|------|------|----------|
| 设计 | UI/UX 设计 | x 天 |
| 开发 | 前后端开发 | x 天 |
| 测试 | 功能测试 | x 天 |

## 7. 风险与依赖

- 技术风险
- 业务依赖
```

### 4. 任务拆解模板

```markdown
## 任务拆解：[需求标题]

### Epic (史诗)

[大需求描述]

### User Stories (用户故事)

1. **[Story 1]** - [简述] - 优先级: P0/P1/P2
2. **[Story 2]** - [简述] - 优先级: P0/P1/P2

### Tasks (开发任务)

#### Story 1 子任务

- [ ] [后端] 任务描述 (~x h)
- [ ] [前端] 任务描述 (~x h)
- [ ] [测试] 任务描述 (~x h)

### 依赖关系

- Task A → Task B (A 完成后才能开始 B)
```

## 🧠 工作流程

### Step 1: 接收原始需求

获取用户输入的"人话"需求，可能是：

- 一句话需求："做个登录功能"
- 简短描述："用户要能查看自己的订单"
- 业务诉求："提升用户留存率"

### Step 2: 需求澄清

使用 5W1H+ 框架提出关键问题：

> [!IMPORTANT]
> 不要假设！如果有不明确的点，必须向用户确认

### Step 3: 需求转译

根据需求复杂度选择输出格式：

| 需求规模 | 推荐输出 |
| :--- | :--- |
| 小需求 (< 1天) | User Story |
| 中需求 (1-5天) | User Stories + 任务拆解 |
| 大需求 (> 5天) | 完整 PRD |

### Step 4: 验收确认

- 与用户确认转译结果
- 补充遗漏的验收标准
- 识别技术风险和依赖

## ✅ Best Practices

1. **不假设**：模糊的地方一定要问清楚
2. **分层拆解**：Epic → Story → Task
3. **可验证**：每个需求都有明确的验收标准
4. **优先级**：标注 P0/P1/P2 优先级
5. **估时**：给出合理的工时估算
6. **边界清晰**：明确说明"不做什么"

## ❌ Anti-Patterns

| 问题 | 严重性 | 解决方案 |
| :--- | :--- | :--- |
| 直接开始拆解，不问清楚需求 | 🔴 高 | 先用 5W1H 澄清 |
| User Story 缺少验收标准 | 🔴 高 | 补充 Given-When-Then |
| 任务粒度太大 | 🟡 中 | 拆分到可在 1 天内完成 |
| 没有标注优先级 | 🟡 中 | 添加 P0/P1/P2 |
| 假设用户意图 | 🔴 高 | 确认不确定的点 |

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| User Story 格式 | `Agile/Scrum 标准` | ✅ 遵循业界标准格式 |
| PRD 结构 | `ChatPRD` (外部) | ✅ 参考专业 PRD 生成工具 |
| 需求澄清框架 | `5W1H 方法论` | ✅ 经典需求分析方法 |
| 验收标准 | `BDD (Given-When-Then)` | ✅ 行为驱动开发规范 |
| 任务拆解 | `Agile Epic-Story-Task` | ✅ 敏捷开发分层模型 |

## 📚 参考资料 (References)

- [ChatPRD](https://chatprd.ai/) - AI PRD 生成工具
- [Writing User Stories](https://www.mountaingoatsoftware.com/agile/user-stories) - User Story 编写指南
- [Spec-Driven Development](https://medium.com/) - AI 需求编译理念
- [BDD - Cucumber](https://cucumber.io/docs/bdd/) - 行为驱动开发
- [INVEST Criteria](https://www.agilealliance.org/glossary/invest/) - User Story 质量标准

## ⚠️ Constraints

- 本技能只负责需求转译，不负责技术方案设计
- 需要用户配合澄清需求细节
- 工时估算仅供参考，需技术团队确认

## 🔗 Related Skills

- [prompt-design](file:///e:/go/go-utils/.agent/skills/core/ai/prompt-design/SKILL.md): 可用于优化需求提问方式
- [doc-generator](file:///e:/go/go-utils/.agent/skills/core/ai/doc-generator/SKILL.md): 可配合生成技术文档
- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md): 需求转技术后的 API 设计
