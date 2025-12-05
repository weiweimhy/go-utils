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

**迁移指南**：

```go
// 旧导入（v1.x）
import "github.com/weiweimhy/go-utils/customUtils"

// 新导入（v2.x）
import "github.com/weiweimhy/go-utils/v2/filesystem"
import "github.com/weiweimhy/go-utils/v2/httputil"
// ...
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

