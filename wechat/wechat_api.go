package wechat

import (
	"fmt"
	"github.com/weiweimhy/go-utils/v2/logger"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
)

type WeChatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

const JSCODE2SESSION_URL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

func GetSession(appid, secret, js_code string) (WeChatSession, error) {
	defer logger.Trace(logger.L(), "wechat.GetSession")()

	if appid == "" {
		return WeChatSession{}, logger.InvalidParam(logger.L(), "appid is required", zap.String("param", "appid"))
	}
	if secret == "" {
		return WeChatSession{}, logger.InvalidParam(logger.L(), "secret is required", zap.String("param", "secret"))
	}
	if js_code == "" {
		return WeChatSession{}, logger.InvalidParam(logger.L(), "js_code is required", zap.String("param", "js_code"))
	}

	url := fmt.Sprintf(JSCODE2SESSION_URL, appid, secret, js_code)
	rsp, err := http.Get(url)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("failed to request wechat api: %w", err)
	}

	if rsp.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.L().Warn("failed to close response body",
					zap.String("func", "wechat.GetSession"),
					zap.Error(err),
				)
			}
		}(rsp.Body)
	}

	if rsp.StatusCode != http.StatusOK {
		return WeChatSession{}, fmt.Errorf("wechat api returned non-200 status: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("failed to read response body: %w", err)
	}

	session := WeChatSession{}
	err = sonic.Unmarshal(body, &session)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return session, nil
}
