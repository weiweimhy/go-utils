# go-utils/tencentocr

`tencentocr` 是 `go-utils` 的腾讯 OCR 子模块，封装了常见 OCR 调用场景。

## 安装

```bash
go get github.com/weiweimhy/go-utils/v5/tencentocr
```

## 特性

- 封装腾讯 OCR SDK 的常见调用
- 暴露明确的 `Client` 类型
- 将云厂商 SDK 依赖隔离在独立子模块中

## 快速开始

```go
import (
	"context"
	"github.com/weiweimhy/go-utils/v5/tencentocr"
)

func main() {
	client, err := tencentocr.NewClient(tencentocr.Config{
		SecretID:  "secret-id",
		SecretKey: "secret-key",
	})
	if err != nil {
		panic(err)
	}

	imageBase64 := "..."
	_, err = client.GetGeneralAccurateData(context.Background(), &imageBase64)
	if err != nil {
		panic(err)
	}
}
```

## 版本说明

- 模块路径：`github.com/weiweimhy/go-utils/v5/tencentocr`
- 作为重依赖子模块独立维护
- 适合按需引入，避免主模块携带云 SDK 依赖
