---
name: C# 代码审查
description: 针对 C# 代码的专家级审查，遵循 .NET 8/9、ASP.NET Core 最佳实践及现代 C# 12/13 地道语法
---

# Skill: CSharp Review

针对 C# 代码进行深度审查，确保其符合 .NET 官方编程准则，重点关注异步编程、资源管理、LINQ 效率及现代语法特性的应用。

## 🎯 触发条件

当以下情况发生时启用：

- "审查这段 C# 代码"
- "这段代码符合 .NET 8/9 最佳实践吗？"
- "检查这段异步代码是否有潜在死锁"
- "利用 C# 12/13 的新特性改进这段代码"

👉 自动启用本 Skill

## 🎯 Purpose

提供地道的 (Idiomatic) C# 代码审查，识别潜在的内存泄漏、并发死锁、性能瓶颈及违反 SOLID 原则的实现。

## 🧩 Capabilities

- **现代语法特性审查 (C# 12/13)**：
  - **主构造函数 (Primary Constructors)**：在类和结构中建议使用更简洁的构造函数语法。
  - **集合表达式 (Collection Expressions)**：推荐使用 `[]` 代替旧有的 `new[]` 或 `new List<T>` 初始化。
  - **切片与范围 (Index & Range)**：审查是否利用了更优雅的切片语法进行数组/集合操作。
- **异步编程与并发控制**：
  - **Task 规范**：确保异步方法后缀为 `Async`，并正确使用 `await`。
  - **死锁预警**：在高频调用的库代码中审查是否需要使用 `.ConfigureAwait(false)`（针对 Legacy .NET 或特定 UI 框架）。
  - **资源回收**：强制在 `IAsyncDisposable` 或 `IDisposable` 对象上使用 `await using` 或 `using` 声明。
- **性能与内存效率**：
  - **Span\\<T\\> 与 Memory\\<T\\>**：在高性能路径下审查是否利用了切片内存管理以减少分派开销。
  - **LINQ 优化**：识别低效的 LINQ 查询（如多次遍历同一 IQurable），建议在必要时使用 `ToList()` 或 `ToArray()`。
  - **String 优化**：优先使用字符串插值 ($"") 或 `StringBuilder`。
- **代码结构与命名标准**：
  - **命名规范**：遵循 `PascalCase` (类/方法/属性) 和 `camelCase` (私有字段/参数)。
  - **可为空引用类型 (NRT)**：强制审查是否启用了 `#nullable enable`，并正确处理了可能的 `null`。
- **架构与依赖注入**：
  - **依赖注入 (DI)**：审查是否通过构造函数注入依赖，避免使用 Service Locator 模式。

## 🧠 Usage

当以下情况发生时调用：

- 作为 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 流程的第二阶段，针对 C# (.NET) 代码进行深度评估。
- 使用分析器 (Analyzers) 捕获到架构规约违背时进行复审。
- 升级旧版 .NET 代码至现代 .NET 8/9 平台时。

## 📥 Input

- 待审查的 `.cs` 文件或 Diff 片段。
- 目标 .NET 版本 (如 .NET 6, .NET 8, .NET 9)。

## 📤 Output

结构化的效果报告（参见 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 的标准格式），并增加：

- **Modernization Suggestions**: 指出可以利用现代 C# 新特性进行简化的代码块。
- **Memory Allocation Profile**: 对潜在代码分派点（Allocations）的风险预警。

## ⚠️ Constraints

- ❌ **严禁忽略异步告警**：禁止使用 `.Result` 或 `.Wait()` 导致同步阻塞。
- ❌ **避免过度使用 `dynamic`**：除非涉及到与动态语言的交互。
- ✅ **优先考虑类型安全**：强类型优于弱类型。
- ✅ **保持简洁**：简洁的 Lambda 表达式优于冗长的匿名方法。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| 现代语法 (12/13) | `csharp-language-design` | ✅ 语言官方最新演进 |
| 异步编程规范 | `dotnet-runtime-guidelines` | ✅ 官方异步避坑指南 |
| 性能与内存优化 | `benchmark-dotnet-patterns` | ✅ 高性能开发实践 |
| 架构与 DI | `aspnetcore-best-practices` | ✅ 框架官方架构指南 |

## 📚 参考资料 (References)

- [Microsoft C# Coding Conventions](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions)
- [Official .NET 8/9 Performance Benchmarks](https://devblogs.microsoft.com/dotnet/performance-improvements-in-net-9/)
- [C# Language Reference (latest)](https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/proposals/latest)

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 基础审查哲学与流程。
