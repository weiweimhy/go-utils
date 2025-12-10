package ocr

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
)

type TencentOCRConfig struct {
	SecretId  string `toml:"SecretId"`
	SecretKey string `toml:"SecretKey"`
}

type TencentOCR struct {
	*tencentocr.Client
}

func NewTencentOCR(config *TencentOCRConfig) (*TencentOCR, error) {
	credential := common.NewCredential(
		config.SecretId,
		config.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"

	client, err := tencentocr.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent OCR client: %w", err)
	}

	return &TencentOCR{client}, nil
}

func (o *TencentOCR) GetPdfInvoiceData(data *string, isPdf bool) (*tencentocr.VatInvoiceOCRResponse, error) {
	request := tencentocr.NewVatInvoiceOCRRequest()
	request.IsPdf = &isPdf
	request.ImageBase64 = data
	return o.Client.VatInvoiceOCR(request)
}

func (o *TencentOCR) GetOfdInvoiceData(data *string) (*tencentocr.VerifyOfdVatInvoiceOCRResponse, error) {
	request := tencentocr.NewVerifyOfdVatInvoiceOCRRequest()
	request.OfdFileBase64 = data
	return o.Client.VerifyOfdVatInvoiceOCR(request)
}

func (o *TencentOCR) GetGeneralAccurateData(data *string) (*tencentocr.GeneralAccurateOCRResponse, error) {
	request := tencentocr.NewGeneralAccurateOCRRequest()
	request.ImageBase64 = data
	return o.Client.GeneralAccurateOCR(request)
}
