package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/model"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

var (
	ErrTokenSigningMethodInvalid = errors.New("invalid signing method")
	ErrTokenInvalidOrExpired     = errors.New("authentication token is invalid or expired")
	ErrTokenGenerationFailed     = errors.New("failed to generate authentication token")
)

// JWTClaims 定义了 JWT 有效载荷中应包含的自定义信息
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// CustomClaims 扩展 jwt.RegisteredClaims 以包含自定义信息
type CustomClaims struct {
	JWTClaims
	jwt.RegisteredClaims
}

// JWTService 定义了管理 JSON Web Token (JWT) 的核心契约
type JWTService interface {
	GenerateToken(user *model.User) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
	RefreshToken(token string) (string, error)
}

// jwtService 是 JWTService 接口的具体实现。
type jwtService struct {
	secretKey []byte
	duration  time.Duration
}

// NewJWTService 创建并返回一个新的 JWTService 实例。
func NewJWTService(cfg *config.Config) JWTService {
	duration, err := time.ParseDuration(cfg.Server.JWTDuration)
	if err != nil {
		logger.Logger.Fatal(
			"JWT duration configuration is invalid",
			zap.String("duration_str", cfg.Server.JWTDuration),
			zap.Error(err),
		)
	}

	if len(cfg.Server.JWTSecret) < 32 {
		logger.Logger.Warn("JWT secret key is too short. Use a long, random string in production.")
	}

	return &jwtService{
		secretKey: []byte(cfg.Server.JWTSecret),
		duration:  duration,
	}
}

// GenerateToken 实现 JWTService 接口
func (s *jwtService) GenerateToken(user *model.User) (string, error) {
	claims := CustomClaims{
		JWTClaims: JWTClaims{
			UserID:   user.ID,
			Username: user.Username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.duration)),
			Issuer:    "go-web3-wallet-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		logger.Logger.Error("Error signing JWT token", zap.Uint("user_id", user.ID), zap.Error(err))
		return "", fmt.Errorf("%w: failed to sign token: %s", ErrTokenGenerationFailed, err.Error())
	}

	return signedToken, nil
}

// ValidateToken 实现 JWTService 接口
func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenSigningMethodInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		logger.Logger.Debug("Token parsing/validation error", zap.Error(err))
		// 检查 Go-JWT 的特定错误类型，以提供更清晰的业务错误
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) ||
			errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrTokenInvalidOrExpired
		}

		// 包装所有其他解析错误
		return nil, fmt.Errorf("%w: %s", ErrTokenInvalidOrExpired, err.Error())
	}

	// token.Valid 检查了 ExpiresAt 和 IssuedAt 等标准声明
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return &claims.JWTClaims, nil
	}

	// 最终的通用失败情况
	return nil, ErrTokenInvalidOrExpired
}

// RefreshToken 刷新 JWT Token
func (s *jwtService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 根据用户 ID 重新生成 Token
	user := &model.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	return s.GenerateToken(user)
}
