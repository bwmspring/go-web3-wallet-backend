package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/api"
	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/middleware"
)

// InitRouter 接受配置对象，用于初始化需要依赖配置的服务
func InitRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Ping 路由 (健康检查，公共接口)
	r.GET("/ping", func(c *gin.Context) {
		api.Success(c, http.StatusOK, gin.H{"status": "up"}, "pong")
	})

	authHandler := api.NewAuthHandler(cfg)
	// TODO: walletHandler := api.NewWalletHandler(cfg)

	v1 := r.Group("/api/v1")
	{

		// 1. 公共路由 (认证)
		v1.POST("/register", authHandler.RegisterHandler)
		v1.POST("/login", authHandler.LoginHandler)

		// 2. 受保护路由 (需要 JWT 认证)
		authRequired := v1.Group("/")
		authRequired.Use(middleware.AuthMiddleware(cfg))
		{
			// 示例：测试受保护接口
			authRequired.GET("/profile", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				api.Success(c, http.StatusOK, gin.H{"user_id": userID}, "成功访问受保护资源")
			})

			// TODO: 钱包核心接口
			// authRequired.POST("/wallet/create", walletHandler.CreateWalletHandler)
		}
	}

	return r
}
