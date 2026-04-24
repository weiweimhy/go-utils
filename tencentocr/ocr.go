package tencentocr

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sdkocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
)

// Config contains Tencent OCR client settings.
type Config struct {
	SecretID  string
	SecretKey string
	Endpoint  string
}

// Client wraps the Tencent OCR SDK client.
type Client struct {
	*sdkocr.Client
}

// NewClient creates a Tencent OCR client with package defaults applied.
func NewClient(config Config) (*Client, error) {
	if config.Endpoint == "" {
		config.Endpoint = "ocr.tencentcloudapi.com"
	}

	credential := common.NewCredential(config.SecretID, config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.Endpoint

	client, err := sdkocr.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent ocr client: %w", err)
	}

	return &Client{client}, nil
}

// GetPDFInvoiceData extracts invoice data from PDF or image content.
func (c *Client) GetPDFInvoiceData(ctx context.Context, data *string, isPDF bool) (*sdkocr.VatInvoiceOCRResponse, error) {
	request := sdkocr.NewVatInvoiceOCRRequest()
	request.IsPdf = &isPDF
	request.ImageBase64 = data
	return c.VatInvoiceOCRWithContext(ctx, request)
}

// GetOFDInvoiceData extracts invoice data from OFD content.
func (c *Client) GetOFDInvoiceData(ctx context.Context, data *string) (*sdkocr.VerifyOfdVatInvoiceOCRResponse, error) {
	request := sdkocr.NewVerifyOfdVatInvoiceOCRRequest()
	request.OfdFileBase64 = data
	return c.VerifyOfdVatInvoiceOCRWithContext(ctx, request)
}

// GetGeneralAccurateData extracts text with Tencent's general accurate OCR API.
func (c *Client) GetGeneralAccurateData(ctx context.Context, data *string) (*sdkocr.GeneralAccurateOCRResponse, error) {
	request := sdkocr.NewGeneralAccurateOCRRequest()
	request.ImageBase64 = data
	return c.GeneralAccurateOCRWithContext(ctx, request)
}
