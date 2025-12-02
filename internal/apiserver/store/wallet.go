package store

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/model"
)

// wallets 实现了 service.WalletStore 接口
type wallets struct {
	db *gorm.DB
}

var _ service.WalletStore = (*wallets)(nil)

// NewWallets 实例化 WalletStore，并返回 service.WalletStore 接口类型
func NewWallets(db *gorm.DB) service.WalletStore {
	return &wallets{db: db}
}

// CreateWallet 在数据库中创建一个新的钱包记录和助记词记录（使用事务确保一致性）
func (r *wallets) CreateWallet(ctx context.Context, wallet *model.Wallet, mnemonic *model.MnemonicSeed) error {
	// 启动数据库事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 1. 创建 MnemonicSeed 记录
	if err := tx.Create(mnemonic).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create mnemonic seed record: %w", err)
	}

	// 2. 关联 Wallet 和 MnemonicSeed (设置外键)
	// 在 service 层必须确保 wallet.MnemonicID = mnemonic.ID 已经设置，
	// 确保外键关系正确建立。
	wallet.MnemonicID = mnemonic.ID

	// 3. 创建 Wallet 记录
	if err := tx.Create(wallet).Error; err != nil {
		tx.Rollback()

		// 检查是否是地址重复错误 (假设 Address 字段有 unique 约束)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			// 虽然理论上 KeyManager 应该生成唯一的地址，但数据库层必须处理这个技术错误
			return fmt.Errorf("wallet address already exists: %w", err)
		}

		return fmt.Errorf("failed to create wallet record: %w", err)
	}

	// 4. 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetWalletByUserID 根据用户ID和链ID获取钱包
// 注意：此方法用于确认用户是否存在某个链的钱包，但不会获取 Keystore 信息。
func (r *wallets) GetWalletByUserID(ctx context.Context, userID uint, chainID uint) (*model.Wallet, error) {
	wallet := &model.Wallet{}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND chain_id = ?", userID, chainID).
		// 安全注意：不使用 Select("*")，避免无意中查询过多字段。
		// 但为了返回完整的 model.Wallet 结构体，这里依赖 GORM 默认查询，
		// 字段过滤应该在 Service 层决定如何使用。
		First(wallet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 记录未找到，返回 nil 钱包，nil 错误
		}
		return nil, fmt.Errorf("failed to query wallet by user ID and chain ID: %w", err)
	}

	return wallet, nil
}

// FindEncryptedKeyByAddress 根据地址查找加密后的私钥（Keystore）
// 作用：专用于 Transfer 业务，用于获取 Keystore 进行解密。
func (r *wallets) FindEncryptedKeyByAddress(ctx context.Context, address string) (string, error) {
	wallet := &model.Wallet{}

	// 性能与安全：只查询 EncryptedKey 字段
	err := r.db.WithContext(ctx).
		Select("encrypted_key").
		Where("address = ?", address).
		First(wallet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // 地址未找到
		}
		return "", fmt.Errorf("failed to query encrypted key by address: %w", err)
	}

	return wallet.EncryptedKey, nil
}
