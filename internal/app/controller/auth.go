package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/internal/app/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/service"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// AuthController 封装了认证相关的控制器
type AuthController struct {
	userService service.UserService
	jwtService  service.JWTService
}

// NewAuthController 创建并返回新的 AuthController 实例
func NewAuthController(userService service.UserService, jwtService service.JWTService) *AuthController {
	return &AuthController{
		userService: userService,
		jwtService:  jwtService,
	}
}

// LoginRequest 定义登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// Login 处理用户登录请求 (POST /login)
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效")
		return
	}

	// 验证用户名和密码
	token, err := ctrl.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码无效")
		return
	}

	// 获取用户信息
	user, err := ctrl.userService.FindUserByUsername(req.Username)
	if err != nil {
		logger.Logger.Error("Login successful but failed to retrieve user data",
			zap.String("username", req.Username), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "登录成功，但无法生成响应")
		return
	}

	// 返回 Token 和用户信息
	response.Success(c, http.StatusOK, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
	}, "登录成功")
}

// Logout 处理用户登出请求 (POST /logout)
func (ctrl *AuthController) Logout(c *gin.Context) {
	// JWT 是无状态的，登出主要由客户端处理（删除 Token）
	// 服务端可以记录登出日志或将 Token 加入黑名单（需要 Redis）
	response.Success(c, http.StatusOK, nil, "登出成功")
}

// Refresh 处理 Token 刷新请求 (POST /refresh)
func (ctrl *AuthController) Refresh(c *gin.Context) {
	// 从请求头或 Body 获取旧 Token
	oldToken := c.GetHeader("Authorization")
	if oldToken == "" {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效")
			return
		}
		oldToken = req.Token
	} else {
		// 移除 "Bearer " 前缀
		if len(oldToken) > 7 && oldToken[:7] == "Bearer " {
			oldToken = oldToken[7:]
		}
	}

	// 刷新 Token
	newToken, err := ctrl.jwtService.RefreshToken(oldToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "Token 刷新失败")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"token": newToken,
	}, "Token 刷新成功")
}
