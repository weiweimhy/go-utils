---
name: Lua 代码审查
description: 针对 Lua 代码的专家级审查，遵循地道的 Lua 编程风格及其在不同环境（如 Love2D, Roblox）下的最佳实践
---

# Skill: Lua Review

针对 Lua 代码进行深度审查，确保其符合地道的 Lua 编程风格，重点关注作用域管理、内存效率及错误处理模式。

## 🎯 触发条件

当以下情况发生时启用：

- "审查这段 Lua 脚本"
- "这段代码有没有全局变量污染？"
- "检查这个元表 (Metatable) 逻辑"
- "给这个 Lua 模块提点优化建议"

👉 自动启用本 Skill

## 🎯 Purpose

提供高质量的 Lua 代码审查，识别潜在的全局污染、性能瓶颈、不当的 Metatable 使用以及违反社区准则的代码。

## 🧩 Capabilities

- **作用域与性能审查**：
  - **局部变量优先**：强制要求使用 `local` 变量，减少全局环境 (`_G`) 污染并提升访问速度。
  - **Upvalues 优化**：在高频调用的函数中，审查是否通过将全局函数/模块存入局部变量来优化性能。
- **地道惯用法校验**：
  - **真假值判定**：利用 Lua 中仅 `nil` 和 `false` 为假的特性，规范 `if value then` 的使用。
  - **伪三元运算**：审查 `and/or` 惯用法的正确性（注意 `and` 后的值不能为假）。
  - **模块化设计**：确保模块通过 `local M = {} ... return M` 模式定义，并使用 `require` 引入。
- **数据结构与 Metatables**：
  - **表 (Tables) 初始化**：审查表作为数组或字典的使用，避免混合使用导致的性能下降。
  - **Metamethods 规范**：审查 `__index`, `__newindex`, `__call` 等元方法的实现，防止过度“魔法”导致的调试困难。
- **错误处理模式**：
  - **多返回值约定**：审查是否遵循 `result, err_msg` 的返回模式。
  - **pcall/xpcall 应用**：在保护模式调用中，确保错误处理逻辑能够捕获并清晰记录异常。
- **环境特定准则**：
  - 针对游戏引擎（如 Love2D, Roblox/Luau）的特定生命周期（如 `update`, `draw`）及服务调用 (`GetService`) 进行规范性检查。

## 🧠 Usage

当以下情况发生时调用：

- 作为 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 流程的第二阶段，针对 Lua 代码进行深度评估。
- 自动化工具（如 LuaCheck）报告异常时进行人工复审。
- 涉及复杂脚本逻辑或核心库开发时。

## 📥 Input

- 待审查的 `.lua` 文件或 Diff 片段。
- 目标运行环境说明（如标准 Lua 5.1/5.4, Luau, LuaJIT）。

## 📤 Output

结构化的效果报告（参见 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 的标准格式），并增加：

- **Memory Efficiency Tips**: 针对 Lua GC 特性的内存分配优化建议。
- **Metatable Analysis**: 对复杂元表逻辑的安全性评估。

## ⚠️ Constraints

- ❌ **严禁滥用全局变量**：除非有明确的架构需求。
- ❌ **避免过深的表嵌套**：降低查找开销及维护难度。
- ✅ **优先考虑简洁性**：Lua 的核心价值在于 Simple & Small。
- ✅ **确保资源释放**：在涉及文件操作或 C-API 调用时，确保有关闭逻辑。
- ✅ **鼓励提问**：如果不确定代码的意图，先提问而非直接下结论。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| 作用域与全局管理 | `lua-style-guide` | ✅ 社区经典地道风格 |
| 并发与 LuaJIT 优化 | `luajit-wiki` | ✅ 高性能运行环境指南 |
| Roblox/Luau 规范 | `roblox-luau-style` | ✅ 现代 Luau 工业实践 |
| 错误处理模式 | `love2d-best-practices` | ✅ 游戏引擎实战模式 |

## 📚 参考资料 (References)

- [Programming in Lua (PiL)](https://www.lua.org/pil/contents.html)
- [Lua Style Guide (GitHub)](https://github.com/Olivine-Labs/lua-style-guide)
- [Roblox Luau Style Guide](https://roblox.github.io/lua-style-guide/)
- [LuaJIT Performance Guidelines](https://luajit.org/performance_lua.html)

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 基础审查哲学与流程。
