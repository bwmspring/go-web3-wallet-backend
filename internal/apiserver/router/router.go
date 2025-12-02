package router

import (
	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/controller"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/internal/pkg/middleware"
)

// RouterConfig holds the dependencies needed to configure the router
type RouterConfig struct {
	ServerCfg   *config.ServerConfig
	CORSConfig  *config.CORSConfig
	LimitConfig *config.LimitConfig

	JWTService service.JWTService

	AuthController   *controller.AuthController
	UserController   *controller.UserController
	WalletController *controller.WalletController
}

// NewRouter initializes and returns the configured Gin Engine
func NewRouter(cfg *RouterConfig) *gin.Engine {
	r := gin.New()

	opts := middleware.Options{
		CORSConfig:  *cfg.CORSConfig,
		LimitConfig: *cfg.LimitConfig,
	}
	r.Use(middleware.DefaultMiddlewares(opts)...)

	// 1. 公共路由 (无需认证): /api/v1/auth/*
	publicV1 := r.Group("/api/v1")
	{
		publicV1.POST("/auth/login", cfg.AuthController.Login)
		publicV1.POST("/auth/logout", cfg.AuthController.Logout)
		publicV1.POST("/auth/refresh", cfg.AuthController.Refresh)

		publicV1.POST("/users/register", cfg.UserController.Register)
	}

	privateV1 := r.Group("/api/v1")
	privateV1.Use(middleware.AuthMiddleware(cfg.JWTService))
	{
		privateV1.GET("/users/profile", cfg.UserController.GetProfile)

		privateV1.POST("/wallet/create", cfg.WalletController.CreateHDWallet)
		privateV1.GET("/wallet/:address/balance", cfg.WalletController.GetBalance)
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})

	return r
}
