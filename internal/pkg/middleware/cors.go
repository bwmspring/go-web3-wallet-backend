package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/config"
)

// CORS 跨域资源共享中间件
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	// 生产环境中，强烈建议将 AllowOrigins 设置为具体的白名单，避免使用 "*"
	if len(cfg.AllowOrigins) == 0 {
		// 默认允许所有域，开发时方便，生产环境需谨慎
		cfg.AllowOrigins = []string{"*"}
	}

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,

		// 自定义 Origin 检查，如果你需要更复杂的逻辑
		AllowOriginFunc: func(origin string) bool {
			// 在 AllowOrigins 包含 "*" 时，AllowOriginFunc 不会起作用
			// 只有在 AllowOrigins 被具体设置时，这里才会被调用
			return true
		},
	}

	return cors.New(corsConfig)
}

// DefaultCORSConfig 返回一个合理的默认 CORS 配置
func DefaultCORSConfig() config.CORSConfig {
	return config.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Accept",
			"X-Request-ID",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}
