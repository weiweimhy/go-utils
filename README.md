# go-utils

一个简洁、易用的 Go 语言工具库集合，提供常用的工具函数和封装。

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 特性

- 🚀 **简洁易用**：提供简单直观的 API，易于上手
- 📦 **模块化设计**：按功能拆分包，职责清晰
- 🔧 **生产就绪**：完善的错误处理，不吞掉错误
- 📝 **代码规范**：遵循 Go 最佳实践，代码简洁清晰
- ⚡ **高性能**：使用高性能库（如 `sonic`、`zap`）

## 安装

```bash
go get github.com/weiweimhy/go-utils/v2
```

## 快速开始

### 文件下载

```go
package main

import (
    "github.com/weiweimhy/go-utils/v2/httputil"
)

func main() {
    // 单文件下载（默认 60 秒超时）
    err := httputil.DownloadFile("https://example.com/file.pdf", "./file.pdf")
    if err != nil {
        panic(err)
    }

    // 自定义超时时间
    err = httputil.DownloadFileWithTimeout(
        "https://example.com/file.pdf",
        "./file.pdf",
        30*time.Second,
    )
}
```

### 批量下载

```go
package main

import (
    "time"
    "github.com/weiweimhy/go-utils/v2/httputil"
    "github.com/weiweimhy/go-utils/v2/task"
)

func main() {
    // 创建工作池（10 个 worker，100 个任务缓冲）
    pool := task.NewWorkerPool(10, 100)
    defer pool.Close(30 * time.Second)

    // 创建下载任务
    tasks := []*httputil.DownloadTask{
        httputil.NewDownloadTask("https://example.com/file1.pdf", "./file1.pdf", nil),
        httputil.NewDownloadTask("https://example.com/file2.pdf", "./file2.pdf", nil),
    }

    // 批量下载
    httputil.DownloadBatch(pool, tasks)
}
```

### 日志记录

```go
package main

import (
    "github.com/weiweimhy/go-utils/v2/logger"
    "go.uber.org/zap"
)

func main() {
    // 初始化日志（生产环境）
    logger.InitProduction()

    // 使用日志
    logger.L().Info("application started", zap.String("version", "v2.0.0"))
    logger.L().Error("something went wrong", zap.Error(err))
}
```

### 文件操作

```go
package main

import (
    "github.com/weiweimhy/go-utils/v2/filesystem"
)

func main() {
    // 检查文件是否存在
    if filesystem.IsFileExist("./file.txt") {
        // 文件存在
    }

    // 保存文件
    err := filesystem.SaveToFile("./data.txt", []byte("hello world"))
    if err != nil {
        panic(err)
    }

    // 获取文件 Base64 编码
    base64, err := filesystem.GetFileBase64("./file.txt")
    if err != nil {
        panic(err)
    }
}
```

## 包列表

### 核心工具包

#### `httputil` - HTTP 工具
- 单文件下载（支持自定义超时）
- 批量下载（基于 WorkerPool）
- HTTP 客户端封装
- GitHub API 工具

#### `filesystem` - 文件系统操作
- 文件/目录存在性检查
- 文件读写操作
- 目录创建
- Base64 编码/解码

#### `crypto` - 加密/编码工具
- SHA256 哈希计算
- Base64 编码/解码

#### `htmlutil` - HTML 处理工具
- 按标签提取文本
- DOM 解析和提取
- HTML 内容清理

#### `strutil` - 字符串工具
- 正则表达式工具函数

#### `runtime` - 运行时工具
- 版本信息获取

### 数据库和存储

#### `localDB` - 本地数据库
基于 `bbolt` 的本地键值数据库封装。

```go
import "github.com/weiweimhy/go-utils/v2/localDB"

db, err := localDB.InitLocalDB("./data.db")
if err != nil {
    panic(err)
}
defer db.Close()
```

#### `mongo` - MongoDB 客户端
MongoDB 数据库操作封装。

### 日志和监控

#### `logger` - 日志工具
基于 `zap` 的结构化日志封装，支持：
- 生产环境配置（文件轮转、压缩）
- 开发环境配置（彩色输出）
- 上下文日志（带 trace ID）
- 日志采样（高并发场景）

```go
import "github.com/weiweimhy/go-utils/v2/logger"

// 初始化
logger.InitProduction()  // 或 logger.InitDevelopment()

// 使用
logger.L().Info("message", zap.String("key", "value"))

// 上下文日志
ctx := logger.WithTraceID(context.Background())
logger.FromContext(ctx).Info("message")
```

### 并发工具

#### `task` - 任务池
统一的 `WorkerPool + Task` 模式实现。

```go
import "github.com/weiweimhy/go-utils/v2/task"

// 定义任务
type MyTask struct {
    data string
}

func (t *MyTask) Execute() {
    // 处理任务
    processData(t.data)
}

// 使用工作池
pool := task.NewWorkerPool(10, 100)
defer pool.Close(30 * time.Second)

pool.Submit(&MyTask{data: "hello"})
```

### 第三方服务集成

#### `OCR` - OCR 识别
腾讯云 OCR 服务封装。

#### `wechat` - 微信 API
微信相关 API 封装。

#### `epub` - EPUB 处理
EPUB 文件解压、修改和重新打包。

## 版本说明

### v2.0.0

**重大变更**：
- 模块路径更新为 `github.com/weiweimhy/go-utils/v2`
- 包拆分：`customUtils` 按功能拆分为独立包
- 错误处理改进：所有函数正确返回 error
- 超时配置：HTTP 下载支持自定义超时

**破坏性变更**：
- `htmlutil.ExtractTextByTagDOM` 等函数签名变更：`[]string` → `([]string, error)`
- `localDB.InitLocalDB` 现在返回 error
- `customUtils` 包已删除

## 从 v1.x 升级到 v2.x

### 为什么需要修改导入路径？

根据 [Go 模块版本管理规范](https://go.dev/doc/modules/major-version)，当主版本号 >= 2 时，模块路径必须包含版本后缀 `/v2`。这允许同一项目同时使用不同主版本。

### 升级步骤

#### 1. 更新 go.mod

在你的项目根目录下，运行：

```bash
# 方法一：直接指定版本（推荐）
go get github.com/weiweimhy/go-utils/v2@v2.0.0
go mod tidy

# 方法二：如果遇到缓存问题，先清理缓存
go clean -modcache
go get github.com/weiweimhy/go-utils/v2@v2.0.0
go mod tidy
```

或者手动编辑 `go.mod`：

```diff
require (
-    github.com/weiweimhy/go-utils v1.0.2
+    github.com/weiweimhy/go-utils/v2 v2.0.0
)
```

**注意**：如果使用 `@latest` 遇到问题，请使用 `@v2.0.0` 明确指定版本号。

#### 2. 更新所有导入路径

**全局替换导入路径**：

```bash
# 使用 sed (Linux/Mac)
find . -name "*.go" -type f -exec sed -i 's|github.com/weiweimhy/go-utils"|github.com/weiweimhy/go-utils/v2"|g' {} +

# 使用 PowerShell (Windows)
Get-ChildItem -Recurse -Filter "*.go" | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/weiweimhy/go-utils"', 'github.com/weiweimhy/go-utils/v2"' | Set-Content $_.FullName
}
```

**手动更新示例**：

```go
// ❌ 旧导入（v1.x）
import (
    "github.com/weiweimhy/go-utils/customUtils"
    "github.com/weiweimhy/go-utils/htmlUtils"
    "github.com/weiweimhy/go-utils/logger"
)

// ✅ 新导入（v2.x）
import (
    "github.com/weiweimhy/go-utils/v2/filesystem"
    "github.com/weiweimhy/go-utils/v2/htmlutil"
    "github.com/weiweimhy/go-utils/v2/logger"
)
```

#### 3. 处理包名变更

**`customUtils` 包拆分**：

```go
// ❌ v1.x
import "github.com/weiweimhy/go-utils/customUtils"

customUtils.SaveToFile(...)
customUtils.StringToHash(...)

// ✅ v2.x
import (
    "github.com/weiweimhy/go-utils/v2/filesystem"
    "github.com/weiweimhy/go-utils/v2/crypto"
)

filesystem.SaveToFile(...)
crypto.StringToHash(...)
```

**包名映射表**：

| v1.x 包名 | v2.x 包名 | 说明 |
|----------|----------|------|
| `customUtils` | `filesystem` | 文件系统操作 |
| `customUtils` | `crypto` | 加密/编码 |
| `customUtils` | `httputil` | HTTP 工具 |
| `customUtils` | `strutil` | 字符串工具 |
| `customUtils` | `runtime` | 运行时工具 |
| `htmlUtils` | `htmlutil` | HTML 处理（全小写） |
| `strings` | `strutil` | 避免与标准库冲突 |

#### 4. 处理 API 变更

**错误处理改进**：

```go
// ❌ v1.x - htmlutil 函数不返回 error
import "github.com/weiweimhy/go-utils/htmlUtils"

texts := htmlUtils.ExtractTextByTagDOM(html, "p")
// 如果解析失败，texts 可能为 nil 或空

// ✅ v2.x - 现在返回 error
import "github.com/weiweimhy/go-utils/v2/htmlutil"

texts, err := htmlutil.ExtractTextByTagDOM(html, "p")
if err != nil {
    // 处理错误
    log.Printf("failed to extract text: %v", err)
    return
}
```

**localDB 初始化**：

```go
// ❌ v1.x
import "github.com/weiweimhy/go-utils/localDB"

db := localDB.InitLocalDB("./data.db")
// 如果失败会 fatal，无法处理

// ✅ v2.x
import "github.com/weiweimhy/go-utils/v2/localDB"

db, err := localDB.InitLocalDB("./data.db")
if err != nil {
    log.Fatalf("failed to init local DB: %v", err)
}
defer db.Close()
```

**完整的迁移示例**：

```go
// v1.x 代码
package main

import (
    "github.com/weiweimhy/go-utils/customUtils"
    "github.com/weiweimhy/go-utils/htmlUtils"
    "github.com/weiweimhy/go-utils/localDB"
)

func main() {
    // 文件操作
    customUtils.SaveToFile("./data.txt", []byte("hello"))
    
    // HTML 提取
    texts := htmlUtils.ExtractTextByTagDOM(html, "p")
    
    // 数据库初始化
    db := localDB.InitLocalDB("./data.db")
    defer db.Close()
}

// v2.x 代码
package main

import (
    "log"
    "github.com/weiweimhy/go-utils/v2/filesystem"
    "github.com/weiweimhy/go-utils/v2/htmlutil"
    "github.com/weiweimhy/go-utils/v2/localDB"
)

func main() {
    // 文件操作（现在返回 error）
    err := filesystem.SaveToFile("./data.txt", []byte("hello"))
    if err != nil {
        log.Fatal(err)
    }
    
    // HTML 提取（现在返回 error）
    texts, err := htmlutil.ExtractTextByTagDOM(html, "p")
    if err != nil {
        log.Fatal(err)
    }
    
    // 数据库初始化（现在返回 error）
    db, err := localDB.InitLocalDB("./data.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

#### 5. 验证升级

```bash
# 清理并更新依赖
go mod tidy

# 编译项目
go build ./...

# 运行测试
go test ./...
```

### 常见问题

**Q: 升级后编译错误 "package not found"**

A: 确保所有导入路径都已更新为 `/v2` 后缀，并运行 `go mod tidy`。

**Q: 可以同时使用 v1 和 v2 吗？**

A: 可以。Go 模块允许同时导入不同主版本：

```go
import (
    v1 "github.com/weiweimhy/go-utils"
    v2 "github.com/weiweimhy/go-utils/v2"
)
```

**Q: 如何回退到 v1.x？**

A: 修改 `go.mod` 并运行 `go mod tidy`：

```go
require github.com/weiweimhy/go-utils v1.0.2
```

### 自动化升级脚本

可以使用以下脚本自动升级（需要根据实际情况调整）：

```bash
#!/bin/bash
# upgrade-to-v2.sh

# 1. 更新 go.mod
go get github.com/weiweimhy/go-utils/v2@latest

# 2. 替换导入路径
find . -name "*.go" -type f -exec sed -i 's|github.com/weiweimhy/go-utils"|github.com/weiweimhy/go-utils/v2"|g' {} +

# 3. 更新依赖
go mod tidy

# 4. 编译验证
go build ./...
```

## 项目规范

本项目遵循严格的代码规范，详见：
- [项目规则](project-rules.md) - 项目特定规则（最高优先级）
- [Go 项目规则](go-project-rules.md) - 通用 Go 项目规则

### 核心原则

1. **代码必须简洁**：作为公共库，代码应该清晰、简洁
2. **不需要无用的注释**：代码应该自解释，避免冗余注释
3. **对外方法简单易用**：提供简洁的 API 接口，提供合理的默认值
4. **不自己消化 error**：所有错误返回给调用方处理

## 依赖

主要依赖：
- `go.uber.org/zap` - 结构化日志
- `github.com/bytedance/sonic` - 高性能 JSON
- `go.etcd.io/bbolt` - 本地数据库
- `go.mongodb.org/mongo-driver` - MongoDB 驱动
- `golang.org/x/sync` - 并发工具

完整依赖列表请查看 [go.mod](go.mod)。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 作者

[weiweimhy](https://github.com/weiweimhy)

