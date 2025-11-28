package app

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/controller"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/repository"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/service"
)

// App 是应用程序的核心容器，管理所有依赖关系
type App struct {
	cfg *config.Config
	db  *gorm.DB

	// Repositories
	userRepo service.UserRepository

	// Services
	jwtService  service.JWTService
	userService service.UserService

	// Controllers
	authController *controller.AuthController
	userController *controller.UserController
}

// NewApp 创建并初始化应用容器
func NewApp(cfg *config.Config, db *gorm.DB) *App {
	app := &App{
		cfg: cfg,
		db:  db,
	}

	// 按依赖顺序初始化各层
	app.initRepositories()
	app.initServices()
	app.initControllers()

	return app
}

// initRepositories 初始化所有 Repository 层
func (a *App) initRepositories() {
	a.userRepo = repository.NewUserRepository(a.db)
}

// initServices 初始化所有 Service 层
func (a *App) initServices() {
	a.jwtService = service.NewJWTService(a.cfg)
	a.userService = service.NewUserService(a.userRepo, a.jwtService)
}

// initControllers 初始化所有 Controller 层
func (a *App) initControllers() {
	a.authController = controller.NewAuthController(a.userService, a.jwtService)
	a.userController = controller.NewUserController(a.userService)
}

// Router 返回配置好的 Gin Engine
func (a *App) Router() *gin.Engine {
	return a.InitRouter()
}

// AuthController 返回认证控制器
func (a *App) AuthController() *controller.AuthController {
	return a.authController
}

// UserController 返回用户控制器
func (a *App) UserController() *controller.UserController {
	return a.userController
}

// JWTService 返回 JWT 服务
func (a *App) JWTService() service.JWTService {
	return a.jwtService
}
