---
name: 架构顾问
description: 系统级架构咨询能力，提供架构决策支持、技术选型分析及 ADR 文档生成
---

# Skill: Architecture Consultant

提供系统级架构咨询能力，帮助进行架构决策、技术选型、系统设计及架构决策记录 (ADR) 生成。

## 🎯 触发条件

当以下情况发生时启用：

- "帮我设计项目架构"
- "这个功能该用什么技术栈？"
- "帮我写一个架构决策记录 (ADR)"
- "对比一下几个技术方案的优劣"

👉 自动启用本 Skill

## 🎯 Purpose

作为"设计时"架构顾问，本 Skill 专注于系统架构的顶层设计与决策过程，确保架构选型有据可依、权衡透明、决策可追溯。

## 🧩 Capabilities

- **架构决策框架**：
  - 需求分析 → 约束识别 → 权衡评估 → 决策记录
  - 遵循 "Simplicity is the ultimate sophistication" 原则
  - 每个决策必须有明确的权衡分析

- **ADR (Architecture Decision Record) 生成**：
  - 标准化 ADR 模板
  - 记录决策背景、选项对比、最终选择及理由
  - 支持后续追溯与变更管理

- **系统设计模式指导**：
  - **Clean Architecture**: 领域与基础设施分离
  - **DDD (Domain-Driven Design)**: 领域建模与限界上下文
  - **微服务架构**: 服务边界与通信模式
  - **CQRS/Event Sourcing**: 读写分离与事件驱动
  - **Modular Monolith**: 渐进式拆分策略

- **技术选型与权衡分析**：
  - 技术栈对比矩阵
  - 团队能力匹配评估
  - 长期维护成本分析
  - 供应商锁定风险评估

- **架构可视化**：
  - 使用 Mermaid 生成架构图
  - C4 模型 (Context, Container, Component, Code)
  - 数据流图与时序图

- **反模式识别与改进**：
  - NIH (Not Invented Here) 综合症
  - 大泥球 (Big Ball of Mud)
  - 过度设计 (Over-Engineering)
  - 循环依赖与层级泄露

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| 架构决策框架 | `architecture` (外部) | ✅ 整合 GitHub 优质技能 |
| ADR 模板 | `architecture` (外部) | ✅ 整合标准化决策记录模板 |
| 系统设计模式 | `software-architecture` (外部) + `backend-patterns` (本地) | 🔁 组合 Clean Arch/DDD 指导 |
| 技术选型分析 | `senior-architect` (外部) | ✅ 整合 Tech Decision Guide |
| 架构可视化 | `senior-architect` (外部) | ✅ 整合 Architecture Diagram Generator |
| 反模式识别 | `software-architecture` (外部) | ✅ 整合 Anti-Patterns 识别 |

## 📚 参考资料 (References)

- [Antigravity Awesome Skills: Architecture](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/architecture)
- [Antigravity Awesome Skills: Software Architecture](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/software-architecture)
- [Antigravity Awesome Skills: Senior Architect](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/senior-architect)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [ADR GitHub Organization](https://adr.github.io/)
- [C4 Model](https://c4model.com/)

## 🧠 Usage

当以下情况发生时调用：

- 启动新项目需要确定技术架构时
- 面临重大技术选型决策时
- 需要记录架构决策并获得团队共识时
- 评估现有架构是否需要重构时
- 讨论微服务拆分或系统边界时

## 📥 Input

- 业务需求与约束条件
- 现有系统架构现状（如适用）
- 团队规模与技术能力
- 非功能性需求（性能、可用性、安全性）
- 预算与时间约束

## 📤 Output

- **ADR 文档**：标准化架构决策记录
- **架构图**：Mermaid/C4 格式的可视化图表
- **技术选型报告**：权衡矩阵与推荐方案
- **风险评估**：潜在问题与缓解策略

## ⚠️ Constraints

- ✅ **先简单后复杂**：优先选择最简单的满足需求的方案
- ✅ **决策必须有理有据**：每个选择都需要权衡分析
- ✅ **避免过度设计**：不为未来可能不会发生的需求设计
- ✅ **团队能力匹配**：架构复杂度不能超出团队驾驭能力
- ❌ **禁止凭直觉决策**：所有决策必须有数据或逻辑支撑

## 🔗 Related Skills

- `backend-patterns`：后端运行时架构模式
- `api-design`：API 接口设计规范
- `doc-generator`：生成架构相关文档
- `product-manager`：需求澄清与任务拆解
