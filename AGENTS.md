# go-utils 协作指南

## 项目定位

`go-utils` 是面向工程复用的 Go 工具库。根模块
`github.com/weiweimhy/go-utils/v6` 保持标准库优先、低依赖与默认静默；重依赖、
本地存储和第三方平台能力位于独立的 v6 子模块。它不是可启动的服务或应用。

## 模块与目录

- 根模块：通用、无业务语义的小包；不得引入第三方依赖。
- 扩展子模块：`download`、`epub`、`htmlutil`、`jwt`、`localdb`、`logger`、
  `mongo`、`tencentocr`、`wechat`；各自拥有 `go.mod`、依赖和测试。
- `internal/bench`：仅用于基准测试辅助，不作为对外 API。
- `docs/`：项目事实、维护流程与详细标准；开始修改前先阅读相关文档。

完整的边界见 [docs/architecture.md](docs/architecture.md)，公共接口与数据边界见
[docs/api-and-data.md](docs/api-and-data.md)。

## 必须遵守的规则

- 新能力先确认具有跨项目、稳定的复用价值；业务规则、事实来源和产品策略留在消费方。
- 根模块只依赖标准库；需要第三方依赖、重运行时或特定平台语义时，建立或修改恰当的
  子模块，不要把依赖带入根模块。
- 包保持单一职责和无导入环；公开包保留 `doc.go`，公开 API 的风险、默认值、错误与
  生命周期必须可从 GoDoc 和测试判断。
- 网络、磁盘、重试、并发任务等可阻塞调用要接收并传递 `context.Context`；纯内存工具
  不为形式而添加 `context`。
- 错误保留可判断性：使用 `%w` 添加操作上下文，或暴露明确的错误类型/哨兵错误；
  通用包不自行记录日志、吞掉关闭错误或泄露敏感数据。
- goroutine 必须有完成路径；锁内不执行用户回调或网络 I/O；并发对象的关闭、背压和
  重复关闭语义要在 API 与测试中明确。
- 不修改密钥、令牌、个人数据或本地环境文件，也不把它们写入测试、错误或文档。

详见 [docs/standards/go.md](docs/standards/go.md)。

## 开发与验证

在变更所属模块目录中执行。根模块还必须通过 race 检测：

```powershell
gofmt -w <changed-go-files>
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
```

子模块执行前三项与 `go test ./...`；完整模块清单和 CI 对齐方式见
[docs/development.md](docs/development.md)。质量要求与风险测试见
[docs/testing-and-quality.md](docs/testing-and-quality.md)。

## 变更流程

1. 确认改动归属的包和模块，并阅读其测试与公开文档。
2. 以最小 API 变更实现，补充涵盖成功、无效输入、失败和生命周期边界的测试。
3. 运行该模块的格式、vet 和测试；修改根模块并涉及并发时运行 race。
4. 公共行为、模块边界或兼容性变化时，同步更新 README、包注释和 `docs/`。
5. 提交前检查 `git diff`，不混入无关格式化、依赖更新或生成文件。
