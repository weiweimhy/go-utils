# 开发与验证

## 环境

仓库和 CI 均使用 Go 1.27.2。根目录的 `go.work` 将根模块和十个扩展模块放入同一
工作区，因此日常开发应从仓库根目录或目标模块目录执行 Go 命令。

本仓库是库项目，没有本地启动服务、环境变量模板或应用构建命令。`go test` 在验证时
同时完成所测包的编译。

## 模块定位

根模块目录是 `.`。扩展模块目录为：

- `download`
- `epub`
- `htmlutil`
- `jwt`
- `localdb`
- `logger`
- `mongo`
- `postgres`
- `tencentocr`
- `wechat`

每次变更先进入其所属模块目录。跨根模块和子模块的改动应分别验证每个受影响模块。

## 日常命令

以下命令均由现有 GitHub Actions 工作流执行；根模块和 `postgres` 额外执行 race 检测：

```powershell
gofmt -l .
go vet ./...
go test ./...
```

根模块的并发安全检查：

```powershell
go test -race ./...
```

`postgres` 模块也运行 race 检测。其集成测试使用 CI 提供的临时 PostgreSQL 18 服务和
测试专用账号、数据库，通过 `POSTGRES_TEST_DSN` 配置连接。在 `postgres` 目录中关闭
工作区模式，以独立模块执行（PowerShell 环境设置仅对当前进程生效）：

```powershell
$env:POSTGRES_TEST_DSN = "postgres://go_utils_test:go_utils_test@localhost:5432/go_utils_test?sslmode=disable"
$env:GOWORK = "off"
go test -race -run '^TestIntegration' ./...
```

不要将真实服务、账号或凭据用于集成测试。CI 的服务容器只用于测试，服务就绪后才运行该
命令。未设置 `POSTGRES_TEST_DSN` 时，普通 `go test` 会跳过真实数据库集成测试。

修改 Go 文件后，先格式化已改文件，再运行上述检查：

```powershell
gofmt -w <changed-go-files>
```

CI 的格式检查以 `gofmt -l .` 没有输出为通过条件。根模块和每个扩展模块都要独立运行
格式、vet 和测试；只修改根模块时不需要为了形式而重跑无关的扩展模块。

## 依赖与工作区

- 在根模块新增依赖前，先确认功能无法使用标准库或既有根模块包完成。
- 需要外部依赖时，将能力放在职责匹配的子模块，并只更新该模块的 `go.mod` 与 `go.sum`。
- 新增本地子模块时，同步更新根目录 `go.work`；使用根模块的子模块应保持其明确的
  本地 `replace github.com/weiweimhy/go-utils/v6 => ..` 开发替换规则。
- 不要用根模块的依赖变更替代子模块的依赖声明，也不要把无关的 tidy 结果混入功能提交。

## 排错边界

若 Go 命令的结果与预期不一致，先确认当前目录属于上面的模块清单，并检查变更是否同时
影响了根模块与子模块。PostgreSQL 集成测试使用上文的隔离服务；涉及 MongoDB、腾讯云、
微信等其他外部系统的行为，应在对应子模块中以可控依赖和测试边界验证，不能假定本地
服务已经存在。
