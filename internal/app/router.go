package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/internal/app/controller"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/response"
	"github.com/bwmspring/go-web3-wallet-backend/middleware"
)

// InitRouter 初始化并返回配置好的 Gin Engine
func (a *App) InitRouter() *gin.Engine {
	g := gin.New()

	// 安装全局中间件
	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	// 健康检查（公开）
	g.GET("/ping", healthCheck)

	// 认证相关路由（公开）
	g.POST("/login", a.authController.Login)
	g.POST("/logout", a.authController.Logout)
	g.POST("/refresh", a.authController.Refresh)

	// 用户注册（公开）
	g.POST("/register", a.userController.Register)

	// API v1 路由组（需要 JWT 认证）
	v1 := g.Group("/v1")
	v1.Use(middleware.JWTAuth(a.jwtService))
	{
		// 用户资源路由
		registerUserRoutes(v1, a.userController)

		// TODO: 未来扩展 - 钱包资源路由
		// registerWalletRoutes(v1, a.walletController)
	}

	return g
}

// healthCheck 健康检查处理器
func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status": "up",
	}, "pong")
}

// registerUserRoutes 注册用户资源路由
func registerUserRoutes(v1 *gin.RouterGroup, uc *controller.UserController) {
	users := v1.Group("/users")
	{
		// GET /v1/users/profile - 获取当前用户资料
		users.GET("/profile", uc.GetProfile)

		// TODO: 未来扩展完整的 CRUD 操作
		// users.GET("", uc.List)                           // 列表
		// users.GET(":id", uc.Get)                         // 详情
		// users.PUT(":id", uc.Update)                      // 更新
		// users.DELETE(":id", uc.Delete)                   // 删除
		// users.PUT(":id/change-password", uc.ChangePassword) // 修改密码
	}
}

// registerWalletRoutes 注册钱包资源路由（未来实现）
// func registerWalletRoutes(v1 *gin.RouterGroup, wc *controller.WalletController) {
// 	wallets := v1.Group("/wallets")
// 	{
// 		wallets.POST("", wc.Create)                          // 创建钱包
// 		wallets.GET("", wc.List)                             // 钱包列表
// 		wallets.GET(":address/balance", wc.GetBalance)       // 获取余额
// 		wallets.POST("/transfer", wc.Transfer)               // 转账
// 	}
// }
