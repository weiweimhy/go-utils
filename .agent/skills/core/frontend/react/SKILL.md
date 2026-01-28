---
name: React 开发
description: 专注于 React 18/19、Hooks、Server Components 及现代状态管理的专家级开发能力
---

# Skill: React Development

本 Skill 专注于 React 生态的开发与架构指导，涵盖从客户端组件到服务器组件 (RSC) 的全栈思维转换。

## 🎯 触发条件

当以下情况发生时启用：

- "写一个 React 组件"
- "帮我优化一下这个 Hooks 的性能"
- "如何设计这个自定义 Hooks？"
- "在 React 项目中实现状态管理方案"

👉 自动启用本 Skill

## 🎯 Purpose

提供现代 React 开发的深度指导，强调 Hooks 的地道用法、并发模式下的性能优化以及现代框架（如 Next.js）的架构设计。

## 🧩 Capabilities

- **Atomic Design 架构引导**：指导将 UI 拆分为原子 (Atoms)、分子 (Molecules)、组织 (Organisms)、模板 (Templates) 和页面 (Pages)，提升组件的可复用性。
- **状态管理决策**：
  - **服务端状态 (TanStack Query)**：规范数据获取、缓存、自动同步及乐观更新逻辑。
  - **客户端状态 (Zustand)**：指导轻量级全局状态管理，强调微 Store 设计与 Selector 性能优化。
- **Hooks 地道开发模式**：
  - **逻辑与 UI 分离**：通过自定义 Hooks 封装极其复杂的业务逻辑，保持组件纯粹。
  - **复合组件模式 (Compound Components)**：解决深层 Props 传递问题，提供灵活的 API。
- **性能优化规范**：利用 React Compiler 或 `useMemo`/`useCallback` 减少非必要渲染，并利用 `Suspense` 与错误边界 (Error Boundaries) 提升韧性。
- **RSC 与全栈思维**：指导 Next.js/Remix 场景下服务器组件与客户端组件的合理分界。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| Atomic Design 架构 | `atomic-design-methodology` | ✅ 整合外部优质模式 |
| TanStack Query 规范 | `server-state-patterns` | ✅ 整合外部优质模式 |
| Zustand 微 Store 设计 | `zustand-best-practices` | ✅ 整合外部优质模式 |
| 复合组件模式 | `react-design-patterns` | ✅ 整合外部优质模式 |

## 📚 参考资料 (References)

- [Atomic Design by Brad Frost](https://atomicdesign.bradfrost.com/)
- [TanStack Query Documentation](https://tanstack.com/query/latest)
- [Zustand Patterns & Best Practices](https://docs.pmnd.rs/zustand/getting-started/introduction)

## 🧠 Usage

当以下情况发生时调用：

- 构建复杂的 React 组件树或自定义 Hooks 时。
- 决定项目在 RSC 与 Client Components 之间的架构分配时。
- 处理 React 合成事件、Refs 或与第三方 DOM 库交互时。

## 📥 Input

- 组件层级图或功能需求。
- 现有的 React JSX/TSX 代码。
- 目标 React 版本（18 或 19）。

## 📤 Output

- 高内聚低耦合的 React 组件定义 (TSX)。
- 健壮的自定义 Hooks 实现。
- 性能评估报告与优化建议。

## ⚠️ Constraints

- ✅ **Props 驱动 UI**：严禁在子组件内部直接修改外部传入的 Props。
- ✅ **逻辑隔离**：复杂的副作用 (Side Effects) 必须封装在自定义 Hooks 中。
- ✅ **非必要渲染预防**：必须使用性能分析工具确认重渲染边界是否合理。
- ✅ **类型完整性**：所有状态、Props 与 Actions 必须提供完整的 TS 定义。
- ✅ **禁止渲染循环副作用**：严禁在渲染过程中直接修改状态或产生可观测副作用。

## 🔗 Related Skills

- [typescript-review](file:///e:/go/go-utils/.agent/skills/core/typescript-review/SKILL.md): 涉及 TSX 中的类型安全审计。
- [backend-patterns](file:///e:/go/go-utils/.agent/skills/core/backend/backend-patterns/SKILL.md): 在全栈开发（如 Next.js）中涉及的架构模式。
