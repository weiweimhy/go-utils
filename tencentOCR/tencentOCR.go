package tencentOCR

import (
	"os"

	"sync"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
	"go.uber.org/zap"
)

type TencentOCRConfig struct {
	SecretId  string `toml:"SecretId"`
	SecretKey string `toml:"SecretKey"`
}

type TencentOCR struct {
	*ocr.Client
}

var (
	instance *TencentOCR
	url      = "ocr.tencentcloudapi.com"
	once     sync.Once
)

func Init(config *TencentOCRConfig) *TencentOCR {
	once.Do(func() {
		credential := common.NewCredential(
			config.SecretId,
			config.SecretKey,
		)
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = url
		client, _ := ocr.NewClient(credential, "", cpf)

		instance = &TencentOCR{client}
	})

	return instance
}

func GetInstance() *TencentOCR {
	if instance == nil {
		zap.L().Fatal("tencent OCR instance not initialized, you should call Init first")
		os.Exit(1)
	}

	return instance
}

func (o *TencentOCR) GetPdfInvoiceData(data *string, isPdf bool) (*ocr.VatInvoiceOCRResponse, error) {
	request := ocr.NewVatInvoiceOCRRequest()

	request.IsPdf = &isPdf
	request.ImageBase64 = data
	return o.Client.VatInvoiceOCR(request)
}

func (o *TencentOCR) GetOfdInvoiceData(data *string) (*ocr.VerifyOfdVatInvoiceOCRResponse, error) {
	request := ocr.NewVerifyOfdVatInvoiceOCRRequest()

	request.OfdFileBase64 = data
	return o.Client.VerifyOfdVatInvoiceOCR(request)
}

func (o *TencentOCR) GetGeneralAccurateData(data *string) (*ocr.GeneralAccurateOCRResponse, error) {
	request := ocr.NewGeneralAccurateOCRRequest()

	request.ImageBase64 = data
	return o.Client.GeneralAccurateOCR(request)
}
