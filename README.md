# Go-Utils

`go-utils` 是一个工业级的通用 Go 语言封装库，旨在提供稳健、高性能且易于测试的基础组件。项目深度集成了 Context 传播、接口化设计、安全 HTTP 客户端以及高性能结构化日志。

## 🚀 核心特性

- **模块化设计**：按功能深度拆分（httputil, download, fsutil, etc.），结构清晰。
- **全链路 Context**：原生支持生命周期管控，防止资源泄漏。
- **Mock-Ready**：关键组件（Mongo, OCR）抽象为接口，方便业务层进行单元测试。
- **内存优化**：Epub 模块采用流式/分块处理，支持超大文件。
- **工业级安全**：锁定的 HTTP 客户端，内置合理的超时与连接池配置。

## 📦 安装

```bash
go get github.com/weiweimhy/go-utils
```

---

## 🛠 组件说明

### 1. 高性能日志 (logger)
支持 Functional Options 配置模式，内置采样和 Trace 功能。

```go
import "github.com/weiweimhy/go-utils/logger"

// 1. 初始化
logger.Init(
    logger.WithLevel(zap.InfoLevel),
    logger.WithFilename("./logs/service.log"),
)

logger.L().Info("service started", zap.String("version", "1.0"))

// 3. 函数跟踪 (Trace)
func MyBiz() {
    defer logger.Trace(logger.L(), "MyBiz")()
    // ... 业务逻辑
}
```

### 2. MongoDB 客户端 (mongo)
接口化设计，支持全量 Context 传递。

```go
import "github.com/weiweimhy/go-utils/mongo"

// 创建实例
client, err := mongo.NewClient(ctx, mongo.Config{
    Uri:          "mongodb://localhost:27017",
    DatabaseName: "mydb",
    OPTimeout:    5 * time.Second,
})

// 使用接口操作
err = client.FindOne(ctx, "users", bson.M{"id": 1}, &user)
```

### 3. 下载管理器 (download)
支持并发控制和级联取消。

```go
import "github.com/weiweimhy/go-utils/download"

// 创建管理器
dm := download.NewDownloadManager(
    download.WithWorkers(10),
    download.WithDelay(100 * time.Millisecond),
)

// 启动
dm.Start(ctx)

// 添加任务
dm.Add("http://example.com/file.zip", "./downloads/file.zip")

// 等待完成
dm.Wait()
```

### 4. EPUB 处理 (epub)
内存优化的按需加载模式。

```go
import "github.com/weiweimhy/go-utils/epub"

e, err := epub.Open("book.epub")
defer e.Close()

// 并发安全地批量处理页面，而不耗尽内存
e.ApplyHTML(func(name, html string) (string, error) {
    return strings.ReplaceAll(html, "old", "new"), nil
})

e.Save("book_new.epub")
```

### 5. 统一错误定义 (errs)
方便开发者进行错误匹配。

```go
import "github.com/weiweimhy/go-utils/errs"

if errors.Is(err, errs.ErrTimeout) {
    // 处理超时逻辑
}
```

---

## 🛡 开发规范 (项目准则)

- **禁止直接使用 `zap.L()`**：请务必使用 `logger.L()` 获通过 `context` 传递 Logger。
- **必需传 Context**：所有 IO 操作必须接收 `context` 参数以确保生命周期受控。
- **优先使用接口**：在业务层注入依赖时，请声明 `IMongoClient` 或 `IOCRClient`。

## 📄 License
MIT License
