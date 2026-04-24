package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/weiweimhy/go-utils/v4/errs"
)

// TokenType 定义令牌类型
type TokenType string

const (
	// TokenTypeAccess 访问令牌
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh 刷新令牌
	TokenTypeRefresh TokenType = "refresh"
)

// TokenPair 包含一对生成的令牌及其过期信息。
type TokenPair struct {
	AccessToken  string    `json:"access_token"`  // 访问令牌，用于常规鉴权
	RefreshToken string    `json:"refresh_token"` // 刷新令牌，用于续期
	ExpiresAt    time.Time `json:"expires_at"`    // AccessToken 的过期时间
}

// Claims 自定义 JWT 声明，包含用户信息和标准注册声明。
type Claims struct {
	UserID               string         `json:"user_id"`         // 用户唯一标识
	TokenType            TokenType      `json:"token_type"`      // 令牌类型 (access/refresh)
	Extra                map[string]any `json:"extra,omitempty"` // 业务自定义扩展信息
	jwt.RegisteredClaims                // JWT 标准声明 (iss, sub, exp, iat, etc.)
}

// Config 包含 JWT 模块的配置参数。
type Config struct {
	Secret             string        `yaml:"secret"`               // HMAC 签名密钥，必须保密
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`  // 访问令牌有效期，默认 15 分钟
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"` // 刷新令牌有效期，默认 7 天
	Issuer             string        `yaml:"issuer"`               // 签发者名称，默认 "go-utils"

	// RSA 密钥对，用于非对称签名（适合分布式场景）
	PrivateKey *rsa.PrivateKey `yaml:"-"` // RSA 私钥，用于签名（认证中心持有）
	PublicKey  *rsa.PublicKey  `yaml:"-"` // RSA 公钥，用于验证（业务服务持有）

	// SigningMethod 签名算法，默认根据密钥类型自动选择
	// - 提供 Secret 时默认 HS256
	// - 提供 PrivateKey/PublicKey 时默认 RS256
	SigningMethod jwt.SigningMethod `yaml:"-"`
}

// DefaultConfig 返回带有工业级合理默认值的配置。
func DefaultConfig() Config {
	return Config{
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "go-utils",
	}
}

// Option 是配置选项函数类型。
type Option func(*Config)

// WithSecret 设置 HMAC 签名密钥。
func WithSecret(secret string) Option {
	return func(c *Config) {
		c.Secret = secret
	}
}

// WithAccessTokenExpiry 设置访问令牌过期时间。
func WithAccessTokenExpiry(d time.Duration) Option {
	return func(c *Config) {
		c.AccessTokenExpiry = d
	}
}

// WithRefreshTokenExpiry 设置刷新令牌过期时间。
func WithRefreshTokenExpiry(d time.Duration) Option {
	return func(c *Config) {
		c.RefreshTokenExpiry = d
	}
}

// WithIssuer 设置签发者。
func WithIssuer(issuer string) Option {
	return func(c *Config) {
		c.Issuer = issuer
	}
}

// WithPrivateKey 设置 RSA 私钥（用于签名）。
func WithPrivateKey(key *rsa.PrivateKey) Option {
	return func(c *Config) {
		c.PrivateKey = key
	}
}

// WithPublicKey 设置 RSA 公钥（用于验证）。
func WithPublicKey(key *rsa.PublicKey) Option {
	return func(c *Config) {
		c.PublicKey = key
	}
}

// WithPrivateKeyPEM 从 PEM 格式的基础数据设置 RSA 私钥。
func WithPrivateKeyPEM(keyData []byte) Option {
	return func(c *Config) {
		block, _ := pem.Decode(keyData)
		if block == nil {
			return
		}

		// 尝试多种 PEM 格式 (PKCS#1, PKCS#8)
		if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			c.PrivateKey = key
			return
		}

		if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if rsaKey, ok := key.(*rsa.PrivateKey); ok {
				c.PrivateKey = rsaKey
			}
		}
	}
}

// WithPublicKeyPEM 从 PEM 格式的数据设置 RSA 公钥。
func WithPublicKeyPEM(keyData []byte) Option {
	return func(c *Config) {
		block, _ := pem.Decode(keyData)
		if block == nil {
			return
		}

		// 尝试多种格式 (PKIX, PKCS#1)
		if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
			if pubKey, ok := key.(*rsa.PublicKey); ok {
				c.PublicKey = pubKey
			}
			return
		}

		if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
			c.PublicKey = key
		}
	}
}

// WithSigningMethod 设置签名算法。
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(c *Config) {
		c.SigningMethod = method
	}
}

// JWT provides token generation, validation, and refresh helpers.
type JWT struct {
	cfg Config
}

// NewJWT 创建并返回 JWT 实例。
// 支持两种模式：
//   - HMAC 模式：提供 Secret，使用 HS256 算法
//   - RSA 模式：提供 PrivateKey 和/或 PublicKey，使用 RS256 算法
func NewJWT(opts ...Option) (*JWT, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// 自动选择签名方法
	if cfg.SigningMethod == nil {
		if cfg.PrivateKey != nil || cfg.PublicKey != nil {
			cfg.SigningMethod = jwt.SigningMethodRS256
		} else if cfg.Secret != "" {
			cfg.SigningMethod = jwt.SigningMethodHS256
		}
	}

	// 验证密钥配置
	if cfg.Secret == "" && cfg.PrivateKey == nil && cfg.PublicKey == nil {
		return nil, errs.ErrJWTKeyMissing
	}

	return &JWT{cfg: cfg}, nil
}

// Generate 生成访问令牌和刷新令牌对。
func (j *JWT) Generate(ctx context.Context, userID string, extra map[string]any) (*TokenPair, error) {
	// 检查是否有签名密钥
	if j.cfg.Secret == "" && j.cfg.PrivateKey == nil {
		return nil, errs.ErrJWTPrivateKeyMissing
	}

	now := time.Now()
	accessExpiresAt := now.Add(j.cfg.AccessTokenExpiry)
	refreshExpiresAt := now.Add(j.cfg.RefreshTokenExpiry)

	// 生成访问令牌
	accessToken, err := j.generateToken(userID, TokenTypeAccess, extra, accessExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := j.generateToken(userID, TokenTypeRefresh, nil, refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// Validate 验证令牌并返回 Claims。
func (j *JWT) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 根据 token 的签名方法返回对应的验证密钥
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			if j.cfg.Secret == "" {
				return nil, fmt.Errorf("HMAC secret not configured")
			}
			return []byte(j.cfg.Secret), nil
		case *jwt.SigningMethodRSA:
			if j.cfg.PublicKey != nil {
				return j.cfg.PublicKey, nil
			}
			// 如果只有私钥，可以从私钥提取公钥
			if j.cfg.PrivateKey != nil {
				return &j.cfg.PrivateKey.PublicKey, nil
			}
			return nil, errs.ErrJWTPublicKeyMissing
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errs.ErrJWTTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errs.ErrJWTTokenMalformed
		}
		return nil, fmt.Errorf("%w: %v", errs.ErrJWTTokenInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errs.ErrJWTClaimsInvalid
	}

	return claims, nil
}

// Refresh 使用刷新令牌获取新的令牌对。
func (j *JWT) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := j.Validate(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errs.ErrJWTTokenInvalid
	}

	return j.Generate(ctx, claims.UserID, claims.Extra)
}

// generateToken 生成单个令牌。
func (j *JWT) generateToken(userID string, tokenType TokenType, extra map[string]any, expiresAt time.Time) (string, error) {
	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		Extra:     extra,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(j.cfg.SigningMethod, claims)

	// 根据签名方法选择密钥
	switch j.cfg.SigningMethod.(type) {
	case *jwt.SigningMethodHMAC:
		return token.SignedString([]byte(j.cfg.Secret))
	case *jwt.SigningMethodRSA:
		if j.cfg.PrivateKey == nil {
			return "", errs.ErrJWTPrivateKeyMissing
		}
		return token.SignedString(j.cfg.PrivateKey)
	default:
		return "", fmt.Errorf("unsupported signing method: %v", j.cfg.SigningMethod.Alg())
	}
}
