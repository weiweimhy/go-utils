---
name: TypeScript 代码审查
description: 针对 TypeScript 代码的专家级审查，遵循 TS 5.x 特性、地道模式及现代 Web 开发最佳实践
---

# Skill: TypeScript Review

针对 TypeScript 代码进行深度审查，确保其类型安全、逻辑严密且符合地道的 TS 编程风格。

## 🎯 触发条件

当以下情况发生时启用：

- "审查这段 TypeScript 代码"
- "这段代码的类型推导是否正确？"
- "帮我优化一下这个泛型函数"
- "看看这段 TS 代码有没有收窄类型"

👉 自动启用本 Skill

## 🎯 Purpose

提供专家级的 TypeScript 代码审查，重点关注类型推导、泛型应用、类型收窄 (Narrowing) 及现代 TS 特性的正确使用。

## 🧩 Capabilities

- **类型系统深度审计**：
  - **严禁滥用 `any`**：强制审查是否可以使用 `unknown`, 泛型或更精确的接口替代 `any`。
  - **类型收窄 (Narrowing)**：审查是否利用了 `typeof`, `instanceof`, 判别式联合 (Discriminated Unions) 或类型谓词 (Type Predicates) 进行安全的类型转换。
  - **泛型应用**：评估泛型约束 (`extends`) 和默认值的使用，确保代码的可复用性与类型安全性。
- **现代特性审查 (TS 5.x)**：
  - **Const 型参数**：审查是否利用了 `const` 类型参数来保留字面量类型的精度。
  - **装饰器 (Decorators)**：针对新版标准装饰器的合法性与性能影响进行评估。
- **地道惯用法校验**：
  - **接口 vs 类型别名**：建议对公开 API 的对象形状优先使用 `interface`（支持扩展），对联合类型或工具类型使用 `type`。
  - **只读属性**：推荐在不应变更的数据结构中使用 `readonly` 关键字或 `Readonly<T>` 工具类型。
  - **可选链与空值合并**：确保正确使用 `?.` 和 `??` 替代繁琐的逻辑判断。
- **异步与副作用管理**：
  - **Promise 合规性**：确保异步函数具备正确的 `Promise<T>` 返回类型，并正确处理并发调用（如 `Promise.allSettled`）。
- **工程实践约束**：
  - **模块系统**：审查是否遵循 ESM 标准，避免使用过时的命名空间 (Namespaces) 或 `require`。

## 🧠 Usage

当以下情况发生时调用：

- 作为 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 流程的第二阶段，针对 TypeScript 代码进行深度评估。
- 在前端或 Node.js 环境中，对复杂的业务逻辑或工具库进行质量审计。
- 当开发者遇到复杂的“类型体操”问题，需要简化建议时。

## 📥 Input

- 待审查的 `.ts` 或 `.tsx` 文件。
- `tsconfig.json` 的关键配置（如 `strict` 模式是否开启）。

## 📤 Output

结构化的效果报告（参见 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 的标准格式），并增加：

- **Type Safety Score**: 针对代码类型覆盖率与严密性的主观评分。
- **Refactoring to Idiomatic TS**: 提供具体的代码示例，展示如何将 JavaScript 风格的代码转化为地道的 TypeScript。

- ✅ **保持类型简洁**：过于复杂的递归类型应考虑拆解或简化。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| 类型安全审计 | `google-typescript-style` | ✅ 行业标准安全规范 |
| 现代特性 (5.x) | `typescript-official-blog` | ✅ 语言最新特性集成 |
| 地道惯用法校验 | `airbnb-javascript-style` | ✅ 社区主流编程风格 |
| 类型收窄与泛型 | `typescript-handbook` | ✅ 官方类型体操规范 |

## 📚 参考资料 (References)

- [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html)
- [Airbnb JavaScript Style Guide (TS Version)](https://github.com/airbnb/javascript)
- [TypeScript 5.x Release Notes](https://devblogs.microsoft.com/typescript/)

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 基础审查哲学与流程。
- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md): 涉及 API 层面的设计审查。
- [backend-patterns](file:///e:/go/go-utils/.agent/skills/core/backend/backend-patterns/SKILL.md): 涉及后端架构层面的架构审计。
