package wechat

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/weiweimhy/go-utils/v3/httputil"
	"github.com/weiweimhy/go-utils/v3/logger"
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

// GetSession 获取微信会话信息，支持 Context 取消
func GetSession(ctx context.Context, appid, secret, js_code string) (WeChatSession, error) {
	defer logger.Trace(zap.L(), "wechat.GetSession")()

	if appid == "" || secret == "" || js_code == "" {
		return WeChatSession{}, fmt.Errorf("appid, secret and js_code are required")
	}

	url := fmt.Sprintf(JSCODE2SESSION_URL, appid, secret, js_code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return WeChatSession{}, err
	}

	// 注意：此处直接使用 http.NewRequestWithContext 配合 DefaultHttpClient
	rsp, err := httputil.DefaultHttpClient.Do(req)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("failed to request wechat api: %w", err)
	}
	defer func() {
		if rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.StatusCode != http.StatusOK {
		return WeChatSession{}, fmt.Errorf("wechat api status: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return WeChatSession{}, fmt.Errorf("failed to read body: %w", err)
	}

	var session WeChatSession
	if err := sonic.Unmarshal(body, &session); err != nil {
		return WeChatSession{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if session.ErrCode != 0 {
		return session, fmt.Errorf("wechat api error: [%d] %s", session.ErrCode, session.ErrMsg)
	}

	return session, nil
}
