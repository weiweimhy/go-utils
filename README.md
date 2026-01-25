# Go-Utils

`go-utils` 是一个工业级的通用 Go 语言封装库，旨在提供稳健、高性能且易于测试的基础组件。项目深度集成了 Context 传播、接口化设计、安全 HTTP 客户端以及高性能结构化日志。

## 🚀 核心特性

- **模块化设计**：功能拆分清晰（httputil, download, fsutil 等），按需导入。
- **全链路 Context**：原生支持生命周期管控与超时控制，防止资源泄漏。
- **Mock-Ready**：关键组件（Mongo, OCR）抽象为接口，极简 Mock 单元测试。
- **性能卓越**：集成 `zap` (日志), `sonic` (JSON), `bbolt` 等高性能底层组件。
- **内存优化**：Epub 模块采用流式/分块处理，轻松应对超大文件。

## 📦 安装

```bash
go get github.com/weiweimhy/go-utils
```

### 🆙 升级库
要升级到最新版本，请运行：
```bash
go get -u github.com/weiweimhy/go-utils
go mod tidy
```
若需升级到特定版本：
```bash
go get github.com/weiweimhy/go-utils@v3.1.2
```

---

## 🛠 常用包速查表

| 包名 | 核心功能 | 适用场景 |
| :--- | :--- | :--- |
| `logger` | 结构化日志、Trace 追踪 | 全局日志记录、函数耗时分析 |
| `httputil` | 安全 HTTP 客户端、GitHub 接口 | 远程 API 调用 |
| `fsutil` | 增强型文件/目录操作 | 读写文件、自动创建父目录 |
| `download` | 高并发下载管理器 | 批量文件下载、并发受控 |
| `cryptoutil` | 常用 Hash & Base64 | 数据校验、编码转换 |
| `mongo` | 接口化 MongoDB 客户端 | 数据库 CRUD |
| `epub` | 高性能 EPUB 解析修改 | 电子书处理 |
| `errs` | 统一错误映射 | 业务错误码规范化 |

---

## 📖 快速上手

### 1. 高性能日志 (logger)
```go
import "github.com/weiweimhy/go-utils/logger"

// 初始化（建议在 main 或 init 中执行一次）
logger.Init(
    logger.WithFilename("./logs/app.log"),
    logger.WithLevel(zap.InfoLevel),
)

// 使用：记录并追踪函数执行
func Process() {
    defer logger.Trace(logger.L(), "Process")()
    logger.L().Info("working", zap.String("id", "123"))
}
```

### 2. 数据库操作 (mongo)
```go
import "github.com/weiweimhy/go-utils/mongo"

// 创建满足接口的实例
client, _ := mongo.NewClient(ctx, mongo.Config{
    Uri: "mongodb://localhost:27017",
    DatabaseName: "app_db",
    OPTimeout: 5 * time.Second,
})

// 使用接口而非具体实现，方便 Mock
var user struct{ Name string }
client.FindOne(ctx, "users", bson.M{"id": 1}, &user)
```

### 3. 高并发下载 (download)
```go
import "github.com/weiweimhy/go-utils/download"

dm := download.NewDownloadManager(download.WithWorkers(5))
dm.Start(ctx)

// 任务式添加，自动并发执行
dm.Add("https://example.com/a.jpg", "./data/a.jpg")
dm.Wait()
```

### 4. 工具包示例 (fsutil/cryptoutil)
```go
import (
    "github.com/weiweimhy/go-utils/fsutil"
    "github.com/weiweimhy/go-utils/cryptoutil"
)

// 安全重命名：自动确保父目录存在
fsutil.SaveToFile("./out/config.json", data)

// 快速 SHA256
hash := cryptoutil.SHA256String("hello")
```

---

## 🛡 开发规范 (项目准则)

为保持工业级代码质量，接入项目请遵循：
1. **必需传 Context**：所有 IO 操作必须接收 `context` 参数以确保生命周期受控。
2. **禁止直接使用 `zap.L()`**：请通过 `logger.L()` 获取受控配置的实例。
3. **接口编程**：注入依赖时优先声明 `IMongoClient` 等接口，提升代码解耦度。

## 📈 版本管理
本项目遵循 [Semantic Versioning](https://semver.org/)。
- **主版本号 (Major)**: 包含不兼容的 API 变更。
- **次版本号 (Minor)**: 包含向后兼容的功能性新增。
- **修订号 (Patch)**: 包含向后兼容的问题修正。

发布新版本时，请确保：
1. 更新 `rules/rules.md` 中的 Go 版本和定位。
2. 运行 `go mod tidy`。
3. 执行 `git tag vX.Y.Z` 并推送到远端。

## 📄 License
MIT License
