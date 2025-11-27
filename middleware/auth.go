package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/api"
	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
	"github.com/bwmspring/go-web3-wallet-backend/service"
)

// UserIDKey 是在 Gin Context 中存储用户 ID 的 Key
const UserIDKey = "userID"

// AuthMiddleware 创建 JWT 认证中间件。
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// 确保 JWTService 实例只创建一次
	jwtService := service.NewJWTService(cfg)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			api.Error(c, http.StatusUnauthorized, api.CodeUnauthorized, "请求头缺少 Authorization 认证信息")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		// 检查格式是否是 "Bearer <token>"
		if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
			api.Error(c, http.StatusUnauthorized, api.CodeUnauthorized, "认证格式错误，应为 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 Token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Logger.Debug("Token validation failed", zap.Error(err))
			api.Error(c, http.StatusUnauthorized, api.CodeUnauthorized, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// 核心：将 UserID 存入 Gin Context
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

// GetUserID 从 Gin Context 中提取当前认证用户的 ID。
func GetUserID(c *gin.Context) (uint, error) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		logger.Logger.Error("Fatal: UserID not found in context after AuthMiddleware")
		return 0, errors.New("内部错误：无法获取用户身份")
	}

	userID, ok := val.(uint)
	if !ok {
		logger.Logger.Error("Fatal: UserID context value is not uint type")
		return 0, errors.New("内部错误：用户身份类型错误")
	}

	return userID, nil
}
