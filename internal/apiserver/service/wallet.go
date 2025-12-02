package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/model"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/conversion"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/crypto"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/web3client"
)

var (
	// 业务层公共错误
	ErrWalletNotFound    = errors.New("wallet not found or insufficient permission")
	ErrPasswordIncorrect = errors.New("wallet password incorrect")
	ErrChainNotSupported = errors.New("chain ID is not supported")

	// Transfer 专用错误
	ErrInvalidAmount   = errors.New("invalid or malformed transfer amount")
	ErrInsufficientGas = errors.New("insufficient balance to cover gas fee")
	ErrInsufficientBal = errors.New("insufficient balance for transfer amount")
)

// WalletStore 定义了钱包数据存储的接口 (DIP: 由 service 层定义)
type WalletStore interface {
	// CreateWallet 在数据库中创建一个新的钱包记录
	CreateWallet(ctx context.Context, wallet *model.Wallet, mnemonic *model.MnemonicSeed) error

	// GetWalletByUserID 根据用户ID和链ID获取钱包
	GetWalletByUserID(ctx context.Context, userID uint, chainID uint) (*model.Wallet, error)

	// FindEncryptedKeyByAddress 根据地址查找加密后的私钥（Keystore）
	FindEncryptedKeyByAddress(ctx context.Context, address string) (string, error)
}

// WalletService 定义了钱包模块的业务逻辑接口
type WalletService interface {
	// CreateHDWallet 生成助记词、派生地址、创建Keystore并存储
	CreateHDWallet(ctx context.Context, userID uint, password string, chainID uint) (*model.Wallet, string, error)

	// Transfer 发起一笔链上交易
	Transfer(
		ctx context.Context,
		fromAddress string,
		toAddress string,
		amount string,
		password string,
		chainID uint,
	) (string, error) // 返回 txHash

	// GetBalance 查询指定地址在指定链上的余额
	GetBalance(ctx context.Context, address string, chainID uint) (string, error)
}

// walletService 实现了 WalletService 接口
type walletService struct {
	store         WalletStore
	keyManager    crypto.KeyManager
	clientManager web3client.ClientManager // 依赖注入的 Web3 客户端管理器
	cfg           *config.Config
}

var _ WalletService = (*walletService)(nil)

// NewWalletService 创建并返回一个新的 WalletService 实例
func NewWalletService(
	store WalletStore,
	keyManager crypto.KeyManager,
	clientManager web3client.ClientManager,
	cfg *config.Config,
) WalletService {
	return &walletService{
		store:         store,
		keyManager:    keyManager,
		clientManager: clientManager,
		cfg:           cfg,
	}
}

// CreateHDWallet implements WalletService.
func (s *walletService) CreateHDWallet(
	ctx context.Context,
	userID uint,
	password string,
	chainID uint,
) (*model.Wallet, string, error) {
	panic("unimplemented")
}

// GetBalance implements WalletService.
func (s *walletService) GetBalance(ctx context.Context, address string, chainID uint) (string, error) {
	// 1. 调用 ClientManager 获取余额 (ClientManager 封装了 RPC 连接和重试)
	balanceWei, err := s.clientManager.GetBalanceByAddress(ctx, chainID, address)
	if err != nil {
		return "", fmt.Errorf("failed to fetch balance from blockchain: %w", err)
	}

	// 2. 转换为人类可读格式 (ETH)
	balanceETH := conversion.WeiToEther(balanceWei)

	return balanceETH.String(), nil
}

// Transfer implements WalletService.
func (w *walletService) Transfer(
	ctx context.Context,
	fromAddress string,
	toAddress string,
	amount string,
	password string,
	chainID uint,
) (string, error) {
	panic("unimplemented")
}
