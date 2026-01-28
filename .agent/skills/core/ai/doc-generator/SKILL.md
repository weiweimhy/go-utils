---
name: 文档生成器
description: 自动生成项目文档，包括 API 接口文档、代码注释、README、Changelog 等
---

# Skill: Doc Generator

帮助自动生成高质量的项目文档，涵盖多种文档类型和格式。

## 🎯 Purpose

从代码和项目结构自动生成清晰、规范的文档，减少手动编写工作量，保持文档与代码同步。

## 🎯 触发条件

当用户说：

- "帮我生成 API 文档"
- "写一下这个函数的注释"
- "生成 README"
- "更新 Changelog"
- "生成接口说明"

👉 自动启用本 Skill

## 🧩 Capabilities

- **API 文档生成**：从代码生成 OpenAPI/Swagger 格式的接口文档
- **代码注释生成**：函数/类/模块级别的注释（JSDoc、GoDoc、docstring 等）
- **README 生成**：项目说明文档，包含 Quick Start、Features、Configuration
- **Changelog 生成**：版本变更日志，遵循 Keep a Changelog 规范
- **ADR 生成**：架构决策记录 (Architecture Decision Record)

## 🔄 文档类型与模板

### 1. API 文档结构

标准 API 文档应包含：

| 章节 | 内容 |
| :--- | :--- |
| **Introduction** | API 概述、Base URL、版本、联系方式 |
| **Authentication** | 认证方式、Token 管理、安全实践 |
| **Quick Start** | 快速开始示例 |
| **Endpoints** | 按资源组织的接口详情 |
| **Data Models** | Schema 定义、字段描述、校验规则 |
| **Error Handling** | 错误码参考、排错指南 |
| **Changelog** | API 版本历史 |

### 2. 接口文档模板

```markdown
## GET /users/:id

获取指定用户信息。

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| id | string | Yes | 用户 ID |

**Response:**

- 200: 用户对象
- 404: 用户不存在

**Example:**

[Request and response example]
```

### 3. README 模板

```markdown
# Project Name

一句话描述。

## Quick Start

[最小化运行步骤]

## Features

- Feature 1
- Feature 2

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | 服务端口 | 3000 |

## License

MIT
```

### 4. Code Comment Guidelines

| ✅ 应该注释 | ❌ 不应注释 |
| :--- | :--- |
| Why (业务逻辑) | What (显而易见的) |
| 复杂算法 | 每一行代码 |
| 非显而易见的行为 | 自解释的代码 |
| API 契约 | 实现细节 |

### 5. Changelog 模板 (Keep a Changelog)

```markdown
# Changelog

## [Unreleased]

### Added

- New feature

## [1.0.0] - 2025-01-01

### Added

- Initial release

### Changed

- Updated dependency

### Fixed

- Bug fix
```

## ⚙️ 生成流程

### API 文档生成流程

1. **分析代码结构**：识别路由、HTTP 方法、参数、响应
2. **生成接口描述**：为每个 endpoint 生成标准文档
3. **添加使用指南**：Quick Start、认证说明、最佳实践
4. **文档化错误处理**：列出所有错误码和解决方案
5. **创建示例**：cURL、JavaScript、Python 代码示例

### 代码注释生成流程

1. **分析函数签名**：参数类型、返回值
2. **推断函数用途**：从命名和实现推断
3. **生成规范注释**：按语言规范（JSDoc/GoDoc/docstring）
4. **补充边缘情况**：异常、限制、注意事项

## ✅ Best Practices

1. **保持同步**：文档应随代码一起更新
2. **示例优先**：每个接口都应有可运行的示例
3. **渐进式详情**：简单 → 复杂的信息层次
4. **可扫描**：使用表格、列表、标题便于快速查找

## ❌ Common Pitfalls

| 问题 | 解决方案 |
| :--- | :--- |
| 文档与代码不同步 | 在 CI 中添加文档检查 |
| 缺少错误文档 | 为每个错误码提供说明 |
| 示例无法运行 | 使用真实可测试的示例 |
| 参数要求不清晰 | 明确标注 required/optional 和校验规则 |

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| API 文档结构 | `api-documentation-generator` (外部) | ✅ 整合 5 步生成流程 |
| 文档模板系统 | `documentation-templates` (外部) | ✅ 整合 README/API/注释/Changelog 模板 |
| Changelog 规范 | `release` (本地) | ✅ 复用 Release Notes 生成规范 |
| AI-Friendly 文档 | `documentation-templates` (外部) | ✅ 整合 llms.txt 和 MCP-Ready 规范 |

## 📚 参考资料 (References)

- [Keep a Changelog](https://keepachangelog.com/) - Changelog 规范
- [OpenAPI Specification](https://swagger.io/specification/) - API 文档标准
- [JSDoc](https://jsdoc.app/) - JavaScript 文档注释
- [GoDoc](https://go.dev/blog/godoc) - Go 文档规范
- [antigravity-awesome-skills/api-documentation-generator](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/api-documentation-generator)
- [antigravity-awesome-skills/documentation-templates](https://github.com/sickn33/antigravity-awesome-skills/tree/main/skills/documentation-templates)

## 🔗 Related Skills

- [release](file:///e:/go/go-utils/.agent/skills/core/release/SKILL.md): Release Notes 生成
- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md): API 设计规范
- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 代码审查（含注释检查）
