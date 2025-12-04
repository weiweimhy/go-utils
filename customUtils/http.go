package customUtils

import (
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func GetBytesFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			zap.L().Warn("failed to close http response body",
				zap.String("url", url),
				zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed, url: %v, status code: %v", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func GetStringFromUrl(url string) (string, error) {
	data, err := GetBytesFromUrl(url)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
