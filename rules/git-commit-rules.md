# Git 提交规范（Git Commit Convention）

> 本文件定义了 Git 提交信息的格式规范，确保提交历史清晰可追溯。

---

## 提交格式

```text
<type>(<scope>): <subject>

<body>

<footer>
```

---

## Type（类型）

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(invoice): 添加 OFD 格式解析` |
| `fix` | 修复 Bug | `fix(email): 修复附件解析失败问题` |
| `docs` | 文档更新 | `docs(readme): 更新安装说明` |
| `style` | 代码格式（不影响逻辑） | `style: 格式化代码` |
| `refactor` | 重构（非新功能、非修复） | `refactor(client): 拆分连接逻辑` |
| `perf` | 性能优化 | `perf(parse): 优化 PDF 解析速度` |
| `test` | 测试相关 | `test(invoice): 添加发票解析单测` |
| `chore` | 构建/工具/依赖更新 | `chore: 升级 Go 版本至 1.24` |
| `revert` | 回滚提交 | `revert: 撤销 feat(xxx)` |
| `ci` | CI/CD 配置 | `ci: 添加 GitHub Actions` |
| `build` | 构建系统变更 | `build: 更新 wails 构建配置` |

---

## Scope（作用域）- 可选

标识影响的模块，常用值：

- `email` - 邮件模块
- `invoice` - 发票模块
- `platform` - 云端平台
- `folder` - 文件夹扫描
- `ui` - 前端界面
- `config` - 配置相关
- `api` - API 层

---

## Subject（简述）

- 使用中文或英文，保持一致
- 不超过 50 个字符
- 不以句号结尾
- 使用动词开头：添加、修复、优化、重构

---

## Body（正文）- 可选

- 解释 **为什么** 做这个改动
- 每行不超过 72 个字符

---

## Footer（脚注）- 可选

- 关联 Issue：`Closes #123`
- 破坏性变更：`BREAKING CHANGE: xxx`

---

## 示例

```text
feat(invoice): 添加 OFD 格式发票解析

- 支持读取 OFD 文件结构
- 提取发票号、金额、日期等字段
- 集成到发票处理流程

Closes #42
```

```text
fix(email): 修复 GBK 编码附件文件名乱码

部分邮件服务器返回 GBK 编码的附件名，
之前未正确处理导致乱码。

现在统一使用 charset 库解码。
```

---

## 禁止事项

- ❌ 无意义的提交信息：`update`、`fix bug`、`修改`
- ❌ 单次提交包含多个不相关改动
- ❌ 提交未测试的代码
