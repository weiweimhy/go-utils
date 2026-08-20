package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/weiweimhy/go-utils/v5/httputil"
)

// Session represents a WeChat mini-program session response.
type Session struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// JSCode2SessionURL is retained for compatibility. Call GetSession rather
// than formatting this URL directly so query values are correctly escaped.
const JSCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

const jsCode2SessionEndpoint = "https://api.weixin.qq.com/sns/jscode2session"

// GetSession fetches a WeChat mini-program session and supports context cancellation.
func GetSession(ctx context.Context, appID, secret, jsCode string) (Session, error) {
	if appID == "" || secret == "" || jsCode == "" {
		return Session{}, fmt.Errorf("appID, secret and jsCode are required")
	}

	endpoint, err := url.Parse(jsCode2SessionEndpoint)
	if err != nil {
		return Session{}, fmt.Errorf("parse wechat API endpoint: %w", err)
	}
	query := endpoint.Query()
	query.Set("appid", appID)
	query.Set("secret", secret)
	query.Set("js_code", jsCode)
	query.Set("grant_type", "authorization_code")
	endpoint.RawQuery = query.Encode()

	body, err := httputil.GetBytes(ctx, endpoint.String(), httputil.Options{
		MaxBytes:      1 << 20,
		AllowedStatus: []int{http.StatusOK},
	})
	if err != nil {
		return Session{}, fmt.Errorf("failed to request wechat API: %w", err)
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
