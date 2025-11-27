package model

import (
	"time"

	"gorm.io/gorm"
)

// User 代表应用用户实体。严格对应 'users' 数据库表。
type User struct {
	// ID 是主键，GORM 会自动设置为 SERIAL PRIMARY KEY
	ID uint `gorm:"primaryKey" json:"id"`

	// 用户名用于登录，要求唯一且非空
	Username string `gorm:"unique;not null;type:varchar(50)" json:"username"`

	// PasswordHash 存储 Bcrypt 哈希后的密码
	PasswordHash string `gorm:"not null" json:"-"`

	// GORM 自动维护时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// **软删除字段**：GORM 约定，添加此字段将启用自动软删除功能。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
