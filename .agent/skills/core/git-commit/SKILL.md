---
name: Git 提交
description: 规范化提交代码到 Git 仓库，包括状态检查、文件暂存、生成符合规范的提交信息并推送
---

# Git 提交 Skill

规范化提交代码变更到 Git 仓库，遵循项目 Git 提交规范。

## 前置条件检查

在提交前，必须检查当前状态：

```powershell
# turbo
# 1. 查看当前工作目录状态
git status

# turbo
# 2. 查看具体变更内容
git diff --stat
```

## 提交类型说明

根据 `rules/git-commit-rules.md` 规范：

| 类型 | 说明 | 示例 |
| :--- | :--- | :--- |
| `feat` | 新功能 | `feat(jwt): 添加 JWT 认证模块` |
| `fix` | 修复 Bug | `fix(email): 修复附件解析失败问题` |
| `docs` | 文档更新 | `docs(readme): 更新安装说明` |
| `style` | 代码格式（不影响逻辑） | `style: 格式化代码` |
| `refactor` | 重构（非新功能、非修复） | `refactor(client): 拆分连接逻辑` |
| `perf` | 性能优化 | `perf(parse): 优化 PDF 解析速度` |
| `test` | 测试相关 | `test(invoice): 添加发票解析单测` |
| `chore` | 构建/工具/依赖更新 | `chore: 升级 Go 版本至 1.24` |
| `revert` | 回滚提交 | `revert: 撤销 feat(xxx)` |
| `ci` | CI/CD 配置 | `ci: 添加 GitHub Actions` |
| `build` | 构建系统变更 | `build: 更新构建配置` |

## 提交流程

### 步骤 1：暂存文件

根据变更类型选择暂存方式：

```powershell
# 暂存所有变更
git add -A

# 或者暂存指定文件
git add <file1> <file2>

# 或者交互式暂存
git add -p
```

### 步骤 2：确认暂存内容

```powershell
# turbo
# 查看已暂存的文件
git diff --cached --stat
```

### 步骤 3：生成提交信息

**询问用户**（如果用户未指定）：

1. 本次提交的类型是什么？（feat/fix/docs/...）
2. 影响的模块/作用域是什么？（可选）
3. 简短描述本次变更（不超过 50 字符）

**提交信息格式**：

```text
<type>(<scope>): <subject>

<body>

<footer>
```

**规范要求**：

- 使用中文
- subject 不超过 50 个字符
- 使用动词开头：添加、修复、优化、重构
- 不以句号结尾

### 步骤 4：执行提交

```powershell
# 提交变更（替换 <MESSAGE> 为实际的提交信息）
git commit -m "<MESSAGE>"

# 带有详细说明的提交
git commit -m "<SUBJECT>" -m "<BODY>"
```

### 步骤 5：推送到远程（可选）

```powershell
# turbo
# 查看当前分支
git branch --show-current

# 推送到远程
git push origin <branch>

# 或者直接推送当前分支
git push
```

### 步骤 6：验证提交

```powershell
# turbo
# 查看最近的提交
git log -3 --oneline
```

## 常用场景

### 场景 1：提交所有变更并推送

```powershell
# turbo-all
git add -A
git status
git commit -m "<TYPE>(<SCOPE>): <SUBJECT>"
git push
```

### 场景 2：仅提交部分文件

```powershell
git add <file1> <file2>
git commit -m "<TYPE>(<SCOPE>): <SUBJECT>"
```

### 场景 3：修改上一次提交

```powershell
# 修改提交信息
git commit --amend -m "<NEW_MESSAGE>"

# 补充遗漏的文件到上一次提交
git add <forgotten_file>
git commit --amend --no-edit
```

## 禁止事项

- ❌ 无意义的提交信息：`update`、`fix bug`、`修改`
- ❌ 单次提交包含多个不相关改动
- ❌ 提交未测试的代码
- ❌ 直接推送到 main/master 分支（除非确认）

## 问题处理

### 撤销暂存

```powershell
# 撤销所有暂存
git reset HEAD

# 撤销指定文件的暂存
git reset HEAD <file>
```

### 撤销上一次提交（保留修改）

```powershell
git reset --soft HEAD~1
```

### 查看提交差异

```powershell
# turbo
# 查看上一次提交的变更
git show --stat HEAD
```
