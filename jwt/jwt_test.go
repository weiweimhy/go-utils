package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
			name:    "empty key should fail",
			opts:    []Option{},
			wantErr: errs.ErrJWTKeyMissing,
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

// ===================== RSA 测试 =====================

// generateTestRSAKeyPair 生成测试用的 RSA 密钥对
func generateTestRSAKeyPair(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key pair: %v", err)
	}
	return privateKey
}

func TestNewJWT_RSA(t *testing.T) {
	privateKey := generateTestRSAKeyPair(t)

	tests := []struct {
		name    string
		opts    []Option
		wantErr error
	}{
		{
			name: "with private key only",
			opts: []Option{
				WithPrivateKey(privateKey),
			},
			wantErr: nil,
		},
		{
			name: "with public key only",
			opts: []Option{
				WithPublicKey(&privateKey.PublicKey),
			},
			wantErr: nil,
		},
		{
			name: "with both keys",
			opts: []Option{
				WithPrivateKey(privateKey),
				WithPublicKey(&privateKey.PublicKey),
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

func TestRSAGenerateAndValidate(t *testing.T) {
	privateKey := generateTestRSAKeyPair(t)

	// 使用私钥创建 JWT 实例（同时用于签名和验证）
	j, err := NewJWT(WithPrivateKey(privateKey))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	userID := "rsa-user-123"
	extra := map[string]any{"role": "admin", "permissions": []string{"read", "write"}}

	// 生成令牌
	tokens, err := j.Generate(ctx, userID, extra)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Generate() AccessToken is empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("Generate() RefreshToken is empty")
	}

	// 验证访问令牌
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

func TestRSADistributedScenario(t *testing.T) {
	// 模拟分布式场景：
	// - 认证中心使用私钥签发令牌
	// - 业务服务使用公钥验证令牌

	privateKey := generateTestRSAKeyPair(t)

	// 认证中心：使用私钥签发
	authCenter, err := NewJWT(WithPrivateKey(privateKey))
	if err != nil {
		t.Fatalf("NewJWT() for auth center error = %v", err)
	}

	ctx := context.Background()
	userID := "distributed-user"

	// 认证中心生成令牌
	tokens, err := authCenter.Generate(ctx, userID, map[string]any{"service": "api"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 业务服务：只使用公钥验证
	businessService, err := NewJWT(WithPublicKey(&privateKey.PublicKey))
	if err != nil {
		t.Fatalf("NewJWT() for business service error = %v", err)
	}

	// 业务服务验证令牌
	claims, err := businessService.Validate(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Validate() UserID = %v, want %v", claims.UserID, userID)
	}

	// 业务服务不能生成令牌（因为没有私钥）
	_, err = businessService.Generate(ctx, "another-user", nil)
	if !errors.Is(err, errs.ErrJWTPrivateKeyMissing) {
		t.Errorf("Generate() without private key error = %v, want %v", err, errs.ErrJWTPrivateKeyMissing)
	}
}

func TestRSARefresh(t *testing.T) {
	privateKey := generateTestRSAKeyPair(t)

	j, err := NewJWT(WithPrivateKey(privateKey))
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	tokens, err := j.Generate(ctx, "refresh-user", nil)
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

	// 验证新令牌
	claims, err := j.Validate(ctx, newTokens.AccessToken)
	if err != nil {
		t.Fatalf("Validate() new token error = %v", err)
	}

	if claims.UserID != "refresh-user" {
		t.Errorf("Validate() UserID = %v, want %v", claims.UserID, "refresh-user")
	}
}

func TestRSAValidateExpired(t *testing.T) {
	privateKey := generateTestRSAKeyPair(t)

	j, err := NewJWT(
		WithPrivateKey(privateKey),
		WithAccessTokenExpiry(-1*time.Hour), // 已过期
	)
	if err != nil {
		t.Fatalf("NewJWT() error = %v", err)
	}

	ctx := context.Background()
	tokens, err := j.Generate(ctx, "expired-user", nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = j.Validate(ctx, tokens.AccessToken)
	if !errors.Is(err, errs.ErrJWTTokenExpired) {
		t.Errorf("Validate() error = %v, want %v", err, errs.ErrJWTTokenExpired)
	}
}

func TestCrossAlgorithmValidation(t *testing.T) {
	// 测试：使用 HS256 签发的令牌不能被 RS256 验证（反之亦然）
	privateKey := generateTestRSAKeyPair(t)

	hmacJWT, err := NewJWT(WithSecret("test-secret-key-256-bits-long!!"))
	if err != nil {
		t.Fatalf("NewJWT() HMAC error = %v", err)
	}

	rsaJWT, err := NewJWT(WithPrivateKey(privateKey))
	if err != nil {
		t.Fatalf("NewJWT() RSA error = %v", err)
	}

	ctx := context.Background()

	// HMAC 生成的令牌不能被 RSA 验证
	hmacTokens, err := hmacJWT.Generate(ctx, "hmac-user", nil)
	if err != nil {
		t.Fatalf("Generate() HMAC error = %v", err)
	}

	_, err = rsaJWT.Validate(ctx, hmacTokens.AccessToken)
	if err == nil {
		t.Error("Validate() HMAC token with RSA should fail")
	}

	// RSA 生成的令牌不能被 HMAC 验证
	rsaTokens, err := rsaJWT.Generate(ctx, "rsa-user", nil)
	if err != nil {
		t.Fatalf("Generate() RSA error = %v", err)
	}

	_, err = hmacJWT.Validate(ctx, rsaTokens.AccessToken)
	if err == nil {
		t.Error("Validate() RSA token with HMAC should fail")
	}
}

func TestRSAGenerateAndValidate_PEM(t *testing.T) {
	privateKey := generateTestRSAKeyPair(t)

	// 将私钥和公钥转换为 PEM 格式
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	// 使用 PEM 创建实例
	j, err := NewJWT(
		WithPrivateKeyPEM(privatePEM),
		WithPublicKeyPEM(publicPEM),
	)
	if err != nil {
		t.Fatalf("NewJWT() with PEM error = %v", err)
	}

	ctx := context.Background()
	userID := "pem-user"

	tokens, err := j.Generate(ctx, userID, nil)
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
}
