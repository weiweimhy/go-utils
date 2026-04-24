package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/weiweimhy/go-utils/v4/httputil"
)

// Session represents a WeChat mini-program session response.
type Session struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

const JSCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

// GetSession fetches a WeChat mini-program session and supports context cancellation.
func GetSession(ctx context.Context, appID, secret, jsCode string) (Session, error) {
	if appID == "" || secret == "" || jsCode == "" {
		return Session{}, fmt.Errorf("appID, secret and jsCode are required")
	}

	url := fmt.Sprintf(JSCode2SessionURL, appID, secret, jsCode)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Session{}, err
	}

	// 注意：此处直接使用 http.NewRequestWithContext 配合默认 HTTP 客户端
	rsp, err := httputil.DefaultHTTPClient.Do(req)
	if err != nil {
		return Session{}, fmt.Errorf("failed to request wechat API: %w", err)
	}
	defer func() {
		if rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.StatusCode != http.StatusOK {
		return Session{}, fmt.Errorf("wechat API status: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return Session{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return Session{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if session.ErrCode != 0 {
		return session, fmt.Errorf("wechat API error: [%d] %s", session.ErrCode, session.ErrMsg)
	}

	return session, nil
}
