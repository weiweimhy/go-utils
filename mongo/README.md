# go-utils/mongo

`mongo` 是 `go-utils` 的 MongoDB 子模块，提供带默认超时控制的轻量封装。

## 安装

```bash
go get github.com/weiweimhy/go-utils/v5/mongo
```

## 特性

- 默认连接超时和操作超时
- 基于 `context.Context` 的调用方式
- 暴露明确的 `Client` 类型，API 更贴近 Go 习惯

## 快速开始

```go
import (
	"context"
	"github.com/weiweimhy/go-utils/v5/mongo"
)

func main() {
	ctx := context.Background()

	client, err := mongo.NewClient(ctx, mongo.Config{
		URI:          "mongodb://localhost:27017",
		DatabaseName: "app",
	})
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	_, err = client.InsertOne(ctx, "users", map[string]any{"name": "alice"})
	if err != nil {
		panic(err)
	}
}
```

## 版本说明

- 模块路径：`github.com/weiweimhy/go-utils/v5/mongo`
- 作为独立子模块维护
- 对外 API 的不兼容变更应通过新的 major 版本发布
