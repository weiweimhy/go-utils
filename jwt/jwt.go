package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/weiweimhy/go-utils/v3/errs"
	"github.com/weiweimhy/go-utils/v3/logger"
	"go.uber.org/zap"
)

// TokenType 定义令牌类型
type TokenType string

const (
	// TokenTypeAccess 访问令牌
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh 刷新令牌
	TokenTypeRefresh TokenType = "refresh"
)

// IJWT 定义了 JWT 操作的标准接口，方便外部 Mock 测试。
type IJWT interface {
	// Generate 生成访问令牌和刷新令牌对。
	Generate(ctx context.Context, userID string, extra map[string]any) (*TokenPair, error)
	// Validate 验证令牌并返回 Claims。
	Validate(ctx context.Context, token string) (*Claims, error)
	// Refresh 使用刷新令牌获取新的令牌对。
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// TokenPair 包含访问令牌和刷新令牌。
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims 自定义 JWT 声明。
type Claims struct {
	UserID    string         `json:"user_id"`
	TokenType TokenType      `json:"token_type"`
	Extra     map[string]any `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// Config 包含 JWT 的配置。
type Config struct {
	Secret             string        `yaml:"secret"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	Issuer             string        `yaml:"issuer"`
}

// DefaultConfig 返回带有合理默认值的配置。
func DefaultConfig() Config {
	return Config{
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "go-utils",
	}
}

// Option 是配置选项函数类型。
type Option func(*Config)

// WithSecret 设置签名密钥。
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

type jwtImpl struct {
	cfg Config
}

// NewJWT 创建并返回满足 IJWT 接口的实例。
func NewJWT(opts ...Option) (IJWT, error) {
	defer logger.Trace(logger.L(), "jwt.NewJWT")()

	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Secret == "" {
		logger.L().Error("invalid param",
			zap.String("error", "invalid_param"),
			zap.String("param", "secret"),
			zap.String("func", "jwt.NewJWT"))
		return nil, errs.ErrJWTSecretEmpty
	}

	return &jwtImpl{cfg: cfg}, nil
}

// Generate 生成访问令牌和刷新令牌对。
func (j *jwtImpl) Generate(ctx context.Context, userID string, extra map[string]any) (*TokenPair, error) {
	defer logger.Trace(logger.L(), "jwt.Generate", zap.String("userID", userID))()

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
func (j *jwtImpl) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	defer logger.Trace(logger.L(), "jwt.Validate")()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.cfg.Secret), nil
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
func (j *jwtImpl) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	defer logger.Trace(logger.L(), "jwt.Refresh")()

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
func (j *jwtImpl) generateToken(userID string, tokenType TokenType, extra map[string]any, expiresAt time.Time) (string, error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.Secret))
}
