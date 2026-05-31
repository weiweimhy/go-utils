# Go-Utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.24-blue.svg)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/weiweimhy/go-utils/v5.svg)](https://pkg.go.dev/github.com/weiweimhy/go-utils/v5)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`go-utils` 是一个面向工程复用的 Go 工具库。v5 根模块只保留标准库优先、低副作用、低依赖的基础能力；日志、JWT、本地 KV、下载、HTML、EPUB 和外部平台集成迁入子模块，避免轻量使用者被重依赖影响。

## 核心特性

- **根模块瘦身**：主模块不再直接依赖 zap、bbolt、jwt、x/net、lumberjack 等第三方包
- **安全默认值**：提供安全写文件、原子写入、路径边界、响应大小限制和敏感信息脱敏
- **Context 克制使用**：网络、重试、并发任务等阻塞链路支持取消；纯工具函数保持简单
- **可组合小包**：JSON 文件、路径、文本、ID、时间、重试、文件元信息等能力按职责拆分
- **默认静默**：通用包不主动输出日志，必要时通过标准库接口注入

## 环境要求

- Go 1.24+

## 安装

```bash
go get github.com/weiweimhy/go-utils/v5
```

按需安装扩展子模块：

```bash
go get github.com/weiweimhy/go-utils/v5/download
go get github.com/weiweimhy/go-utils/v5/epub
go get github.com/weiweimhy/go-utils/v5/htmlutil
go get github.com/weiweimhy/go-utils/v5/jwt
go get github.com/weiweimhy/go-utils/v5/localdb
go get github.com/weiweimhy/go-utils/v5/logger
go get github.com/weiweimhy/go-utils/v5/mongo
go get github.com/weiweimhy/go-utils/v5/wechat
go get github.com/weiweimhy/go-utils/v5/tencentocr
```

从 v4.x 迁移到 v5：

```bash
go get github.com/weiweimhy/go-utils/v5@latest
go mod tidy
```

将 import 路径从 `/v4` 改为 `/v5`。如果使用 `logger`、`jwt`、`localdb`、`htmlutil`、`download`、`epub`，请改为对应子模块路径，例如：

```go
import "github.com/weiweimhy/go-utils/v5/jwt"
```

## 包速查

### 根模块核心包

| 包名 | 核心功能 | 依赖 |
| :--- | :--- | :--- |
| `auditlog` | JSONL 审计写入、脱敏 hook | 标准库 |
| `configcheck` | 配置文件基础安全检查 | 标准库 |
| `cryptoutil` | SHA-256、Base64 | 标准库 |
| `errs` | 跨项目通用错误 | 标准库 |
| `envutil` | 环境变量类型转换 | 标准库 |
| `event` | 轻量进程内事件总线 | 标准库 |
| `filemeta` | 文件大小、时间、可选 SHA-256 | 标准库 |
| `fsutil` | 文件/目录操作、安全写入、原子写入 | 标准库 |
| `httputil` | 默认超时 HTTP 客户端、大小限制、JSON helpers | 标准库 |
| `idutil` | 序列号、随机 Hex/Base64URL ID | 标准库 |
| `jsonfile` | 泛型 JSON 文件读写 | 标准库 |
| `maputil` | Sorted keys、clone、merge | 标准库 |
| `pathutil` | 路径边界、相对路径清理、轻量 pattern | 标准库 |
| `regexputil` | regexp 便捷函数 | 标准库 |
| `retry` | context-aware 重试 | 标准库 |
| `runtimeutil` | 运行时版本信息 | 标准库 |
| `securityutil` | 保守脱敏工具 | 标准库 |
| `sliceutil` | Chunk、Unique、Map、Filter | 标准库 |
| `syncutil` | OnceValueWithError、Semaphore | 标准库 |
| `task` | 标准库日志接口的 worker pool | 标准库 |
| `textutil` | 截断、单行摘要、控制字符清理 | 标准库 |
| `timeutil` | 可取消 sleep、Clock、时间辅助 | 标准库 |

### 扩展子模块

| 子模块 | 核心功能 |
| :--- | :--- |
| `download` | 并发下载管理器 |
| `epub` | EPUB 读取和批量 HTML 修改 |
| `htmlutil` | 基于 `golang.org/x/net/html` 的 HTML 文本提取 |
| `jwt` | HMAC/RSA JWT 生成、验证、刷新 |
| `localdb` | BBolt 本地 KV 存储 |
| `logger` | zap + lumberjack 日志便捷封装 |
| `mongo` | MongoDB 客户端封装 |
| `wechat` | 微信小程序 session API |
| `tencentocr` | 腾讯云 OCR API 封装 |

文档入口：

- `mongo`：[mongo/README.md](mongo/README.md)
- `wechat`：[wechat/README.md](wechat/README.md)
- `tencentocr`：[tencentocr/README.md](tencentocr/README.md)

## 快速上手

### 安全写文件

```go
import "github.com/weiweimhy/go-utils/v5/fsutil"

err := fsutil.SecureWriteFile("./data/secret.json", []byte(`{"ok":true}`))
```

### JSON 文件

```go
import "github.com/weiweimhy/go-utils/v5/jsonfile"

type Config struct {
    Port int `json:"port"`
}

cfg, err := jsonfile.Load("./config/app.json", Config{Port: 8080}, jsonfile.Options{
    MissingOK: true,
})
err = jsonfile.Save("./config/app.json", cfg, jsonfile.Options{AtomicSave: true})
```

### HTTP 大小限制

```go
import "github.com/weiweimhy/go-utils/v5/httputil"

data, err := httputil.GetBytes(ctx, "https://api.example.com/data", httputil.Options{
    MaxBytes: 2 << 20,
})
```

### 路径边界

```go
import "github.com/weiweimhy/go-utils/v5/pathutil"

rel, err := pathutil.CleanRelative(userInput)
if err != nil {
    return err
}
if !pathutil.IsWithin(filepath.Join(root, rel), root) {
    return fmt.Errorf("path escapes root")
}
```

### Worker Pool

```go
import "github.com/weiweimhy/go-utils/v5/task"

pool := task.NewWorkerPool(ctx, task.WithWorkerCount(4), task.WithBufferSize(16))
defer pool.Close(time.Second)

pool.SubmitFunc(func(ctx context.Context) {
    // work
})
```

## 设计原则

1. 根模块只放低依赖、高复用、无业务语义的小能力。
2. 默认值要安全，名称要清楚表达风险，例如 `SecureWriteFile`。
3. 错误尽量可判断，场景包错误定义在各自包内。
4. 通用包默认静默，日志通过接口注入。
5. 泛型只用于收益明确的场景，例如 `jsonfile.Load[T]`。

## 验证

根模块和所有当前子模块均使用 `go test ./...` 验证。

PowerShell 环境可能输出一条 PSReadLine 预测功能警告，不影响测试结果。

## License

[MIT License](LICENSE)
