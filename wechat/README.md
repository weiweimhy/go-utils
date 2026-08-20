# go-utils/wechat

`wechat` 是 `go-utils` 的微信相关子模块，目前提供小程序 `code2Session` 接口的轻量封装。

## 安装

```bash
go get github.com/weiweimhy/go-utils/v6/wechat
```

## 特性

- 原生支持 `context.Context`
- 依赖主模块中的 `httputil.DefaultHTTPClient`
- 保持轻量，不额外引入大型第三方 SDK

## 快速开始

```go
import (
	"context"
	"github.com/weiweimhy/go-utils/v6/wechat"
)

func main() {
	ctx := context.Background()

	session, err := wechat.GetSession(ctx, "appid", "secret", "js-code")
	if err != nil {
		panic(err)
	}

	_ = session.OpenID
}
```

## 版本说明

- 模块路径：`github.com/weiweimhy/go-utils/v6/wechat`
- 依赖根模块 `github.com/weiweimhy/go-utils/v6`
- 发布时需要与根模块兼容的版本策略一并考虑
