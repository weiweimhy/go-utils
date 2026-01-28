---
name: 后端架构模式
description: 专注于项目结构、模块拆分及服务边界设计的架构级能力
---

# Skill: Backend Patterns

本 Skill 专注于后端系统的内在逻辑架构，确保代码结构清晰、高内聚低耦合，并具备长期的可维护性。

## 🎯 触发条件

当以下情况发生时启用：

- "帮我搭建项目基础架构"
- "代码模块该怎么拆分？"
- "如何解决这个循环依赖？"
- "设计一个数据库缓存方案"
- "实现一个带指数退避的重试机制"

👉 自动启用本 Skill

## 🎯 Purpose

作为“写久了才知道值钱”的核心技能，本 Skill 提供从单体到分布式的架构模式指导，帮助建立稳健的代码组织形式。

## 🧩 Capabilities

- **项目结构建议**：指导标准化的目录结构（如 Go 项目的 `cmd/`, `internal/`, `pkg/` 布局）。
- **职责边界设计**：清晰定义领域模型、服务层 (Service Layer) 与基础设施层 (Repository) 的职责边界，确保依赖方向由外向内。
- **中间件模式 (Middleware)**：设计请求预处理链路（鉴权、日志、限流）。
- **数据库最佳实践**：
  - **N+1 预防**：强制使用批量获取 (Batch Fetch) 逻辑替代循环查询。
  - **查询优化**：精确选择列名 (Select Fields)，严禁 `SELECT *`。
  - **事务模式**：指导实现跨资源的原子性操作。
- **缓存策略**：指导 Redis 缓存层设计，应用 Cache-Aside 模式及合理的缓存失效策略。
- **系统韧性与错误处理**：
  - **重试机制**：实现带指数退避 (Exponential Backoff) 的自动重试。
  - **错误码设计**：设计具备自解释性的错误响应（错误码 + 错误信息 + 指引），确保错误链路可追踪。
  - **全局错误归一化**：建立中心化的错误处理机制，返回一致的错误结构。
- **安全加固**：指导 JWT 验证流及基于角色 (RBAC) 的权限校验逻辑。
- **异步与观测**：
  - **后台任务**：使用任务队列处理非阻塞业务。
  - **结构化日志**：输出包含 Context（TraceID, UserID）的 JSON 日志。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| 逻辑架构 (Repo/Service) | `cc-skill-backend-patterns` | ✅ 整合 GitHub 优质后端模式 |
| 数据库 & 缓存优化 | `cc-skill-backend-patterns` | ✅ 整合 GitHub 优质后端模式 |
| 系统韧性 (Retry/Error) | `cc-skill-backend-patterns` | ✅ 整合 GitHub 优质后端模式 |
| 安全 & 运维模式 | `cc-skill-backend-patterns` | ✅ 整合 GitHub 优质后端模式 |

## 📚 参考资料 (References)

- [Awesome Antigravity Skills: Backend Patterns](https://github.com/sickn33/antigravity-awesome-skills/blob/main/skills/cc-skill-backend-patterns/SKILL.md)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout Standard](https://github.com/golang-standards/project-layout)

## 🧠 Usage

当以下情况发生时调用：

- 初始化新项目的基础代码布局时。
- 重构臃肿的模块或解决循环依赖问题时。
- 讨论微服务拆分逻辑或定义接口契约边界时。

## 📥 Input

- 业务领域概览。
- 当前代码结构现状（如果是重构）。
- 团队规模与维护预期。

## 📤 Output

- 建议的文件夹层次结构图。
- 核心层次（Layer）的交互时序或逻辑说明。
- 模块耦合度分析与建议。

## ⚠️ Constraints

- ✅ 优先考虑“高内聚低耦合”。
- ✅ **严禁 N+1 查询**：必须在审计中确认是否使用了批量获取。
- ✅ **严禁 SELECT ***：必须明确指定所需的字段。
- ✅ **依赖倒置**：严禁在领域层泄露基础设施（如数据库驱动、HTTP 框架）相关的逻辑。
- ✅ **防御性设计**：必须包含错误重试逻辑或明确的错误处理边界。
- ✅ **权限最小化**：所有敏感接口必须经过鉴权与权限校验。
- ✅ **结构化响应规范**：跨模块调用时必须遵循统一的 Error/Result 包装模式。
