---
name: Vue 开发
description: 专注于 Vue 3、Composition API、Pinia 及 Vite 生态的专家级开发能力
---

# Skill: Vue Development

本 Skill 专注于 Vue 生态的开发与架构指导，确保项目遵循现代 Vue 3 最佳实践，具备高性能与高可维护性。

## 🎯 触发条件

当以下情况发生时启用：

- "写一个 Vue 3 组件"
- "帮我重构成 Composition API 写法"
- "如何设计这个 Composable (useXxx)？"
- "优化这个 Vue 应用的响应式性能"

👉 自动启用本 Skill

## 🎯 Purpose

提供地道的 Vue 3 开发建议，特别是在 Composition API 模式下的逻辑提取、响应式优化及企业级状态管理。

## 🧩 Capabilities

- **高级 Composable 模式**：
  - **单一职责 SRP**：指导编写功能单一、可测试的逻辑函数。
  - **服务式 Composable**：利用组合式 API 构建轻量级的业务服务层。
- **Pinia 模块化架构**：指导去中心化 Store 设计，强调一个 Store 仅负责一个领域/功能，并深度利用 TypeScript 类型推断。
- **响应式性能极限优化**：
  - **浅层响应式 (shallowRef)**：处理大型对象或数组时的性能加速。
  - **内置指令优化**：合理使用 `v-memo`, `v-once` 及 `markRaw` 减少计算压力。
- **工程化与可观测性**：
  - **异步组件**：利用 `defineAsyncComponent` 实现精准的包体积控制。
  - **组合式逻辑审计**：严禁在 Composable 之外维护跨组件共享的非响应式状态。
- **RSC 与全栈思维**：指导 Next.js/Remix 场景下服务器组件与客户端组件的合理分界。
- **UI 组件库集成**：指导 shadcn/vue 或 Nuxt UI 等现代组件库的模式应用。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| 高级 Composable 模式 | `vue-best-practices` | ✅ 整合外部优质模式 |
| Pinia 模块化架构 | `pinia-patterns` | ✅ 整合外部优质模式 |
| 响应式性能优化 | `vue-perf-handbook` | ✅ 整合外部优质模式 |
| 异步组件工程化 | `vite-bundling-guide` | ✅ 整合外部优质模式 |

## 📚 参考资料 (References)

- [Vue 3 Composition API Best Practices](https://github.com/vuejs/core)
- [Mastering Pinia Patterns](https://masteringpinia.com/)
- [Nuxt UI & Vue Patterns](https://nuxt.com/docs/guide/concepts/typescript)

## 🧠 Usage

当以下情况发生时调用：

- 使用 Vue 3 初始化新功能或重构旧的 Options API 代码时。
- 需要设计跨组件共享的业务逻辑（Composables）时。
- 优化前端性能瓶颈（如长列表渲染、复杂表单状态）时。

## 📥 Input

- 业务 UI 交互原型。
- 现有的 Vue 组件代码段。
- 预期的技术栈（Vue 3 + Vite + Pinia/Vuex）。

## 📤 Output

- 结构严密的 `.vue` 组件代码示例。
- 逻辑清晰的 `useXxx` Composable 实现。
- Pinia Store 的定义规范。

## ⚠️ Constraints

- ✅ **组合式 API 优先**：强制使用 `<script setup>` 语法及 Composition API。
- ✅ **状态单流向**：严禁通过 `parent` 或 `root` 直接修改状态。
- ✅ **禁止 Mixins**：严禁在新代码中使用 Mixins，必须使用 Composables 替代。
- ✅ **内存泄漏预防**：所有在 Composable 中开启的底层监听或长连接必须提供销毁钩子。
- ✅ **SSR 友好**：严禁在非客户端生命周期（如 Setup）中访问 `window` 或 `document`。

## 🔗 Related Skills

- [typescript-review](file:///e:/go/go-utils/.agent/skills/core/typescript-review/SKILL.md): 涉及 Vue 中的 TS 类型审计。
- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md): 涉及前端与后端接口的集成规范。
