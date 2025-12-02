package model

import (
	"time"

	"gorm.io/gorm"
)

// Wallet 代表一个链上钱包实体。严格对应 'wallets' 数据库表。
type Wallet struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 基础信息
	UserID  uint   `gorm:"not null;index"`
	ChainID uint   `gorm:"not null"`
	Name    string `gorm:"size:100;not null"`
	Address string `gorm:"size:42;uniqueIndex;not null"` // 钱包地址

	// 关联到助记词表
	MnemonicID   uint         `gorm:"not null;index"`
	MnemonicSeed MnemonicSeed `gorm:"foreignKey:MnemonicID"` // GORM 关系定义

	// 安全信息
	EncryptedKey   string `gorm:"type:text;not null"` // Keystore JSON
	DerivationPath string `gorm:"size:255;not null"`  // BIP-44 路径

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"          gorm:"index"`
}

// MnemonicSeed 用于存储用户的助记词信息（也需高度加密）
type MnemonicSeed struct {
	ID            uint   `gorm:"primarykey"`
	UserID        uint   `gorm:"index;not null"`
	EncryptedSeed string `gorm:"type:text;not null;comment:加密后的BIP39助记词"` // 使用不同的密钥或方法加密
	CreatedAt     time.Time
}
