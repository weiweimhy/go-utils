# Go-Utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.24-blue.svg)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/weiweimhy/go-utils.svg)](https://pkg.go.dev/github.com/weiweimhy/go-utils)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`go-utils` 是一个工业级的通用 Go 语言封装库，旨在提供稳健、高性能且易于测试的基础组件。

## 🚀 核心特性

- **模块化设计**：功能拆分清晰，按需导入
- **全链路 Context**：原生支持生命周期管控与超时控制
- **Mock-Ready**：关键组件抽象为接口，方便单元测试
- **高性能**：集成 `zap`（日志）、`sonic`（JSON）、`bbolt`（KV存储）
- **内存优化**：Epub 模块采用流式处理，轻松应对超大文件

## � 环境要求

- **Go 1.24+**（使用了 Go 1.24 的新特性）

## �📦 安装

```bash
go get github.com/weiweimhy/go-utils
```

### 升级到最新版本

```bash
go get -u github.com/weiweimhy/go-utils
go mod tidy
```

---

## 🛠 包速查表

| 包名 | 核心功能 | 适用场景 |
| :--- | :--- | :--- |
| `logger` | 结构化日志、Trace 追踪 | 全局日志记录、函数耗时分析 |
| `task` | 通用工作池、任务分组 | 并发任务管理、goroutine 复用 |
| `download` | 高并发下载管理器 | 批量文件下载 |
| `httputil` | 安全 HTTP 客户端 | 远程 API 调用 |
| `fsutil` | 文件/目录操作 | 读写文件、自动创建父目录 |
| `cryptoutil` | Hash &amp; Base64 | 数据校验、编码转换 |
| `mongo` | MongoDB 客户端（接口化） | 数据库 CRUD |
| `localdb` | BBolt 本地 KV 存储 | 轻量本地存储 |
| `htmlutil` | HTML DOM 解析 | 网页内容提取 |
| `tencentocr` | 腾讯 OCR 封装 | 发票识别、文字提取 |
| `errs` | 预定义错误 | 统一错误处理 |

---

## 📖 快速上手

### 1. 日志 (logger)

```go
import (
    "github.com/weiweimhy/go-utils/v3/logger"
    "go.uber.org/zap"
)

func main() {
    // 生产环境初始化
    logger.Init(
        logger.WithFilename("./logs/app.log"),
        logger.WithLevel(zap.InfoLevel),
    )
    // 或开发环境（彩色控制台输出）
    // logger.InitDevelopment()

    // 使用
    logger.L().Info("server started", zap.Int("port", 8080))

    // 函数追踪（自动记录耗时和 panic）
    defer logger.Trace(logger.L(), "main")()
}
```

### 2. 工作池 (task)

```go
import (
    "context"
    "time"
    "github.com/weiweimhy/go-utils/v3/task"
)

func main() {
    ctx := context.Background()

    // 创建工作池：10 个 worker，缓冲区 100
    pool := task.NewWorkerPool(ctx, 10, 100)
    defer pool.Close(5 * time.Second)

    // 提交单个任务
    pool.SubmitFunc(func(ctx context.Context) {
        // 业务逻辑
    })

    // 使用任务组批量提交并等待
    items := []string{"a", "b", "c"}
    group := pool.NewGroup()
    for _, item := range items {
        item := item // 捕获循环变量
        group.SubmitFunc(func(ctx context.Context) {
            process(item)
        })
    }
    group.Wait()
}
```

### 3. 下载管理器 (download)

```go
import (
    "context"
    "github.com/weiweimhy/go-utils/v3/download"
)

func main() {
    ctx := context.Background()

    dm := download.NewDownloadManager(download.WithWorkers(5))
    dm.Start(ctx)
    defer dm.Close()

    // 添加下载任务
    dm.Add("https://example.com/file.zip", "./downloads/file.zip")
    dm.Wait() // 等待所有任务完成
}
```

### 4. HTTP 请求 (httputil)

```go
import (
    "context"
    "github.com/weiweimhy/go-utils/v3/httputil"
)

func main() {
    ctx := context.Background()

    // 获取字节流
    data, err := httputil.GetBytesFromUrl(ctx, "https://api.example.com/data")

    // 获取字符串
    html, err := httputil.GetStringFromUrl(ctx, "https://example.com")
}
```

### 5. 文件操作 (fsutil)

```go
import "github.com/weiweimhy/go-utils/v3/fsutil"

// 保存文件（自动创建父目录）
err := fsutil.SaveToFile("./data/output/result.json", data)

// 检查文件是否存在
exists := fsutil.IsFileExist("./config.json")

// 获取文件的 Base64 编码
base64Str, err := fsutil.GetFileBase64("./image.png")
```

### 6. MongoDB (mongo)

```go
import (
    "context"
    "time"
    "github.com/weiweimhy/go-utils/v3/mongo"
)

func main() {
    ctx := context.Background()

    // 创建客户端（超时有默认值）
    client, err := mongo.NewClient(ctx, mongo.Config{
        Uri:          "mongodb://localhost:27017",
        DatabaseName: "mydb",
    })
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(ctx)

    // CRUD 操作
    client.InsertOne(ctx, "users", map[string]any{"name": "Alice"})
    client.FindOne(ctx, "users", map[string]any{"name": "Alice"}, &result)
}
```

### 7. 本地 KV 存储 (localdb)

```go
import "github.com/weiweimhy/go-utils/v3/localdb"

// 打开数据库
db, err := localdb.Open("./data", "cache.db")
if err != nil {
    panic(err)
}
defer db.Close()

// 存取数据
db.Set("bucket", "key", []byte("value"))
data, _ := db.Get("bucket", "key")

// JSON 序列化存取
db.SetJSON("users", "user:1", user)
db.GetJSON("users", "user:1", &user)
```

### 8. HTML 解析 (htmlutil)

```go
import "github.com/weiweimhy/go-utils/v3/htmlutil"

html := `<div class="content"><p>Hello</p><p>World</p></div>`

// 按标签提取
texts, err := htmlutil.ExtractTextByTag(html, "p")
// texts = ["Hello", "World"]

// 按 class 提取
texts, err := htmlutil.ExtractTextByClass(html, "content")

// 按 ID 提取
text, err := htmlutil.ExtractTextByID(html, "main")

// 提取所有文本
allText, err := htmlutil.ExtractAllText(html)
```

### 9. 加密工具 (cryptoutil)

```go
import "github.com/weiweimhy/go-utils/v3/cryptoutil"

// SHA256 哈希
hash := cryptoutil.StringToHash("hello")      // 完整 64 字符
short := cryptoutil.StringToHash16("hello")  // 前 16 字符

// Base64 编码
encoded := cryptoutil.GetBase64FromBytes(data)
```

### 10. 错误处理 (errs)

```go
import (
    "errors"
    "github.com/weiweimhy/go-utils/v3/errs"
)

// 使用预定义错误
if err != nil {
    if errors.Is(err, errs.ErrNotFound) {
        // 处理未找到
    }
    if errors.Is(err, errs.ErrDownloadManagerClosed) {
        // 下载管理器已关闭
    }
}
```

---

## 🛡 开发规范

1. **必传 Context**：所有 IO 操作必须接收 `context` 参数
2. **使用 logger.L()**：禁止直接使用 `zap.L()`
3. **接口编程**：注入依赖时使用 `IMongoClient` 等接口

---

## 📈 版本管理

遵循 [Semantic Versioning](https://semver.org/)：

- **Major**：不兼容的 API 变更
- **Minor**：向后兼容的功能新增
- **Patch**：向后兼容的问题修正

---

## 📄 License

[MIT License](LICENSE)
