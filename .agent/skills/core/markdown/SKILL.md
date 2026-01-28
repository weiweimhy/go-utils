---
name: Markdown 规范
description: Markdown 编写规范与 Lint 规则，确保文档格式一致、可读性强
---

# Skill: Markdown

Markdown 编写规范和最佳实践，确保所有 Markdown 文件符合 markdownlint 规范。

## 🎯 Purpose

确保项目中所有 Markdown 文件格式统一、结构清晰、符合 markdownlint 规范，提升文档可读性和可维护性。

## 🎯 触发条件

当用户说：

- "检查 markdown 格式"
- "这个 md 文件有什么问题"
- "帮我格式化 markdown"
- "表格/代码块怎么写"

👉 自动启用本 Skill

## 🧩 Capabilities

- **Markdownlint 规则**：核心规则解读与应用
- **文档结构规范**：标题层级、段落组织
- **组件规范**：表格、代码块、链接、列表的正确写法
- **格式化最佳实践**：空行、缩进、换行规则

## 📏 Markdownlint 核心规则

### 标题规则

| 规则 | 说明 | 示例 |
| :--- | :--- | :--- |
| MD001 | 标题层级只能递增 1 级 | `# → ## → ###` ✅ `# → ###` ❌ |
| MD003 | 标题风格统一 | 使用 ATX 风格 `#` 而非 Setext |
| MD018 | `#` 后必须有空格 | `# Title` ✅ `#Title` ❌ |
| MD022 | 标题前后需有空行 | 标题与正文间空一行 |
| MD024 | 避免重复标题 | 同级标题不能重名 |
| MD025 | 文档只能有一个 H1 | 顶级标题唯一 |
| MD041 | 首行应为 H1 标题 | 文件以 `#` 开头 |

### 空白与格式规则

| 规则 | 说明 | 示例 |
| :--- | :--- | :--- |
| MD009 | 行尾不能有空格 | 除非用于换行 |
| MD010 | 不能使用 Tab 缩进 | 使用空格代替 |
| MD012 | 连续空行最多 1 行 | 避免多个空行 |
| MD047 | 文件末尾需有空行 | 确保以换行符结尾 |

### 列表规则

| 规则 | 说明 | 示例 |
| :--- | :--- | :--- |
| MD004 | 无序列表符号统一 | 全部使用 `-` |
| MD005 | 列表缩进一致 | 同级列表对齐 |
| MD007 | 列表缩进使用 2 空格 | 默认 2 空格 |
| MD030 | 列表符号后有空格 | `- item` ✅ `-item` ❌ |
| MD032 | 列表前后需有空行 | 与正文分隔 |

### 代码块规则

| 规则 | 说明 | 示例 |
| :--- | :--- | :--- |
| MD014 | 命令行不需要 `$` | 除非展示输出 |
| MD031 | 代码块前后需有空行 | 与正文分隔 |
| MD040 | 代码块需指定语言 | \`\`\`go 而非 \`\`\` |
| MD046 | 代码块风格统一 | 使用 fenced 风格 |

### 链接与图片规则

| 规则 | 说明 | 示例 |
| :--- | :--- | :--- |
| MD034 | URL 应使用链接语法 | `[link](url)` 而非裸 URL |
| MD042 | 链接不能为空 | `[text]()` ❌ |
| MD045 | 图片需有 alt 文本 | `![alt](img.png)` |

## 📝 组件规范写法

### 表格

```markdown
| Column 1 | Column 2 | Column 3 |
| :--- | :--- | :--- |
| Left align | Left align | Left align |
```

> [!TIP]
> 使用 `:---` 左对齐，`---:` 右对齐，`:---:` 居中

### 代码块

```markdown
\`\`\`go
func main() {
    fmt.Println("Hello")
}
\`\`\`
```

> [!IMPORTANT]
> 始终指定语言标识符

### 链接

```markdown
[显示文本](https://example.com)
[相对路径](./docs/api.md)
[锚点链接](#section-name)
```

### 列表

```markdown
- 无序列表项 1
- 无序列表项 2
  - 嵌套项（2 空格缩进）

1. 有序列表项 1
2. 有序列表项 2
```

## ✅ Best Practices

1. **单一 H1**：每个文档只有一个顶级标题
2. **标题层级**：严格按 1 → 2 → 3 递增，不跳级
3. **空行分隔**：标题、代码块、列表前后各留一空行
4. **一致风格**：全文统一使用 `-` 作为无序列表符号
5. **代码语言**：所有代码块都指定语言
6. **链接格式**：避免裸 URL，使用 `[text](url)` 格式
7. **行长控制**：建议每行不超过 80-120 字符

## ❌ Common Mistakes

| 错误 | 正确 |
| :--- | :--- |
| `###` 直接跟在 `#` 后 | `#` → `##` → `###` |
| `#Title` 无空格 | `# Title` |
| 代码块不指定语言 | \`\`\`python |
| 行尾有多余空格 | 删除尾随空格 |
| 连续多个空行 | 最多一个空行 |
| 裸 URL | 使用链接语法包裹 |

## 🔍 能力溯源 (Source Mapping)

| 能力 | 来源 | 说明 |
| :--- | :--- | :--- |
| Lint 规则 | `markdownlint v0.38.0` | ✅ 官方规范 |
| 结构最佳实践 | `documentation-templates` (外部) | ✅ 整合文档结构原则 |
| 组件规范 | 新建 | ✅ 基于 CommonMark 规范 |

## 📚 参考资料 (References)

- [markdownlint Rules](https://github.com/DavidAnson/markdownlint/tree/v0.38.0/doc) - 官方规则文档
- [CommonMark Spec](https://spec.commonmark.org/) - Markdown 标准规范
- [GitHub Flavored Markdown](https://github.github.com/gfm/) - GitHub 扩展规范

## 🔗 Related Skills

- [doc-generator](file:///e:/go/go-utils/.agent/skills/core/ai/doc-generator/SKILL.md): 文档生成时应遵循本规范
- [code-review](file:///e:/go/go-utils/.agent/skills/core/code-review/SKILL.md): 代码审查包含文档质量检查
