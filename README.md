# Go-Utils

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.24-blue.svg)](https://golang.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/weiweimhy/go-utils/v4.svg)](https://pkg.go.dev/github.com/weiweimhy/go-utils/v4)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

`go-utils` 是一个面向工程复用的 Go 语言工具库，主模块聚焦通用基础组件，外部集成按子模块拆分维护。

## 🚀 核心特性

- **模块收敛**：主模块只保留通用能力，重依赖集成拆为独立子模块
- **关键链路 Context**：网络、数据库、并发任务等关键链路原生支持生命周期管控与超时控制
- **Mock-Ready**：关键组件抽象为接口，方便单元测试
- **默认低副作用**：通用包默认不主动打印日志，日志能力按需启用
- **内存优化**：Epub 模块采用流式处理，轻松应对超大文件

## 📋 环境要求

- **Go 1.24+**（使用了 Go 1.24 的新特性）

## 📦 安装

```bash
go get github.com/weiweimhy/go-utils/v4
```

### 安装外部集成子模块

```bash
go get github.com/weiweimhy/go-utils/v4/mongo
go get github.com/weiweimhy/go-utils/v4/wechat
go get github.com/weiweimhy/go-utils/v4/tencentocr
```

### 升级到最新版本

```bash
go get -u github.com/weiweimhy/go-utils/v4
go mod tidy
```

### 安装指定版本

```bash
go get github.com/weiweimhy/go-utils/v4@v4.0.0
```

### 从 v3.x 或更早版本迁移

1. 更新 go.mod 中的依赖：

   ```bash
   go get github.com/weiweimhy/go-utils/v4@latest
   ```

2. 更新所有 import 路径，添加 `/v3` 后缀：

   ```go
   // 旧版本
   import "github.com/weiweimhy/go-utils/logger"

   // 新版本 (v4.x)
   import "github.com/weiweimhy/go-utils/v4/logger"
   ```

3. 执行 `go mod tidy` 清理依赖

### 强制刷新模块缓存

如果遇到缓存问题（如版本未更新），可强制清除并重新拉取：

```bash
# 清除模块缓存
go clean -modcache

# Bash / Zsh: 重新拉取指定版本
GOPROXY=https://proxy.golang.org,direct go get github.com/weiweimhy/go-utils/v4@v4.0.0

# PowerShell: 重新拉取指定版本
$env:GOPROXY = "https://proxy.golang.org,direct"
go get github.com/weiweimhy/go-utils/v4@v4.0.0
```

---

## 🛠 包速查表

| 包名 | 核心功能 | 适用场景 |
| :--- | :--- | :--- |
| `logger` | 结构化日志、Trace 追踪 | 全局日志记录、函数耗时分析 |
| `event` | 轻量事件总线 | 进程内发布订阅、模块解耦 |
| `task` | 通用工作池、任务分组 | 并发任务管理、goroutine 复用 |
| `download` | 高并发下载管理器 | 批量文件下载 |
| `httputil` | 安全 HTTP 客户端 | 远程 API 调用 |
| `fsutil` | 文件/目录操作 | 读写文件、自动创建父目录 |
| `cryptoutil` | Hash &amp; Base64 | 数据校验、编码转换 |
| `jwt` | JWT 用户鉴权认证 | Token 生成、验证、刷新 |
| `localdb` | BBolt 本地 KV 存储 | 轻量本地存储 |
| `htmlutil` | HTML DOM 解析 | 网页内容提取 |
| `errs` | 预定义错误 | 统一错误处理 |

---

## 📚 包分层建议

- **核心工具**：`cryptoutil`、`fsutil`、`regexputil`、`htmlutil`、`runtimeutil`
- **基础设施**：`event`、`task`、`httputil`、`logger`、`localdb`、`download`、`jwt`
- **外部集成子模块**：`github.com/weiweimhy/go-utils/v4/mongo`、`github.com/weiweimhy/go-utils/v4/wechat`、`github.com/weiweimhy/go-utils/v4/tencentocr`

如果你只需要轻量工具，优先依赖主模块；外部集成请按需引入对应子模块，避免把数据库驱动和云厂商 SDK 带进主模块。

## 🧭 发布策略

- 根模块：只承载通用工具与轻依赖基础设施能力
- 子模块：承载数据库、第三方平台、云厂商 SDK 等重依赖集成
- 不兼容变更：优先通过新 major 版本处理，而不是在同一 major 下保留大量历史兼容别名
- 文档入口：
  - `mongo`：[mongo/README.md](/E:/go/go-utils/mongo/README.md)
  - `wechat`：[wechat/README.md](/E:/go/go-utils/wechat/README.md)
  - `tencentocr`：[tencentocr/README.md](/E:/go/go-utils/tencentocr/README.md)

## 📖 快速上手

### 1. 日志 (logger)

```go
import (
    "github.com/weiweimhy/go-utils/v4/logger"
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
    "github.com/weiweimhy/go-utils/v4/logger"
    "github.com/weiweimhy/go-utils/v4/task"
)

func main() {
    ctx := context.Background()

    // 使用默认配置：worker 数和缓冲区都为 CPU 核心数
    defaultPool := task.NewWorkerPool(ctx)
    defer defaultPool.Close(5 * time.Second)

    // 创建工作池：10 个 worker，缓冲区 100，并附带业务名称
    pool := task.NewWorkerPool(
        ctx,
        task.WithWorkerCount(10),
        task.WithBufferSize(100),
        task.WithName("batch-process"),
        task.WithLogger(logger.L()),
    )
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

### 3. 事件总线 (event)

```go
import "github.com/weiweimhy/go-utils/v4/event"

func main() {
    bus := event.NewBus()
    defer bus.Close()

    unsubscribe, ok := bus.Subscribe("user.login", func(eventType string, data any) {
        println(eventType, data.(string))
    })
    if !ok {
        return
    }
    defer unsubscribe()

    bus.PublishSync("user.login", "alice")
}
```

### 4. 下载管理器 (download)

```go
import (
    "context"
    "github.com/weiweimhy/go-utils/v4/download"
    "github.com/weiweimhy/go-utils/v4/logger"
)

func main() {
    ctx := context.Background()

    dm := download.NewDownloadManager(
        download.WithWorkers(5),
        download.WithLogger(logger.L()),
    )
    dm.Start(ctx)
    defer dm.Close()

    // 添加下载任务
    dm.Add("https://example.com/file.zip", "./downloads/file.zip")
    dm.Wait() // 等待所有任务完成
}
```

### 5. HTTP 请求 (httputil)

```go
import (
    "context"
    "github.com/weiweimhy/go-utils/v4/httputil"
)

func main() {
    ctx := context.Background()

    // 获取字节流
    data, err := httputil.GetBytesFromURL(ctx, "https://api.example.com/data")

    // 获取字符串
    html, err := httputil.GetStringFromURL(ctx, "https://example.com")
}
```

### 5. 文件操作 (fsutil)

```go
import "github.com/weiweimhy/go-utils/v4/fsutil"

// 保存文件（自动创建父目录）
err := fsutil.SaveToFile("./data/output/result.json", data)

// 检查文件是否存在
exists := fsutil.FileExists("./config.json")

// 获取文件的 Base64 编码
base64Str, err := fsutil.GetFileBase64("./image.png")
```

### 6. MongoDB (独立子模块)

```go
import (
    "context"
    "time"
    "github.com/weiweimhy/go-utils/v4/mongo"
)

func main() {
    ctx := context.Background()

    // 创建客户端（超时有默认值）
    client, err := mongo.NewClient(ctx, mongo.Config{
        URI:          "mongodb://localhost:27017",
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
import "github.com/weiweimhy/go-utils/v4/localdb"

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
db.SetJSONValue("users", "user:1", user)
db.GetJSONValue("users", "user:1", &user)
```

### 8. HTML 解析 (htmlutil)

```go
import "github.com/weiweimhy/go-utils/v4/htmlutil"

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
import "github.com/weiweimhy/go-utils/v4/cryptoutil"

// SHA256 哈希
hash := cryptoutil.SHA256HexFromString("hello")     // 完整 64 字符
short := cryptoutil.SHA256Hex16FromString("hello")  // 前 16 字符

// Base64 编码
encoded := cryptoutil.Base64FromBytes(data)
```

### 10. 错误处理 (errs)

```go
import (
    "errors"
    "github.com/weiweimhy/go-utils/v4/errs"
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

### 11. JWT 认证 (jwt)

支持 **HMAC (HS256)** 和 **RSA (RS256)** 两种签名算法，适用于单体应用和分布式系统。

#### HMAC 模式（对称加密）

```go
import (
    "context"
    "time"
    "github.com/weiweimhy/go-utils/v4/jwt"
)

func main() {
    ctx := context.Background()

    // 创建 JWT 实例
    j, err := jwt.NewJWT(
        jwt.WithSecret("your-256-bit-secret-key!!"),
        jwt.WithAccessTokenExpiry(15 * time.Minute),
        jwt.WithRefreshTokenExpiry(7 * 24 * time.Hour),
    )
    if err != nil {
        panic(err)
    }

    // 生成令牌对
    tokens, err := j.Generate(ctx, "user123", map[string]any{"role": "admin"})

    // 验证令牌
    claims, err := j.Validate(ctx, tokens.AccessToken)
    fmt.Println(claims.UserID) // "user123"

    // 刷新令牌
    newTokens, err := j.Refresh(ctx, tokens.RefreshToken)
}
```

#### RSA 模式（非对称加密，适合分布式系统）

```go
import (
    "context"
    "crypto/rsa"
    "github.com/weiweimhy/go-utils/v4/jwt"
)

// 认证中心：使用私钥签发令牌
func authCenter(privateKey *rsa.PrivateKey) {
    ctx := context.Background()

    j, _ := jwt.NewJWT(jwt.WithPrivateKey(privateKey))
    tokens, _ := j.Generate(ctx, "user123", map[string]any{"role": "admin"})
    // 返回 tokens 给客户端
}

// 业务服务：只需公钥验证令牌
func businessService(publicKey *rsa.PublicKey, accessToken string) {
    ctx := context.Background()

    j, _ := jwt.NewJWT(jwt.WithPublicKey(publicKey))
    claims, err := j.Validate(ctx, accessToken)
    if err != nil {
        // 验证失败
        return
    }
    fmt.Println(claims.UserID) // "user123"
}
```

#### 可用 Option

| Option | 说明 |
| :--- | :--- |
| `WithSecret(string)` | HMAC 密钥 |
| `WithPrivateKey(*rsa.PrivateKey)` | RSA 私钥（用于签名） |
| `WithPublicKey(*rsa.PublicKey)` | RSA 公钥（用于验证） |
| `WithPrivateKeyPEM([]byte)` | PEM 格式私钥 |
| `WithPublicKeyPEM([]byte)` | PEM 格式公钥 |
| `WithSigningMethod(jwt.SigningMethod)` | 自定义签名算法 |
| `WithAccessTokenExpiry(time.Duration)` | 访问令牌有效期 |
| `WithRefreshTokenExpiry(time.Duration)` | 刷新令牌有效期 |
| `WithIssuer(string)` | 签发者标识 |

---

## 🛡 开发规范

1. **优先传递 Context**：网络、数据库、并发任务等长生命周期操作应接收 `context` 参数
2. **日志按需启用**：通用包默认保持静默，业务侧需要时显式注入 logger
3. **优先清晰 API**：公共包优先暴露明确的具体类型，避免为封装而封装

---

## 📈 版本管理

遵循 [Semantic Versioning](https://semver.org/)：

- **Major**：不兼容的 API 变更
- **Minor**：向后兼容的功能新增
- **Patch**：向后兼容的问题修正

---

## 📄 License

[MIT License](LICENSE)
