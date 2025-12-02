package store

import (
	"errors"

	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/model"
)

type users struct {
	db *gorm.DB
}

var _ service.UserStore = (*users)(nil)

// NewUsers 创建并返回一个新的 UserStore 实例
func NewUsers(db *gorm.DB) service.UserStore {
	return &users{
		db: db,
	}
}

// CreateUser 在数据库中创建新用户记录
func (r *users) CreateUser(user *model.User) error {
	result := r.db.Create(user)
	return result.Error
}

// FindByUsername 通过用户名查找用户
func (r *users) FindByUsername(username string) (*model.User, error) {
	var user model.User
	// GORM 自动添加 deleted_at IS NULL 条件
	result := r.db.Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID 通过 ID 查找用户
func (r *users) FindByID(id uint) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}
