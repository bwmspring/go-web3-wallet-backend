package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/controller"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/middleware"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
)

// RouterConfig holds the dependencies needed to configure the router
type RouterConfig struct {
	Cfg              *config.Config
	JWTService       service.JWTService
	AuthController   *controller.AuthController
	UserController   *controller.UserController
	WalletController *controller.WalletController
}

// NewRouter initializes and returns the configured Gin Engine
func NewRouter(cfg *RouterConfig) *gin.Engine {
	g := gin.New()

	// Install global middleware
	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	// Health check (public)
	g.GET("/ping", healthCheck)

	// Auth routes (public)
	g.POST("/login", cfg.AuthController.Login)
	g.POST("/logout", cfg.AuthController.Logout)
	g.POST("/refresh", cfg.AuthController.Refresh)

	// User registration (public)
	g.POST("/register", cfg.UserController.Register)

	// API v1 route group (requires JWT authentication)
	v1 := g.Group("/v1")
	v1.Use(middleware.JWTAuth(cfg.JWTService))
	{
		// User resource routes
		registerUserRoutes(v1, cfg.UserController)

		// Wallet resource routes
		registerWalletRoutes(v1, cfg.WalletController)
	}

	return g
}

// healthCheck health check handler
func healthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status": "up",
	}, "pong")
}

// registerUserRoutes registers user resource routes
func registerUserRoutes(v1 *gin.RouterGroup, uc *controller.UserController) {
	users := v1.Group("/users")
	{
		// GET /v1/users/profile - Get current user profile
		users.GET("/profile", uc.GetProfile)
	}
}

// registerWalletRoutes registers wallet resource routes
func registerWalletRoutes(v1 *gin.RouterGroup, wc *controller.WalletController) {
	wallets := v1.Group("/wallets")
	{
		// POST /v1/wallets/create - Create HD wallet (requires auth)
		wallets.POST("/create", wc.CreateHDWallet)

		// POST /v1/wallets/transfer - Initiate transfer transaction (requires auth)
		wallets.POST("/transfer", wc.Transfer)

		// GET /v1/wallets/:address/balance?chain_id=... - Get balance (requires auth)
		wallets.GET("/:address/balance", wc.GetBalance)
	}
}
