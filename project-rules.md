# Go Utils 项目规则

> **优先级说明**：本文档的规则优先级**高于** `go-project-rules.md`。当两者冲突时，以本文档为准。
> 
> 本文档定义的是 `go-utils` 项目特定的规则和约定，而 `go-project-rules.md` 定义的是所有 Go 项目的通用规则。

## 规则优先级

1. **项目规则**（`project-rules.md`）- 最高优先级
2. **通用规则**（`go-project-rules.md`）- 通用优先级

> **重要**：在回答问题和检查代码时，需要同时读取本文档和 `go-project-rules.md`，优先遵循本文档的规则。

## 项目特定规则

### 1. 包命名规范

- **工具类包统一使用全小写命名**：
  - ✅ `httputil`、`htmlutil`、`strutil`、`filesystem`、`crypto`、`runtime`
  - ❌ `htmlUtils`、`httpUtil`（驼峰命名）
  - ❌ `strings`（与标准库冲突）

### 2. 日志使用规范

- **禁止直接使用** `zap.L()`，必须使用 `logger.L()` 或 `logger.FromContext()`
- **原因**：统一日志管理，确保 logger 已正确初始化

### 3. 工作池实现规范

- **所有工作池必须使用** `WorkerPool` + `Task` 模式实现
- **禁止**自己实现工作池（如 `DownloadManager` 等）
- **原因**：统一工作池实现，符合项目规范

### 4. 下载功能规范

- **单文件下载**：使用 `httputil.DownloadFile(url, path)`
- **批量下载**：使用 `httputil.DownloadTask` + `task.WorkerPool`
- **禁止**使用 `DownloadManager`（已删除，不符合规范）

### 5. 包拆分规范

- **按功能拆分包**，避免 `customUtils` 这种大杂烩包
- **包命名统一**：全小写，避免与标准库冲突

### 6. 公共库代码规范

- **代码必须简洁**：作为公共库，代码应该清晰、简洁，避免过度设计
- **不需要无用的注释**：作为公共库，代码应该自解释，避免冗余注释
  - ✅ 保留必要的文档注释（导出函数的说明）
  - ❌ 禁止无意义的注释（如 `// 设置变量`、`// 返回结果` 等）
- **对外方法简单易用**：作为公共库，应该保持对外方法简单易用的原则
  - 提供简洁的 API 接口
  - 避免复杂的参数结构
  - 提供合理的默认值
- **不自己消化 error**：作为公共库，不应该自己消化 error，应该抛出给调用方处理
  - ✅ 推荐：返回 error，由调用方决定如何处理
  - ❌ 禁止：在库内部吞掉 error 或只记录日志而不返回
  - **例外**：日志记录、资源清理等辅助操作的 error 可以在内部处理（如 defer 中的 Close 错误）

## 项目结构规范

```
go-utils/
├── crypto/          # 编码/加密工具
├── filesystem/      # 文件系统操作
├── httputil/        # HTTP/网络工具（全小写）
├── htmlutil/        # HTML 处理工具（全小写）
├── strutil/         # 字符串处理工具（避免与标准库冲突）
├── runtime/         # 运行时工具
├── localDB/         # 本地数据库
├── logger/          # 日志工具
├── task/            # 任务池（WorkerPool + Task）
└── ...              # 其他包
```

## 迁移指南

### 包名变更

```go
// 旧导入（已废弃）
import "github.com/weiweimhy/go-utils/customUtils"
import "github.com/weiweimhy/go-utils/strings"
import "github.com/weiweimhy/go-utils/htmlUtils"

// 新导入（v2）
import "github.com/weiweimhy/go-utils/v2/filesystem"  // 文件系统操作
import "github.com/weiweimhy/go-utils/v2/httputil"    // HTTP 操作
import "github.com/weiweimhy/go-utils/v2/htmlutil"    // HTML 处理
import "github.com/weiweimhy/go-utils/v2/strutil"     // 字符串处理
import "github.com/weiweimhy/go-utils/v2/crypto"      // 编码/加密
import "github.com/weiweimhy/go-utils/v2/runtime"     // 运行时工具
```

### 下载功能变更

```go
// ✅ 单文件下载（保持不变）
err := httputil.DownloadFile("https://example.com/file.pdf", "./file.pdf")

// ✅ 批量下载（新方式，使用 WorkerPool + Task）
pool := task.NewWorkerPool(10, 100)
defer pool.Close(30 * time.Second)

tasks := []*httputil.DownloadTask{
    httputil.NewDownloadTask("https://example.com/file1.pdf", "./file1.pdf", nil),
    httputil.NewDownloadTask("https://example.com/file2.pdf", "./file2.pdf", nil),
}
httputil.DownloadBatch(pool, tasks)

// ❌ 旧方式（已删除）
// dm := httputil.NewDownloadManager(...)  // 已删除
```

### 日志使用变更

```go
// ✅ 推荐：使用 logger.L()
import "github.com/weiweimhy/go-utils/v2/logger"
import "go.uber.org/zap"

logger.L().Info("message", zap.String("key", "value"))

// ❌ 禁止：直接使用 zap.L()
zap.L().Info("message", zap.String("key", "value")) // 禁止
```

