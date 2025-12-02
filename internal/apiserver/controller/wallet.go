package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/middleware"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/response"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// WalletController 封装了钱包相关的控制器方法
type WalletController struct {
	walletService service.WalletService
}

// NewWalletController 创建并返回新的 WalletController 实例（依赖注入）
func NewWalletController(walletService service.WalletService) *WalletController {
	return &WalletController{
		walletService: walletService,
	}
}

// CreateWalletRequest 定义创建 HD 钱包的请求体
type CreateWalletRequest struct {
	Password string `json:"password" binding:"required,min=8"`
	ChainID  uint   `json:"chain_id" binding:"required"`
}

// CreateWalletResponse 定义创建钱包的成功响应体
type CreateWalletResponse struct {
	Address  string `json:"address"`
	ChainID  uint   `json:"chain_id"`
	Mnemonic string `json:"mnemonic,omitempty"` // 助记词只在创建时返回
}

// TransferRequest 定义转账交易的请求体
type TransferRequest struct {
	FromAddress string `json:"from_address" binding:"required"`
	ToAddress   string `json:"to_address"   binding:"required"`
	Amount      string `json:"amount"       binding:"required"` // 字符串格式以避免精度问题
	Password    string `json:"password"     binding:"required"`
	ChainID     uint   `json:"chain_id"     binding:"required"`
}

// CreateHDWallet 处理创建新的 HD 钱包请求 (POST /v1/wallets/create)
func (h *WalletController) CreateHDWallet(c *gin.Context) {
	var req CreateWalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效或格式错误")
		return
	}

	// 从中间件中获取用户ID
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户未登录或认证信息无效")
		return
	}

	// 设置 30 秒超时上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	wallet, mnemonic, err := h.walletService.CreateHDWallet(ctx, userID, req.Password, req.ChainID)
	if err != nil {
		// 1. 业务错误映射
		if errors.Is(err, service.ErrChainNotSupported) {
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "不支持的区块链 ID")
			return
		}

		// 2. 内部系统错误
		// 对于其他系统性错误（如 KeyManager/Store 失败），记录日志并返回通用错误
		logger.Logger.Error("Failed to create HD wallet due to internal error",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "钱包创建失败，请稍后重试")
		return
	}

	// 成功响应 (第一次返回助记词，用户应安全备份)
	response.Success(c, http.StatusCreated, CreateWalletResponse{
		Address:  wallet.Address,
		ChainID:  wallet.ChainID,
		Mnemonic: mnemonic, // 仅在创建成功时返回
	}, "HD 钱包创建成功") // 修正：添加 message 参数
}

// Transfer 处理发起转账交易请求 (POST /v1/wallets/transfer)
func (h *WalletController) Transfer(c *gin.Context) {
	var req TransferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求参数无效或格式错误")
		return
	}

	// 忽略用户 ID 检查 (Transfer 依赖于 from_address, 但在实际业务中可能需要验证 from_address 属于当前用户)
	// 这里简化处理，直接调用 service

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	txHash, err := h.walletService.Transfer(
		ctx,
		req.FromAddress,
		req.ToAddress,
		req.Amount,
		req.Password,
		req.ChainID,
	)

	if err != nil {
		// 1. 业务错误映射
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			response.Error(c, http.StatusNotFound, response.CodeResourceNotFound, "发送地址不存在或您无权操作")
			return
		case errors.Is(err, service.ErrPasswordIncorrect):
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "密码错误，无法解锁钱包")
			return
		case errors.Is(err, service.ErrChainNotSupported):
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "不支持的区块链 ID")
			return
		case errors.Is(err, service.ErrInvalidAmount):
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "无效的转账金额")
			return
		case errors.Is(err, service.ErrInsufficientBal), errors.Is(err, service.ErrInsufficientGas):
			// 余额不足和 Gas 不足都映射为 400 Bad Request
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "余额不足以完成交易（包括矿工费）")
			return
		}

		// 2. 内部系统错误 (包括 RPC 失败等)
		logger.Logger.Error("Transaction failed due to internal error",
			zap.String("from", req.FromAddress),
			zap.String("to", req.ToAddress),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "交易处理失败，请稍后重试")
		return
	}

	// 成功响应
	response.Success(c, http.StatusOK, gin.H{
		"tx_hash":  txHash,
		"chain_id": req.ChainID,
	}, "交易发送成功") // any修正：添加 message 参数
}

// GetBalance 处理查询地址余额请求 (GET /v1/wallets/:address/balance)
func (h *WalletController) GetBalance(c *gin.Context) {
	address := c.Param("address")
	chainIDStr := c.Query("chain_id")

	if address == "" || chainIDStr == "" {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "请求缺少地址或链 ID")
		return
	}

	chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "链 ID 格式错误")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	balance, err := h.walletService.GetBalance(ctx, address, uint(chainID))
	if err != nil {
		// 1. 业务错误映射
		if errors.Is(err, service.ErrChainNotSupported) {
			response.Error(c, http.StatusBadRequest, response.CodeInvalidParam, "不支持的区块链 ID")
			return
		}

		// 2. 内部系统错误（如 RPC 连接超时、节点失败）
		logger.Logger.Error("Failed to get balance from blockchain",
			zap.String("address", address),
			zap.Uint64("chain_id", chainID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "查询余额失败，请检查网络或稍后重试")
		return
	}

	// 成功响应
	response.Success(c, http.StatusOK, gin.H{
		"address":     address,
		"chain_id":    chainID,
		"balance_eth": balance, // 余额已在 Service 层转换为 ETH 格式
	}, "余额查询成功")
}
