package tencentOCR

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
	"github.com/weiweimhy/go-utils/logger"
)

// IOCRClient 定义了腾讯 OCR 操作的标准接口。
type IOCRClient interface {
	GetPdfInvoiceData(ctx context.Context, data *string, isPdf bool) (*ocr.VatInvoiceOCRResponse, error)
	GetOfdInvoiceData(ctx context.Context, data *string) (*ocr.VerifyOfdVatInvoiceOCRResponse, error)
	GetGeneralAccurateData(ctx context.Context, data *string) (*ocr.GeneralAccurateOCRResponse, error)
}

type Config struct {
	SecretId  string
	SecretKey string
	Endpoint  string
}

type clientImpl struct {
	*ocr.Client
}

// NewClient 创建并返回满足 IOCRClient 接口的 OCR 实例。
func NewClient(config Config) (IOCRClient, error) {
	if config.Endpoint == "" {
		config.Endpoint = "ocr.tencentcloudapi.com"
	}

	credential := common.NewCredential(config.SecretId, config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.Endpoint

	client, err := ocr.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent ocr client: %w", err)
	}

	return &clientImpl{client}, nil
}

func (c *clientImpl) GetPdfInvoiceData(ctx context.Context, data *string, isPdf bool) (*ocr.VatInvoiceOCRResponse, error) {
	defer logger.Trace(logger.L(), "tencentOCR.GetPdfInvoiceData")()
	request := ocr.NewVatInvoiceOCRRequest()
	request.IsPdf = &isPdf
	request.ImageBase64 = data
	return c.VatInvoiceOCRWithContext(ctx, request)
}

func (c *clientImpl) GetOfdInvoiceData(ctx context.Context, data *string) (*ocr.VerifyOfdVatInvoiceOCRResponse, error) {
	defer logger.Trace(logger.L(), "tencentOCR.GetOfdInvoiceData")()
	request := ocr.NewVerifyOfdVatInvoiceOCRRequest()
	request.OfdFileBase64 = data
	return c.VerifyOfdVatInvoiceOCRWithContext(ctx, request)
}

func (c *clientImpl) GetGeneralAccurateData(ctx context.Context, data *string) (*ocr.GeneralAccurateOCRResponse, error) {
	defer logger.Trace(logger.L(), "tencentOCR.GetGeneralAccurateData")()
	request := ocr.NewGeneralAccurateOCRRequest()
	request.ImageBase64 = data
	return c.GeneralAccurateOCRWithContext(ctx, request)
}
