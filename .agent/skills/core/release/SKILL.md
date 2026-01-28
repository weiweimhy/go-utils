---
name: 发布版本
description: 发布新版本到 GitHub，包括版本号选择、变更日志生成、Git 标签创建和推送
---

# 发布版本 Skill

发布 go-utils 库的新版本到 GitHub，遵循语义化版本规范。

## 前置条件检查

在发布前，必须确认以下条件：

```powershell
# turbo
# 1. 确保工作目录干净（无未提交的修改）
git status

# turbo
# 2. 确保所有测试通过
go test ./...

# turbo
# 3. 确保代码可以正常构建
go build ./...

# turbo
# 4. 确保 go.mod 整洁
go mod tidy
```

## 版本号确定

### 查看当前版本

```powershell
# turbo
git tag --list --sort=-v:refname | Select-Object -First 5
```

### 语义化版本规则

根据 [Semantic Versioning](https://semver.org/)：

| 版本类型 | 何时使用 | 示例 |
| :--- | :--- | :--- |
| **Major** (X.0.0) | 不兼容的 API 变更 | v3.0.0 → v4.0.0 |
| **Minor** (x.Y.0) | 向后兼容的功能新增 | v3.2.0 → v3.3.0 |
| **Patch** (x.y.Z) | 向后兼容的问题修正 | v3.2.0 → v3.2.1 |

### 决定新版本号

**询问用户**：

1. 本次发布包含哪些改动？（新功能、Bug 修复、破坏性变更）
2. 建议的新版本号是什么？

## 发布流程

### 步骤 1：生成变更日志

```powershell
# turbo
# 查看自上个版本以来的提交（替换 <LAST_TAG> 为实际的上个版本标签）
git log <LAST_TAG>..HEAD --oneline --no-decorate
```

### 步骤 2：确认 go.mod 版本路径

对于 v2+ 版本，确保 `go.mod` 中的 module 路径包含版本后缀：

```go
// v3.x 版本应该是：
module github.com/weiweimhy/go-utils/v3
```

### 步骤 3：创建 Git 标签

```powershell
# 创建带注释的标签（替换 <VERSION> 和 <MESSAGE>）
git tag -a <VERSION> -m "<MESSAGE>"

# 示例：
# git tag -a v3.3.0 -m "feat: 添加 JWT 认证模块"
```

### 步骤 4：推送标签

```powershell
# 推送标签到远程仓库
git push origin <VERSION>
```

### 步骤 5：验证发布

```powershell
# turbo
# 确认标签已推送
git ls-remote --tags origin | Select-String "<VERSION>"
```

等待几分钟后，在 [pkg.go.dev](https://pkg.go.dev/github.com/weiweimhy/go-utils/v3) 验证新版本是否可用。

## 发布后清理

如果需要强制 Go 模块代理更新缓存：

```powershell
# turbo
# 请求代理更新（替换 <VERSION>）
$env:GOPROXY="https://proxy.golang.org,direct"; go list -m github.com/weiweimhy/go-utils/v3@<VERSION>
```

## 常见问题

### 删除错误的标签

```powershell
# 删除本地标签
git tag -d <VERSION>

# 删除远程标签
git push origin --delete <VERSION>
```

### 版本号格式

- ✅ 正确：`v3.2.0`、`v3.2.1`
- ❌ 错误：`3.2.0`、`V3.2.0`、`v3.2`

## 快速发布命令汇总

```powershell
# 完整发布流程（替换变量后执行）
$VERSION = "v3.3.0"
$MESSAGE = "feat: 添加新功能描述"

git status
go test ./...
go build ./...
go mod tidy
git tag -a $VERSION -m $MESSAGE
git push origin $VERSION
```
