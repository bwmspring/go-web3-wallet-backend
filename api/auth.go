package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
	"github.com/bwmspring/go-web3-wallet-backend/service"
)

// AuthHandler 封装了认证相关的 API 处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建并返回新的 AuthHandler 实例。
func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(cfg),
	}
}

// AuthRequest 定义注册和登录的通用请求体
type AuthRequest struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// AuthData 定义返回给前端的用户数据
type AuthData struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token,omitempty"`
}

// RegisterHandler 处理用户注册请求 (POST /api/v1/register)
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidParam, "请求参数无效或格式错误")
		return
	}

	user, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		if err.Error() == "用户名已存在" {
			Error(c, http.StatusConflict, CodeResourceExists, "用户名已存在")
			return
		}

		logger.Logger.Error("Auth registration failed due to internal error", zap.Error(err))
		Error(c, http.StatusInternalServerError, CodeInternalError, "注册失败，服务器内部错误")
		return
	}

	// 成功响应
	Success(c, http.StatusCreated, AuthData{
		ID:       user.ID,
		Username: user.Username,
	}, "用户注册成功")
}

// LoginHandler 处理用户登录请求 (POST /api/v1/login)
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req AuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, CodeInvalidParam, "请求参数无效")
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		Error(c, http.StatusUnauthorized, CodeUnauthorized, "用户名或密码无效")
		return
	}

	// 重新查询用户以获取 ID 等信息
	user, findErr := h.authService.FindUserByUsername(req.Username)
	if findErr != nil {
		logger.Logger.Error("Login successful but failed to retrieve user data for response",
			zap.String("username", req.Username), zap.Error(findErr))
		Error(c, http.StatusInternalServerError, CodeInternalError, "登录成功，但无法生成响应")
		return
	}

	// 成功响应 (包含 Token)
	Success(c, http.StatusOK, AuthData{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	}, "登录成功")
}
