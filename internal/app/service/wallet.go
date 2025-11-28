package service

import (
	"context"

	"github.com/bwmspring/go-web3-wallet-backend/model"
)

// WalletRepository 定义了钱包数据存储的接口 (DIP: 由 service 层定义)
type WalletRepository interface {
	// CreateWallet 在数据库中创建一个新的钱包记录
	CreateWallet(ctx context.Context, wallet *model.Wallet, mnemonic *model.MnemonicSeed) error

	// GetWalletByUserID 根据用户ID和链ID获取钱包
	GetWalletByUserID(ctx context.Context, userID uint, chainID uint) (*model.Wallet, error)

	// FindEncryptedKeyByAddress 根据地址查找加密后的私钥（Keystore）
	FindEncryptedKeyByAddress(ctx context.Context, address string) (string, error)

	// ... 其他数据操作接口，如更新余额，查询交易历史等
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

	// ... 其他业务接口，如导入钱包、导出Keystore、查询Gas费等
}
