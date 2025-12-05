# Go 项目开发规则（通用）

> 本文档定义了所有 Go 项目都应该遵守的通用开发规则和最佳实践。这些规则适用于任何 Go 项目，与具体业务无关。

## 1. 并发编程规范

### 1.1 Goroutine 管理
- **核心原则**：**所有 goroutine 泄漏 = 线上事故**
- **推荐在「需要协同取消/整体失败」的场景使用** `errgroup` + `context` 管理并发（例如：任一子任务失败就整体报错的聚合类任务）
- **允许根据场景选择**：
  - 长时间运行的后台 goroutine、worker pool、定时任务等，可以使用 `context` + `sync.WaitGroup` 等方式管理生命周期
  - 关键点是：**每一个 goroutine 都必须能被追踪、能被取消、能在需要时等待其退出**
- **禁止**创建无法追踪和管理的 goroutine
- **示例**：
  ```go
  // ✅ 推荐：使用 errgroup + ctx 管理并发
  func (m *Manager) Start(ctx context.Context) error {
      g, gctx := errgroup.WithContext(ctx)
      
      for _, cfg := range m.configs {
          cfg := cfg // 避免闭包问题
          g.Go(func() error {
              return m.processEmail(gctx, cfg)
          })
      }
      
      return g.Wait() // 等待所有 goroutine 完成，自动处理错误
  }
  
  // ❌ 禁止：无法追踪和管理的 goroutine
  func (m *Manager) Start() {
      for _, cfg := range m.configs {
          go func(c ClientConfig) {
              // ❌ 无法追踪、无法取消、无法等待完成
              m.processEmail(c)
          }(cfg)
      }
  }
  ```

### 1.2 sync.WaitGroup 使用规范
- **必须配合** `defer wg.Done()` 使用
- **禁止**将 `wg.Add(1)` 放在 goroutine 内部
- **原因**：避免竞态条件，确保 Add 和 Wait 的调用顺序正确
- **示例**：
  ```go
  // ✅ 推荐：Add 在 goroutine 外，Done 使用 defer
  func processTasks(tasks []Task) {
      var wg sync.WaitGroup
      for _, task := range tasks {
          wg.Add(1) // ✅ 在 goroutine 外调用
          go func(t Task) {
              defer wg.Done() // ✅ 使用 defer 确保执行
              processTask(t)
          }(task)
      }
      wg.Wait()
  }
  
  // ❌ 禁止：Add 放在 goroutine 内部
  func processTasks(tasks []Task) {
      var wg sync.WaitGroup
      for _, task := range tasks {
          go func(t Task) {
              wg.Add(1) // ❌ 禁止：可能竞态条件
              defer wg.Done()
              processTask(t)
          }(task)
      }
      wg.Wait() // 可能提前返回
  }
  
  // ❌ 禁止：不使用 defer
  func processTasks(tasks []Task) {
      var wg sync.WaitGroup
      for _, task := range tasks {
          wg.Add(1)
          go func(t Task) {
              processTask(t)
              wg.Done() // ❌ 禁止：如果 panic 会导致 Wait 永远阻塞
          }(task)
      }
      wg.Wait()
  }
  ```

### 1.4 Worker Pool 实现规范
- **所有工作池必须使用** `WorkerPool` + `Task` 模式实现
- **原因**：
  - 统一工作池实现，便于管理和维护
  - 提供标准的生命周期管理（启动、提交任务、优雅关闭）
  - 避免重复实现，减少 bug
- **示例**：
  ```go
  // ✅ 推荐：使用 WorkerPool + Task 模式
  import "github.com/weiweimhy/go-utils/v2/task"
  
  type MyTask struct {
      data string
  }
  
  func (t *MyTask) Execute() {
      // 处理任务
      processData(t.data)
  }
  
  func main() {
      pool := task.NewWorkerPool(10, 100) // 10 个 worker，缓冲区 100
      defer pool.Close(30 * time.Second)
      
      // 提交任务
      pool.Submit(&MyTask{data: "test"})
  }
  
  // ❌ 禁止：自己实现工作池
  func main() {
      jobs := make(chan Job, 100)
      var wg sync.WaitGroup
      for i := 0; i < 10; i++ {
          wg.Add(1)
          go func() {
              defer wg.Done()
              for job := range jobs {
                  processJob(job)
              }
          }()
      }
      // ❌ 禁止：重复实现，容易出错
  }
  ```

### 1.3 Channel 使用规范
- **必须**带缓冲或者配合 `select` + `ctx.Done()`
- **原因**：避免 goroutine 泄漏和死锁
- **示例**：
  ```go
  // ✅ 推荐：带缓冲的 channel
  func processWithBufferedChannel() {
      ch := make(chan Task, 10) // ✅ 带缓冲
      go func() {
          for task := range ch {
              processTask(task)
          }
      }()
      
      for _, task := range tasks {
          ch <- task // 不会阻塞（有缓冲）
      }
      close(ch)
  }
  
  // ✅ 推荐：配合 select + ctx.Done()
  func processWithSelect(ctx context.Context) {
      ch := make(chan Task) // 无缓冲
      go func() {
          for {
              select {
              case task := <-ch:
                  processTask(task)
              case <-ctx.Done():
                  return // ✅ 可以取消，避免泄漏
              }
          }
      }()
      
      for _, task := range tasks {
          select {
          case ch <- task:
          case <-ctx.Done():
              return
          }
      }
      close(ch)
  }
  
  // ❌ 禁止：无缓冲且无 select + ctx.Done()
  func processUnsafe() {
      ch := make(chan Task) // ❌ 无缓冲
      go func() {
          for task := range ch {
              processTask(task)
          }
      }()
      
      for _, task := range tasks {
          ch <- task // ❌ 可能阻塞，如果接收方未准备好
      }
      close(ch)
  }
  ```

## 2. 性能优化规范

### 2.1 日志性能优化
- **禁止使用** `zap.SugaredLogger`（性能差 5~10 倍）
- **必须使用** `zap.Logger`（结构化日志，零分配）
- **生产环境必须开启 Sampling**：`Initial: 100, Thereafter: 100`
  - 防止高并发场景下日志刷屏
  - 配置示例：
    ```go
    logger := zap.New(
        core,
        zap.AddCaller(),
        zap.AddStacktrace(zap.ErrorLevel),
        zap.WrapCore(func(c zapcore.Core) zapcore.Core {
            return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
        }),
    )
    ```
- **示例**：
  ```go
  // ✅ 推荐：使用 zap.Logger（结构化日志）
  logger.Info("message", zap.String("key", "value"))
  
  // ❌ 禁止：使用 zap.SugaredLogger（性能差）
  sugar := logger.Sugar()
  sugar.Infof("message: %s", value) // ❌ 性能差 5~10 倍
  ```

### 2.2 JSON 序列化性能优化
- **禁止在生产环境使用** `encoding/json.Marshal`（性能差）
- **必须使用**高性能 JSON 库：
  - `github.com/bytedance/sonic`（推荐，性能最优）
  - `github.com/json-iterator/go`（备选）
- **示例**：
  ```go
  // ✅ 推荐：使用 sonic（性能最优）
  import "github.com/bytedance/sonic"
  data, err := sonic.Marshal(obj)
  
  // ✅ 备选：使用 jsoniter（性能好）
  import jsoniter "github.com/json-iterator/go"
  var json = jsoniter.ConfigCompatibleWithStandardLibrary
  data, err := json.Marshal(obj)
  
  // ❌ 禁止：使用标准库 json（性能差）
  import "encoding/json"
  data, err := json.Marshal(obj) // ❌ 性能差
  ```

### 2.3 字符串拼接性能优化
- **禁止使用** `strings.Builder` + `fmt.Sprintf`（性能差）
- **推荐使用**：
  - `strings.Join`：适用于字符串切片拼接
  - `bytes.Buffer`：适用于复杂拼接场景
- **示例**：
  ```go
  // ✅ 推荐：使用 strings.Join
  parts := []string{"a", "b", "c"}
  result := strings.Join(parts, ",")
  
  // ✅ 推荐：使用 bytes.Buffer（复杂场景）
  var buf bytes.Buffer
  buf.WriteString("prefix")
  buf.WriteString(str)
  buf.WriteString("suffix")
  result := buf.String()
  
  // ❌ 禁止：strings.Builder + fmt.Sprintf
  var builder strings.Builder
  builder.WriteString(fmt.Sprintf("value: %s", str)) // ❌ 性能差
  result := builder.String()
  ```

### 2.4 Reflect 使用规范
- **禁止在性能敏感路径使用** `reflect` 包
- **原因**：reflect 性能开销大，影响程序性能
- **替代方案**：
  - 使用代码生成（如 `go generate`）
  - 使用泛型（Go 1.18+）
  - 手动编写类型转换代码
- **示例**：
  ```go
  // ❌ 禁止：在性能敏感路径使用 reflect
  func process(data interface{}) {
      v := reflect.ValueOf(data) // ❌ 性能差
      // ...
  }
  
  // ✅ 推荐：使用泛型或代码生成
  func process[T any](data T) {
      // 类型安全，无反射开销
  }
  ```

### 2.5 []byte 复用优化
- **所有 []byte 复用必须使用** `sync.Pool`
- **原因**：减少内存分配，提升性能
- **示例**：
  ```go
  // ✅ 推荐：使用 sync.Pool 复用 []byte
  var bufferPool = sync.Pool{
      New: func() interface{} {
          return make([]byte, 0, 1024)
      },
  }
  
  func process() {
      buf := bufferPool.Get().([]byte)
      defer bufferPool.Put(buf[:0]) // 重置长度，保留容量
      
      buf = append(buf, "data"...)
      // 使用 buf
  }
  
  // ❌ 禁止：每次都分配新的 []byte
  func process() {
      buf := make([]byte, 0, 1024) // ❌ 每次都分配
      buf = append(buf, "data"...)
  }
  ```

## 3. 配置与启动终极铁律

### 3.1 启动参数校验
- **启动参数校验失败必须** `os.Exit(1)`，不要 return error 让 main 继续跑
- **原因**：启动参数错误是致命错误，程序不应该继续运行
- **示例**：
  ```go
  // ✅ 推荐：启动参数校验失败直接退出
  func main() {
      configPath := flag.String("config", "config.toml", "config file path")
      flag.Parse()
      
      if *configPath == "" {
          fmt.Fprintf(os.Stderr, "error: config path is required\n")
          os.Exit(1) // ✅ 直接退出
      }
      
      cfg, err := config.LoadConfig(*configPath)
      if err != nil {
          fmt.Fprintf(os.Stderr, "error: failed to load config: %v\n", err)
          os.Exit(1) // ✅ 直接退出
      }
      
      // ... 继续启动
  }
  
  // ❌ 禁止：return error 让 main 继续跑
  func main() {
      cfg, err := loadConfig()
      if err != nil {
          log.Error("failed to load config", zap.Error(err))
          return // ❌ 禁止：程序可能继续运行，导致不可预期行为
      }
  }
  ```

### 3.2 配置解析规范
- **所有配置必须使用** `pflag` + `viper`
- **禁止**自己解析 flag 或配置文件
- **原因**：统一配置管理，支持环境变量、配置文件、命令行参数等多种来源
- **示例**：
  ```go
  // ✅ 推荐：使用 pflag + viper
  import (
      "github.com/spf13/pflag"
      "github.com/spf13/viper"
  )
  
  func initConfig() {
      pflag.String("config", "config.toml", "config file path")
      pflag.String("port", "8080", "server port")
      pflag.Parse()
      viper.BindPFlags(pflag.CommandLine)
      
      viper.SetConfigFile(viper.GetString("config"))
      viper.AutomaticEnv() // 支持环境变量
      viper.ReadInConfig()
  }
  
  // ❌ 禁止：自己解析 flag
  func initConfig() {
      configPath := flag.String("config", "config.toml", "config file path")
      flag.Parse()
      // ❌ 禁止：功能单一，不支持环境变量等
  }
  ```

### 3.3 Graceful Shutdown
- **所有服务启动必须实现** graceful shutdown（监听 SIGINT/SIGTERM）
- **原因**：优雅关闭，确保资源正确释放，数据不丢失
- **示例**：
  ```go
  // ✅ 推荐：实现 graceful shutdown
  func main() {
      ctx, cancel := context.WithCancel(context.Background())
      defer cancel()
      
      // 启动服务
      srv := startServer(ctx)
      
      // 监听信号
      sigChan := make(chan os.Signal, 1)
      signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
      
      <-sigChan
      log.Info("shutting down gracefully...")
      
      // 优雅关闭
      shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
      defer shutdownCancel()
      
      if err := srv.Shutdown(shutdownCtx); err != nil {
          log.Fatal("server shutdown failed", zap.Error(err))
      }
      
      log.Info("server stopped")
  }
  
  // ❌ 禁止：不实现 graceful shutdown
  func main() {
      srv := startServer()
      // ❌ 禁止：直接运行，无法优雅关闭
      select {} // 永远阻塞
  }
  ```

### 3.4 端口监听失败处理
- **所有端口监听失败必须** `log.Fatal`，禁止 return error
- **原因**：端口监听失败是致命错误，程序无法继续运行
- **示例**：
  ```go
  // ✅ 推荐：端口监听失败直接 Fatal
  func startServer() {
      listener, err := net.Listen("tcp", ":8080")
      if err != nil {
          log.Fatal("failed to listen on port 8080", zap.Error(err))
          // ✅ 直接退出，不会继续执行
      }
      
      log.Info("server listening on :8080")
      // ... 继续启动
  }
  
  // ❌ 禁止：return error
  func startServer() error {
      listener, err := net.Listen("tcp", ":8080")
      if err != nil {
          return err // ❌ 禁止：调用方可能忽略错误，程序继续运行
      }
      return nil
  }
  ```

## 4. 依赖与构建终极铁律

### 4.1 依赖版本管理
- **所有第三方库必须使用** `go.mod` 的版本，不要用 master/latest
- **原因**：
  - 保证构建的可重复性和稳定性
  - 避免因依赖更新导致的意外破坏
  - 便于版本追踪和回滚
- **示例**：
  ```go
  // ✅ 推荐：使用 go.mod 指定版本
  // go.mod
  require (
      github.com/spf13/viper v1.18.2
      github.com/bytedance/sonic v1.11.0
      go.uber.org/zap v1.26.0
  )
  
  // ❌ 禁止：使用 master/latest 分支
  require (
      github.com/spf13/viper master // ❌ 禁止
      github.com/bytedance/sonic latest // ❌ 禁止
  )
  ```

### 4.2 依赖更新规范
- **必须使用** `go get package@version` 更新依赖
- **禁止**直接修改 `go.mod` 文件中的版本号
- **必须**在更新后运行 `go mod tidy` 清理未使用的依赖
- **示例**：
  ```bash
  # ✅ 推荐：使用 go get 更新依赖
  go get github.com/spf13/viper@v1.18.2
  go mod tidy
  
  # ❌ 禁止：直接修改 go.mod
  # 手动编辑 go.mod 文件修改版本号
  ```

### 4.3 依赖版本选择原则
- **优先选择**稳定版本（如 `v1.18.2`）
- **避免使用**预发布版本（如 `v1.18.2-beta.1`）除非必要
- **禁止使用** `@latest`、`@master`、`@main` 等标签
- **必须锁定**所有直接依赖的版本

## 5. 日志使用规范（通用原则）

### 5.0 日志库使用规范
- **禁止直接使用**标准库 `log` 包
- **禁止直接使用** `zap.L()`，必须使用 `logger` 封装库的方法
- **必须使用** `logger` 封装库（`github.com/weiweimhy/go-utils/v2/logger`）
- **原因**：
  - 统一日志格式和输出
  - 支持结构化日志、日志级别、采样等高级功能
  - 便于日志收集和分析
  - 确保 logger 已正确初始化
- **示例**：
  ```go
  // ✅ 推荐：使用 logger 封装库
  import (
      "github.com/weiweimhy/go-utils/v2/logger"
      "go.uber.org/zap"
  )
  
  logger.InitProduction()
  logger.L().Info("message", zap.String("key", "value"))
  
  // ✅ 推荐：使用 FromContext（如果有 context）
  log := logger.FromContext(ctx)
  log.Info("message", zap.String("key", "value"))
  
  // ❌ 禁止：直接使用 zap.L()
  zap.L().Info("message", zap.String("key", "value")) // ❌ 禁止：必须使用 logger.L()
  
  // ❌ 禁止：使用标准库 log
  import "log"
  log.Println("message") // ❌ 禁止：功能单一，无法统一管理
  ```

### 5.1 错误日志打印规则

**核心原则**：**谁调用，谁负责打 Error 日志**。函数内部只打 Debug/Info/Warn 路径日志，绝不打 Error（除非是不可恢复的致命错误）。

1. **函数内部职责**：
   - 只负责打印「进入/退出/关键路径」等上下文日志
     - 函数进入/退出日志：使用 **Debug** 级别（避免生产环境日志爆炸）
     - 关键路径日志：使用 **Info/Warn** 级别（重要的业务操作节点）
   - **禁止**在函数内部打印调用失败的 Error 日志（依赖调用失败、业务逻辑错误等）
   - 函数内部遇到错误时，直接返回 error，由调用方决定如何处理和记录
   - **例外情况**：
     - **参数校验失败**：参数校验失败 = 代码 Bug = 必须在函数内部打 Error 日志 + 必须返回 error
     - **不可恢复的致命错误**：如 panic recover 等，必须在函数内部打 Error/Fatal 日志

2. **调用方职责**：
   - **必须**负责打印业务错误日志（Error 级别）
   - 包括：依赖调用失败（网络/IO）、业务逻辑错误、预期外情况等
   - 调用方有更完整的上下文信息，可以记录更准确的错误信息

### 5.2 日志格式规范

**核心原则**：日志必须使用**英文**，不同类型的日志必须带上相应的上下文字段。

#### 5.2.1 函数进入/退出日志（函数内部）
- **位置**：函数内部
- **级别**：**Debug**（必须使用 Debug 级别，避免生产环境日志爆炸）
- **必须字段**：
  - `func=<函数名>`：标识函数
  - `step=enter`：函数进入
  - `step=exit`：函数退出
- **说明**：函数进入/退出日志主要用于开发调试，生产环境通常只记录 Info 及以上级别，使用 Debug 可以避免日志过多影响性能

#### 5.2.2 参数校验失败（函数内部）
- **位置**：函数内部
- **级别**：**Error**（必须使用 Error 级别）
- **说明**：**参数校验失败 = 代码 Bug**，必须在函数内部打 Error 日志 + 必须返回 error。这是函数内部唯一允许打 Error 日志的情况（除了不可恢复的致命错误）。
- **必须字段**：
  - `error=invalid_param`：标识参数校验错误
  - `param=<参数名>`：具体参数名，如 `param=host`、`param=userName`
  - `func=<函数名>`：标识函数（包含行号，如 `"Client.Start:45"`）

#### 5.2.3 依赖调用失败（调用方）
- **位置**：调用方
- **级别**：Error
- **必须字段**：
  - `action=<操作名称>`：如 `action=create_email`、`action=connect_imap`
  - `uid=<唯一标识>`：如 `uid=123`（邮件ID、用户ID等）
  - `upstream=<上游服务>`：如 `upstream=imap`、`upstream=database`
  - `zap.Error(err)`：错误详情

#### 5.2.4 业务逻辑错误（调用方）
- **位置**：调用方
- **级别**：Error/Warn
- **必须字段**：
  - `action=<操作名称>`：如 `action=parse_body`、`action=validate_invoice`
  - `reason=<错误原因>`：如 `reason=invalid_html`、`reason=missing_field`

#### 5.2.5 致命错误（函数内部）
- **位置**：函数内部（panic recover）
- **级别**：Error/Fatal
- **必须字段**：
  - `func=<函数名>`：标识函数（包含行号）
  - `stacktrace`：zap 自动带（通过 `zap.AddStacktrace` 配置）

### 5.3 日志字段总结
| 日志类型 | 位置 | 级别 | 必须字段 | 说明 |
|---------|------|------|---------|------|
| 函数进入/退出 | 函数内部 | **Debug** | `func`, `step=enter/exit` | 避免生产环境日志爆炸 |
| 参数校验失败 | 函数内部 | **Error** | `error=invalid_param`, `param`, `func` | **代码 Bug**，必须在函数内部打 Error |
| 依赖调用失败 | 调用方 | Error | `action`, `upstream`, `uid`(可选), `err` | 由调用方负责打 Error |
| 业务逻辑错误 | 调用方 | Error/Warn | `action`, `reason` | 由调用方负责打 Error |
| 致命错误 | 函数内部 | Error/Fatal | `func`, `stacktrace`(自动) | panic recover，必须在函数内部打 Error |

## 6. 必须事项 (Do's)

- **必须**使用 `errgroup` + `context` 管理所有并发，防止 goroutine 泄漏
- **必须**在使用 `sync.WaitGroup` 时配合 `defer wg.Done()`
- **必须**在 `sync.WaitGroup` 中将 `wg.Add(1)` 放在 goroutine 外部
- **必须**使用带缓冲的 channel 或配合 `select` + `ctx.Done()` 使用 channel
- **必须**使用 `WorkerPool` + `Task` 模式实现所有工作池
- **必须**在生产环境使用 `zap.Logger`（禁止使用 `zap.SugaredLogger`）
- **必须**在生产环境开启日志 Sampling（Initial: 100, Thereafter: 100）
- **必须**在生产环境使用高性能 JSON 库（`sonic` 或 `jsoniter`，禁止使用标准库 `json`）
- **必须**使用 `strings.Join` 或 `bytes.Buffer` 进行字符串拼接（禁止 `strings.Builder` + `fmt.Sprintf`）
- **必须**使用 `sync.Pool` 复用所有 `[]byte`
- **必须**在启动参数校验失败时使用 `os.Exit(1)`，禁止 return error
- **必须**使用 `pflag` + `viper` 管理所有配置，禁止自己解析 flag
- **必须**实现 graceful shutdown（监听 SIGINT/SIGTERM）
- **必须**在端口监听失败时使用 `log.Fatal`，禁止 return error
- **必须**使用 `go.mod` 指定所有第三方库的版本，禁止使用 master/latest
- **必须**使用 `go get package@version` 更新依赖，禁止直接修改 `go.mod`
- **必须**在更新依赖后运行 `go mod tidy` 清理未使用的依赖
- **必须**使用 `logger` 封装库，禁止直接使用标准库 `log`
- **必须**使用 `logger.L()` 或 `logger.FromContext()` 获取 logger，禁止直接使用 `zap.L()`

## 7. 禁止事项 (Don'ts)

- **禁止**使用标准库 `log`，必须使用 `logger` 封装库
- **禁止**直接使用 `zap.L()`，必须使用 `logger.L()` 或 `logger.FromContext()`
- **禁止**在函数内部打印业务错误日志（Error 级别），错误日志应在调用方打印
  - **例外**：参数校验失败（代码 Bug）和不可恢复的致命错误（panic recover）必须在函数内部打 Error
- **禁止**使用中文日志，所有日志必须使用英文
- **禁止**缺少必要的上下文字段（如 `func`、`step`、`action`、`upstream` 等）
- **禁止**创建无法追踪和管理的 goroutine，必须使用 `errgroup` + `context` 管理所有并发
- **禁止**将 `wg.Add(1)` 放在 goroutine 内部
- **禁止**使用 `sync.WaitGroup` 时不配合 `defer wg.Done()`
- **禁止**使用无缓冲 channel 且不配合 `select` + `ctx.Done()`
- **禁止**自己实现工作池，必须使用 `WorkerPool` + `Task` 模式
- **禁止**在生产环境使用 `zap.SugaredLogger`（性能差 5~10 倍）
- **禁止**在生产环境使用 `encoding/json.Marshal`，必须使用 `sonic` 或 `jsoniter`
- **禁止**使用 `strings.Builder` + `fmt.Sprintf`，必须使用 `strings.Join` 或 `bytes.Buffer`
- **禁止**在性能敏感路径使用 `reflect` 包
- **禁止**不使用 `sync.Pool` 复用 `[]byte`
- **禁止**启动参数校验失败时 return error，必须使用 `os.Exit(1)`
- **禁止**自己解析 flag，必须使用 `pflag` + `viper`
- **禁止**服务启动时不实现 graceful shutdown
- **禁止**端口监听失败时 return error，必须使用 `log.Fatal`
- **禁止**使用 master/latest/main 等标签作为依赖版本，必须使用具体版本号
- **禁止**直接修改 `go.mod` 文件中的版本号，必须使用 `go get` 命令
- **禁止**更新依赖后不运行 `go mod tidy`

