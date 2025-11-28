package model

import (
	"time"

	"gorm.io/gorm"
)

// Wallet 代表一个链上钱包实体。严格对应 'wallets' 数据库表。
type Wallet struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// UserID 是外键，关联到 User 表
	UserID uint `gorm:"not null" json:"user_id"`

	// 公钥地址 (例如: 0x...)，必须唯一
	Address string `gorm:"unique;not null;type:varchar(42)" json:"address"`

	// EncryptedKey 存储 Keystore JSON
	EncryptedKey string `gorm:"type:text;not null" json:"-"`

	// DerivationPath 派生路径 (如：m/44'/60'/0'/0/0)
	DerivationPath string `gorm:"type:varchar(50)" json:"derivation_path"`

	Name      string    `json:"name"` // 钱包名称 (用户可自定义)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// **软删除字段**
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MnemonicSeed 用于存储用户的助记词信息（也需高度加密）
type MnemonicSeed struct {
	ID            uint   `gorm:"primarykey"`
	UserID        uint   `gorm:"index;not null"`
	EncryptedSeed string `gorm:"type:text;not null;comment:加密后的BIP39助记词"` // 使用不同的密钥或方法加密
	CreatedAt     time.Time
}
