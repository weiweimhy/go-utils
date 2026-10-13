# go-utils/postgres

`postgres` 是 `go-utils` 的独立 PostgreSQL 子模块，使用 pgx/v5 创建并验证
原生 `pgxpool.Pool`。模块路径为
`github.com/weiweimhy/go-utils/postgres/v6`，最低 Go 版本为 1.27.2。

## 安装

```bash
go get github.com/weiweimhy/go-utils/postgres/v6
```

## 创建和复用连接池

应用启动时创建一次连接池，并在应用生命周期内复用。不要为每个请求创建一个新池。

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/weiweimhy/go-utils/postgres/v6"
)

func main() {
    ctx := context.Background()
    pool, err := postgres.NewPool(ctx, postgres.Config{
        ConnString: os.Getenv("DATABASE_URL"),
    })
    if err != nil {
        log.Printf("create PostgreSQL pool: %v", err)
        return
    }
    defer pool.Close()

    var answer int
    if err := pool.QueryRow(ctx, "select 42").Scan(&answer); err != nil {
        log.Printf("query database: %v", err)
        return
    }
    log.Println(answer)

    // 将同一个 pool 复用给应用服务和请求处理器。
}
```

`Config.StartupTimeout` 为零时使用 10 秒；负数无效。该超时只约束创建池和启动时的
`Ping`，不会限制后续查询。`NewPool` 必须收到非 nil 的 context。启动连接或 Ping 失败时，
函数会关闭已经创建的池。

`ConnString` 使用 pgx 连接串格式。连接池参数由 pgxpool 的原生 `pool_*` 参数解析，
例如 `pool_max_conns`、`pool_min_conns` 和 `pool_min_idle_conns`。健康检查周期必须大于零；
最小连接数和最小空闲连接数不能为负，也不能超过最大连接数。空连接串沿用 pgx 的环境变量
语义，包括 `PGHOST`、`PGPORT`、`PGDATABASE`、`PGUSER` 和 `PGPASSWORD`。

## 错误和资源释放

无效配置返回可由 `errors.Is` 判断的 `postgres.ErrInvalidConfig`；nil context 返回
`postgres.ErrNilContext`。`NewPool` 返回的错误文本经过固定脱敏，并保留底层错误供
`errors.Is` 和 `errors.As` 检查。解包后的驱动错误可能含连接信息，包括连接串中的凭据。

返回值是原生 `pgxpool.Pool`。`Close` 无返回值、无 context，按原生实现同步等待已借出的连接
和事务释放，可以重复调用。查询返回的 `Rows` 使用完后必须关闭，连接才能归还池中。

## 可选集成测试

设置 `POSTGRES_TEST_DSN` 可启用 `TestIntegrationPostgres`。测试只执行 `SELECT 1`；未设置该
环境变量时跳过，不需要真实数据库即可运行模块单元测试。
