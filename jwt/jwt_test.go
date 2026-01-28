package jwt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/weiweimhy/go-utils/v3/errs"
	"github.com/weiweimhy/go-utils/v3/logger"
)

func init() {
	logger.Init()
}

// MockJWT 实现了 IJWT 接口，用于测试。
type MockJWT struct {
	IJWT
}

func (m *MockJWT) Generate(ctx context.Context, userID string, extra map[string]any) (*TokenPair, error) {
	return &TokenPair{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}, nil
}

func TestIJWTInterface(t *testing.T) {
	// 验证 MockJWT 是否满足接口
	var _ IJWT = (*MockJWT)(nil)
}

func TestNewJWT(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr error
	}{
		{
			name:    "empty secret should fail",
			opts:    []Option{},
			wantErr: errs.ErrJWTSecretEmpty,
		},
		{
			name: "valid secret should succeed",
			opts: []Option{
				WithSecret("test-secret-key"),
			},
			wantErr: nil,
		},
		{
			name: "with all options",
			opts: []Option{
				WithSecret("test-secret-key"),
				WithAccessTokenExpiry(30 * time.Minute),
				WithRefreshTokenExpiry(24 * time.Hour),
				WithIssuer("test-issuer"),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJWT(tt.opts...)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewJWT() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("NewJWT() unexpected error = %v", err)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	j, err := NewJWT(WithSecret("test-secret-key-256-bits-long!!"))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	tokens, err := j.Generate(ctx, "user123", map[string]any{"role": "admin"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Generate() AccessToken is empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("Generate() RefreshToken is empty")
	}
	if tokens.ExpiresAt.Before(time.Now()) {
		t.Error("Generate() ExpiresAt is in the past")
	}
}

func TestValidate(t *testing.T) {
	j, err := NewJWT(WithSecret("test-secret-key-256-bits-long!!"))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	userID := "user123"
	extra := map[string]any{"role": "admin"}

	tokens, err := j.Generate(ctx, userID, extra)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := j.Validate(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Validate() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.TokenType != TokenTypeAccess {
		t.Errorf("Validate() TokenType = %v, want %v", claims.TokenType, TokenTypeAccess)
	}
}

func TestValidateExpired(t *testing.T) {
	j, err := NewJWT(
		WithSecret("test-secret-key-256-bits-long!!"),
		WithAccessTokenExpiry(-1*time.Hour), // 已过期
	)
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	tokens, err := j.Generate(ctx, "user123", nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = j.Validate(ctx, tokens.AccessToken)
	if !errors.Is(err, errs.ErrJWTTokenExpired) {
		t.Errorf("Validate() error = %v, want %v", err, errs.ErrJWTTokenExpired)
	}
}

func TestValidateInvalid(t *testing.T) {
	j, err := NewJWT(WithSecret("test-secret-key-256-bits-long!!"))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "malformed token",
			token:   "not-a-valid-token",
			wantErr: errs.ErrJWTTokenMalformed,
		},
		{
			name:    "invalid signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlcjEyMyJ9.invalid",
			wantErr: errs.ErrJWTTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.Validate(ctx, tt.token)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	j, err := NewJWT(WithSecret("test-secret-key-256-bits-long!!"))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	tokens, err := j.Generate(ctx, "user123", nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 使用 refresh token 刷新
	newTokens, err := j.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if newTokens.AccessToken == "" {
		t.Error("Refresh() AccessToken is empty")
	}
	if newTokens.RefreshToken == "" {
		t.Error("Refresh() RefreshToken is empty")
	}

	// 使用 access token 刷新应该失败
	_, err = j.Refresh(ctx, tokens.AccessToken)
	if !errors.Is(err, errs.ErrJWTTokenInvalid) {
		t.Errorf("Refresh() with access token error = %v, want %v", err, errs.ErrJWTTokenInvalid)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("DefaultConfig() AccessTokenExpiry = %v, want %v", cfg.AccessTokenExpiry, 15*time.Minute)
	}
	if cfg.RefreshTokenExpiry != 7*24*time.Hour {
		t.Errorf("DefaultConfig() RefreshTokenExpiry = %v, want %v", cfg.RefreshTokenExpiry, 7*24*time.Hour)
	}
	if cfg.Issuer != "go-utils" {
		t.Errorf("DefaultConfig() Issuer = %v, want %v", cfg.Issuer, "go-utils")
	}
}
