package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/internal/app/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/service"
	"github.com/bwmspring/go-web3-wallet-backend/middleware"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// UserController 封装了用户身份认证和管理相关的控制器
type UserController struct {
	userService service.UserService
}

// NewUserController 创建并返回新的 UserController 实例（依赖注入）
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// UserRequest 定义注册请求体
type UserRequest struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserData 定义返回给前端的用户数据
type UserData struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// Register 处理用户注册请求 (POST /register)
func (ctrl *UserController) Register(c *gin.Context) {
	var req UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效或格式错误")
		return
	}

	user, err := ctrl.userService.Register(req.Username, req.Password)
	if err != nil {
		if err.Error() == "用户名已存在" {
			response.Error(c, http.StatusConflict, response.CodeResourceExists, "用户名已存在")
			return
		}

		logger.Logger.Error("User registration failed due to internal error", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "注册失败，服务器内部错误")
		return
	}

	// 成功响应
	response.Success(c, http.StatusCreated, UserData{
		ID:       user.ID,
		Username: user.Username,
	}, "用户注册成功")
}

// GetProfile 处理获取当前用户资料请求 (GET /v1/users/profile)
// 需要通过 JWT 认证中间件
func (ctrl *UserController) GetProfile(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "无法获取用户身份")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"user_id": userID,
	}, "成功访问用户资料")
}
