package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/database"
	"github.com/bwmspring/go-web3-wallet-backend/model"
)

// UserRepository 定义了用户数据访问的方法接口
type UserRepository interface {
	CreateUser(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建并返回一个新的 UserRepository 实例
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.GetDB(),
	}
}

// CreateUser 在数据库中创建新用户记录
func (r *userRepository) CreateUser(user *model.User) error {
	result := r.db.Create(user)
	return result.Error
}

// FindByUsername 通过用户名查找用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
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
func (r *userRepository) FindByID(id uint) (*model.User, error) {
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
