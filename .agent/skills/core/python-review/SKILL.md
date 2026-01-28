---
name: Python 代码审查
description: 针对 Python 代码的专家级审查，遵循 PEP 8、The Zen of Python 及 2024-2025 现代 Python 最佳实践
---

# Skill: Python Review

针对 Python 代码进行深度审查，确保其符合 [The Zen of Python](https://peps.python.org/pep-0020/) 哲学、[PEP 8](https://peps.python.org/pep-0008/) 规范及现代 Pythonic 编程准则。

## 🎯 触发条件

当以下情况发生时启用：

- "审查这段 Python 代码"
- "这段代码符合 PEP 8 吗？"
- "帮我重构成更 Pythonic 的写法"
- "检查这段异步代码是否有阻塞调用"

👉 自动启用本 Skill

## 🎯 Purpose

提供地道的 (Pythonic) 代码审查，重点关注可读性、简洁性、类型安全、资源管理及现代特性应用。

## 🧩 Capabilities

- **Pythonic 惯用法校验**：
  - **列表推导式与推导式**：审查是否合理使用 List/Dict/Set 推导式以增强简洁性，同时避免过度复杂的推导式。
  - **内置函数应用**：优先使用 `enumerate`, `zip`, `any`, `all`, `sorted` 等高效内置工具。
  - **解构赋值**：利用元组拆解、`*` 运算符进行优雅的变量赋值与参数整流。
- **现代特性审查 (3.10+)**：
  - **类型提示 (Type Hinting)**：确保公共 API 具备完善的类型标注，使用 Python 3.10+ 的联合类型语法 (`X | Y`)。
  - **Match-Case**：在复杂的条件分支逻辑中建议使用结构化模式匹配。
  - **Data Classes**：建议使用 `dataclasses` 替代笨重的自定义类。
- **可读性与命名 (PEP 8)**：
  - **命名约定**：强制执行 `snake_case` (变量/函数) 和 `PascalCase` (类)。
  - **文档字符串**：确保模块、类和函数具备 [PEP 257](https://peps.python.org/pep-0257/) 格式的 Docstrings。
- **资源与错误处理**：
  - **上下文管理器**：强制使用 `with` 语句管理文件、锁及网络连接等资源。
  - **异常处理**：严禁捕获过于宽泛的异常（如 `except Exception:`），提倡具体的 `Exception` 类型。
- **并发与异步**：
  - **AsyncIO 规范**：针对 `async/await` 路径下的阻塞调用进行预警，确保并发调用的正确性（如 `asyncio.gather`）。

## 🧠 Usage

当以下情况发生时调用：

- 作为 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 流程的第二阶段，针对 Python 代码进行深度评估。
- 提交 Python 项目 PR 前进行的代码质量检查。
- 重构旧版 Python (2.x/3.6) 代码至现代 Python 架构时。

## 📥 Input

- 待审查的 `.py` 文件或代码片段。
- 项目的 `pyproject.toml` 或 `requirements.txt`（可选，用于识别依赖环境）。

## 📤 Output

结构化的效果报告（参见 [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md) 的建议格式），特别增加：

- **Pythonic Refactoring**: 具体的代码重构对比，展示如何将“普通的”代码转化为“Pythonic”代码。
- **Linter Recommendations**: 建议使用的特定 Linter 规则（如 Black, Ruff, MyPy 指令）。

## ⚠️ Constraints

- ❌ **严禁使用过时的特性**：如 `format()` (除非必要) 优先使用 f-strings；避免 `old-style` 字符串拼接。
- ❌ **避免“非 Pythonic”倾向**：例如在 Python 中强制使用 Java 风格的 `get/set`。
- ✅ **优先考虑可读性**：Readability counts.
- ✅ **显式胜于隐式**：Explicit is better than implicit.

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| Pythonic 惯用法 | `the-zen-of-python` | ✅ 语言哲学核心准则 |
| 命名与风格规范 | `pep-8` | ✅ 官方代码风格指南 |
| 现代特性 (3.10+) | `python-release-notes` | ✅ 语言官方演进特性 |
| 类型提示与安全 | `google-python-style` | ✅ 工业界安全编程实践 |

## 📚 参考资料 (References)

- [The Zen of Python (PEP 20)](https://peps.python.org/pep-0020/)
- [PEP 8 – Style Guide for Python Code](https://peps.python.org/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [Ruff: Fast Python Linter Guidelines](https://docs.astral.sh/ruff/)

## 🔗 Related Skills

- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 基础审查哲学与流程。
