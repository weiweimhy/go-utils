---
name: 后端校验与 Lint
description: 专注于后端参数设计、接口一致性以及配置文件格式的校验能力
---

# Skill: Validation Lint

本 Skill 专注于后端系统的契约与配置质量，确保参数设计合理、接口对外表现一致，并且各种配置文件（Config）满足语法与业务逻辑约束。

## 🎯 触发条件

当以下情况发生时启用：

- "定义一个请求参数结构体"
- "校验一下这个 API 的参数设计"
- "检查配置文件格式是否正确"
- "为这个 DTO 添加校验规则"

👉 自动启用本 Skill

## 🎯 Purpose

通过自动化的 Lint 与校验思维，减少因参数理解歧义、接口不规范或配置错误导致的线上问题。

## 🧩 Capabilities

- **校验参数设计**：
  - **命名规范**：检查请求参数是否符合项目定义的命名风格（如 `snake_case` 或 `camelCase`）。
  - **类型合理性**：验证字段类型是否最优化（例如：UUID 使用 string，金额使用 long/decimal 而非 float）。
  - **约束完备性**：检查是否定义了必要的校验注解（如 `required`, `length`, `range`, `pattern`）。
- **校验接口一致性**：
  - **路径与动词**：确保 API 路径结构符合 RESTful 规范（联动 `api-design` 技能）。
  - **响应结构**：验证所有接口返回的 JSON 结构是否遵循统一的数据信封（Data Envelope）模式。
  - **错误语义**：检查错误返回是否包含预定义的错误码（Error Code）及其对应的可读信息。
- **校验配置格式 (Config Checker)**：
  - **语法校验**：静态检查 YAML, JSON, Toml 等配置文件的基本语法。
  - **模式校验 (Schema Validation)**：根据预定义的 Schema 验证配置项的键值对是否合法。
  - **环境一致性**：识别并校验不同环境（Dev/Staging/Prod）之间的关键配置项差异是否合理。

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| 参数 Lint | `lint-and-validate` | ✅ 参考经典 API 校验哲学 |
| 接口一致性 | `api-design` | ✅ 继承项目现有的接口规范 |
| 配置检查 | `config-checker` | ✅ 整合配置语义化检查能力 |

## 📚 参考资料 (References)

- [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html)
- [Google API Improvement Proposals (AIPs)](https://google.aip.dev/)
- [JSON Schema Standard](https://json-schema.org/)

## 🧠 Usage

当以下情况发生时调用：

- 正在定义新的 DTO (Data Transfer Object) 或请求/响应结构时。
- 审核现有 API 的命名或结构变更时。
- 修改 `.yaml`, `.json` 等全局配置文件或环境变量模板时。

## 📥 Input

- API 定义文档（Swagger/OpenAPI/Markdown）。
- 代码中的结构体或参数定义。
- 配置文件内容。

## 📤 Output

- 发现的不一致项或潜在风险列表。
- 建议的修改方案。
- 校验报告（包含通过项与待修复项）。

## ⚠️ Constraints

- ✅ **安全性优先**：所有输入参数必须有明确的边界定义。
- ✅ **一致性强制**：新接口必须对齐现有的错误码体系。
- ✅ **配置零容忍**：配置文件中的语法错误必须作为阻断性问题（Blocker）指出。
- ✅ **联动 API 设计**：任何违反 `api-design` 规范的行为都应被记录。

## 🔗 Related Skills

- [api-design](file:///e:/go/go-utils/.agent/skills/core/backend/api-design/SKILL.md)
- [backend-patterns](file:///e:/go/go-utils/.agent/skills/core/backend/backend-patterns/SKILL.md)
