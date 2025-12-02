package service

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/model"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

var (
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")

	ErrStoreOperationFailed = errors.New("store operation failed")
	ErrPasswordHashFailed   = errors.New("password hashing failed")

	ErrUserNotFound = errors.New("user not found")
)

// UserStore 定义了用户数据访问的接口（在 service 层定义，由 store 层实现）
type UserStore interface {
	CreateUser(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

// UserService 定义了用户相关的业务逻辑接口
type UserService interface {
	Register(username string, password string) (*model.User, error)
	Login(username string, password string) (string, error)
	FindUserByUsername(username string) (*model.User, error)
}

type userService struct {
	userStore  UserStore
	jwtService JWTService
}

var _ UserService = (*userService)(nil)

// NewUserService 创建并返回新的 UserService 实例（依赖注入）
func NewUserService(userStore UserStore, jwtService JWTService) UserService {
	return &userService{
		userStore:  userStore,
		jwtService: jwtService,
	}
}

// Register 处理用户注册业务逻辑
func (s *userService) Register(username string, password string) (*model.User, error) {
	_, err := s.userStore.FindByUsername(username)

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, ErrUserAlreadyExists
		}

		// 数据库查询出现底层错误
		// 不记录日志，包装错误并向上层传递
		return nil, fmt.Errorf("%w: failed to check username existence: %w", ErrStoreOperationFailed, err)
	}

	// 2. 核心安全步骤：密码哈希 (Bcrypt)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// 记录底层错误（系统故障），包装错误并返回业务错误
		logger.Logger.Error("Error hashing password", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
	}

	// 3. 创建用户实体
	user := &model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// 4. 持久化存储
	if err := s.userStore.CreateUser(user); err != nil {
		// 记录底层错误（数据库写入失败），包装错误并返回业务错误
		logger.Logger.Error("Error creating user in store", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("%w: failed to create user: %w", ErrStoreOperationFailed, err)
	}

	// 5. 注册成功
	logger.Logger.Info("User registered successfully", zap.Uint("user_id", user.ID))
	return user, nil
}

// Login 处理用户登录业务逻辑
func (s *userService) Login(username string, password string) (string, error) {
	// 1. 查找用户
	user, err := s.userStore.FindByUsername(username)

	// 检查用户不存在或数据库返回的 RecordNotFound 错误
	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		return "", ErrInvalidCredentials // 隐藏用户不存在的细节
	}

	if err != nil {
		return "", fmt.Errorf("%w: failed to retrieve user: %w", ErrStoreOperationFailed, err)
	}

	// 2. 密码验证
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		logger.Logger.Error("Error comparing password hash", zap.Error(err))
		return "", ErrInvalidCredentials
	}

	// 3. 登录成功，生成 JWT Token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// FindUserByUsername 实现 UserService 接口
func (s *userService) FindUserByUsername(username string) (*model.User, error) {
	user, err := s.userStore.FindByUsername(username)

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("%w: failed to query user: %w", ErrStoreOperationFailed, err)
	}

	return user, nil
}
