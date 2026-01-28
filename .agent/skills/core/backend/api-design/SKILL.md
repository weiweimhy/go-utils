---
name: API 设计
description: 专注于 API 命名、RESTful 结构、返回格式及错误码设计的专项能力
---

# Skill: API Design

本 Skill 专注于后端 API 的外在接口设计，确保接口的一致性、易用性及符合行业标准的 RESTful 准则。

## 🎯 触发条件

当以下情况发生时启用：

- "设计一个 RESTful 接口"
- "帮我规范一下 API 路径"
- "这个接口的返回格式该怎么写？"
- "统一定义一下项目错误码"

👉 自动启用本 Skill

## 🎯 Purpose

提供专业的接口设计指导，减少前后端集成成本，建立清晰的资源导向型架构。

## 🧩 Capabilities

- **API 命名建议**：遵循资源导向命名规范（名词、复数形式、避免动词进入路径，如 `/api/v1/markets` 代替 `/api/v1/getMarkets`）。
- **资源建模规范**：合理设计资源层级（嵌套 vs 扁平化）。
- **RESTful 结构设计**：利用正确的 HTTP 动词 (GET, POST, PUT, PATCH, DELETE) 进行状态转换，并合理使用 HTTP 状态码。
- **查询参数标准化**：制定统一的过滤 (filtering)、排序 (sorting) 及分页 (pagination) 参数规范（如 `?limit=20&offset=100`）。
- **返回格式规范**：统一的 JSON 响应封装，包含数据信封、元数据 (Meta) 及分页对象。
- **错误码设计**：设计具备自解释性的错误响应（错误码 + 错误信息 + 指引），确保错误链路可追踪。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 类型 |
| :--- | :--- | :--- |
| RESTful 准则 | `cc-skill-backend-patterns` | ✅ 整合 GitHub 优质模式 |
| 分页与查询标准化 | `google-api-design-guide` | ✅ 参考行业顶级实践 |
| 响应格式封装 | `json-api-spec` | ✅ 参考行业标准规范 |

## 📚 参考资料 (References)

- [Google API Design Guide](https://cloud.google.com/apis/design)
- [RESTful Web Services Best Practices](https://github.com/Microsoft/api-guidelines)

## 🧠 Usage

当以下情况发生时调用：

- 设计新功能或模块的对外接口时。
- 优化现有 API 的路径结构或参数设计时。
- 定义全局统一的响应与错误处理规范时。

## 📥 Input

- 业务资源模型描述。
- 交互流（Action Flow）。

## 📤 Output

- 详细的 API 路径结构图。
- 响应报文与错误报文的示例。
- API 变更（Breaking Changes）的影响分析。

## ⚠️ Constraints

- ✅ 优先考虑开发者的集成体验。
- ✅ 强制执行资源导向 (Resource-Oriented) 准则。
- ✅ **命名一致性**：整个 API 集合必须使用统一的命名风格（推荐 snake_case）。
- ✅ **分页必选**：对于可能返回列表的接口，必须设计分页参数。
- ✅ 错误码必须全局唯一且有明确分类。
- ✅ **破坏性变更显式化**：任何 Breaking Change 必须在变更描述中明确标记。
