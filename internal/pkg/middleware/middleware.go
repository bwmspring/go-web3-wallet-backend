package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
)

// Options 包含了配置中间件所需的所有参数
type Options struct {
	CORSConfig  config.CORSConfig
	LimitConfig config.LimitConfig
	JWTService  any
}

// DefaultMiddlewares 返回一组核心的中间件列表，用于注册到 Gin 引擎。
// 顺序非常关键，遵循洋葱模型，从外层到内层执行。
func DefaultMiddlewares(opts Options) []gin.HandlerFunc {
	middlewares := []gin.HandlerFunc{}

	// 1. Context: 必须是第一个，用于注入 RequestID 和 StartTime，供后续中间件使用。
	middlewares = append(middlewares, Context())

	// 2. Logger: 紧跟 Context 之后，记录请求的整个生命周期。
	middlewares = append(middlewares, Logger())

	// 3. Recovery: 捕获 Panic 并恢复，防止服务崩溃。
	//    需要引入 gin.Recovery() 或自定义的 Recovery 中间件。
	middlewares = append(middlewares, gin.Recovery())

	// 4. CORS: 跨域处理，放在限流之前。
	if opts.CORSConfig.AllowCredentials {
		middlewares = append(middlewares, CORS(opts.CORSConfig))
	}

	// 5. Limit (Rate Limiting): 频率限制，保护核心资源。
	if opts.LimitConfig.Enable {
		// 传递 LimitConfig 到 Limit 函数
		middlewares = append(middlewares, Limit(opts.LimitConfig))
	}

	// TODO: 6. AuthN/AuthZ: 身份验证和授权（例如 JWT 验证，将在下一步实现）

	return middlewares
}

// AuthMiddleware 返回认证中间件，它需要一个 JWTService 实例
func AuthMiddleware(jwtService interface{}) gin.HandlerFunc {
	// 确保传入的是 JWTService
	// 这是一个设计上的取舍，为了保持中间件的通用性，这里进行断言
	if svc, ok := jwtService.(service.JWTService); ok {
		// 7. AuthN/AuthZ: 身份验证
		return JWTAuth(svc)
	}
	// 如果传入的不是 service.JWTService，返回一个拒绝所有请求的中间件
	return func(c *gin.Context) {
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "内部错误：认证服务配置失败")
		c.Abort()
	}
}
