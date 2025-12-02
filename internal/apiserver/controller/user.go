package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/internal/pkg/middleware"
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
		// 参数校验失败，可能是格式错误或缺失字段
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效或格式错误")
		return
	}

	user, err := ctrl.userService.Register(req.Username, req.Password)
	if err != nil {
		// 1. 捕获业务逻辑错误 (http.StatusConflict / 409)
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.Error(c, http.StatusConflict, response.CodeResourceExists, "用户名已存在，请更换")
			return
		}

		// 2. 捕获系统内部错误 (http.StatusInternalServerError / 500)
		// 捕获 PasswordHashFailed 或 StoreOperationFailed 这种无法修复的错误
		if errors.Is(err, service.ErrPasswordHashFailed) || errors.Is(err, service.ErrStoreOperationFailed) {
			// 在 Controller 边界记录日志，并提供完整的上下文
			logger.Logger.Error("User registration failed due to internal system error",
				zap.String("username", req.Username),
				zap.Error(err),
			)
			// 返回给用户通用的错误信息
			response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "注册失败，服务器内部错误，请稍后重试")
			return
		}

		// 3. 捕获未预期的其他错误 (作为 500 处理)
		// 如果 service 层返回了未映射的错误，也记录并返回通用错误
		logger.Logger.Error("User registration failed due to unexpected error",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "注册失败，发生未知错误")
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
		// 通常 GetUserID 失败（如 token 解析失败）应该由中间件处理，
		// 但如果 GetUserID 仍返回错误，Controller 需处理它。
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "无法获取用户身份")
		return
	}

	// 此时应通过 FindByID 查询用户详细信息，这里为了简化只返回 ID
	// user, err := ctrl.userService.FindByID(userID)

	response.Success(c, http.StatusOK, gin.H{
		"user_id": userID,
	}, "成功访问用户资料")
}
