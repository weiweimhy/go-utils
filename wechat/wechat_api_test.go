package wechat

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/weiweimhy/go-utils/v5/httputil"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func TestGetSessionValidatesInput(t *testing.T) {
	_, err := GetSession(context.Background(), "", "secret", "code")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetSessionSuccess(t *testing.T) {
	oldClient := httputil.DefaultHTTPClient
	httputil.DefaultHTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"openid":"oid","session_key":"sk"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	defer func() {
		httputil.DefaultHTTPClient = oldClient
	}()

	session, err := GetSession(context.Background(), "appid", "secret", "code")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.OpenID != "oid" {
		t.Fatalf("expected openid oid, got %q", session.OpenID)
	}
}

func TestGetSessionAPIError(t *testing.T) {
	oldClient := httputil.DefaultHTTPClient
	httputil.DefaultHTTPClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"errcode":40029,"errmsg":"invalid code"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	defer func() {
		httputil.DefaultHTTPClient = oldClient
	}()

	_, err := GetSession(context.Background(), "appid", "secret", "code")
	if err == nil || !strings.Contains(err.Error(), "40029") {
		t.Fatalf("expected api error, got %v", err)
	}
}
