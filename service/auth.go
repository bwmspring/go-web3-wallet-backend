package service

import (
	"errors"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/model"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
	"github.com/bwmspring/go-web3-wallet-backend/repository"
)

// AuthService 定义了认证相关的业务逻辑方法。
type AuthService interface {
	Register(username string, password string) (*model.User, error)
	Login(username string, password string) (string, error)
	FindUserByUsername(username string) (*model.User, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtService JWTService
}

// NewAuthService 创建并返回新的 AuthService 实例。
func NewAuthService(cfg *config.Config) AuthService {
	return &authService{
		userRepo:   repository.NewUserRepository(),
		jwtService: NewJWTService(cfg),
	}
}

// Register 处理用户注册业务逻辑。
func (s *authService) Register(username string, password string) (*model.User, error) {
	// 1. 业务校验：检查用户名是否已存在
	_, err := s.userRepo.FindByUsername(username)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, errors.New("用户名已存在")
		}
		logger.Logger.Error("Error checking username existence", zap.Error(err))
		return nil, errors.New("服务器错误，请稍后重试")
	}

	// 2. 核心安全步骤：密码哈希 (Bcrypt)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Logger.Error("Error hashing password", zap.String("username", username), zap.Error(err))
		return nil, errors.New("内部错误：密码处理失败")
	}

	// 3. 创建用户实体
	user := &model.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// 4. 持久化存储
	if err := s.userRepo.CreateUser(user); err != nil {
		logger.Logger.Error("Error creating user in repository", zap.String("username", username), zap.Error(err))
		return nil, errors.New("数据库写入失败")
	}

	logger.Logger.Info("User registered successfully", zap.Uint("user_id", user.ID))
	return user, nil
}

// Login 处理用户登录业务逻辑。
func (s *authService) Login(username string, password string) (string, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		return "", errors.New("用户名或密码无效")
	}
	if err != nil {
		logger.Logger.Error("Error retrieving user during login", zap.Error(err))
		return "", errors.New("服务器错误")
	}

	// 2. 密码验证
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("用户名或密码无效")
	}

	// 3. 登录成功，生成 JWT Token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		logger.Logger.Error("Failed to generate JWT token", zap.Uint("user_id", user.ID), zap.Error(err))
		return "", errors.New("无法生成认证令牌，请重试")
	}

	return token, nil
}

// FindUserByUsername 实现 AuthService 接口
func (s *authService) FindUserByUsername(username string) (*model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户不存在")
	}
	return user, err
}
