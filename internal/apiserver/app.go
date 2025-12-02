package apiserver

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/controller"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/router"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/service"
	"github.com/bwmspring/go-web3-wallet-backend/internal/apiserver/store"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/crypto"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/web3client"
)

// App 是应用程序的核心容器，管理所有依赖关系
type App struct {
	cfg *config.Config
	db  *gorm.DB

	// 驱动层/工具层 (Drivers)
	keyManager    crypto.KeyManager
	clientManager web3client.ClientManager

	// 存储层 (Stores)
	userStore   service.UserStore
	walletStore service.WalletStore

	// 业务层 (Services)
	jwtService    service.JWTService
	userService   service.UserService
	walletService service.WalletService

	// 控制器层 (Controllers)
	authController   *controller.AuthController
	userController   *controller.UserController
	walletController *controller.WalletController
}

// NewApp 创建并初始化应用容器
func NewApp(cfg *config.Config, db *gorm.DB) (*App, error) {
	app := &App{
		cfg: cfg,
		db:  db,
	}

	if err := app.initDrivers(); err != nil {
		return nil, fmt.Errorf("failed to init drivers: %w", err)
	}

	app.initStores()
	app.initServices()
	app.initControllers()

	return app, nil
}

// initDrivers 初始化所有底层驱动组件（如 Web3 客户端、密钥管理器）
func (a *App) initDrivers() error {
	a.keyManager = crypto.NewKeyManager()

	clientManager, err := web3client.NewClientManager(a.cfg.Chains)
	if err != nil {
		return fmt.Errorf("failed to create web3 client manager: %w", err)
	}
	a.clientManager = clientManager

	return nil
}

func (a *App) initStores() {
	a.userStore = store.NewUsers(a.db)
	a.walletStore = store.NewWallets(a.db)
}

func (a *App) initServices() {
	a.jwtService = service.NewJWTService(a.cfg)
	a.userService = service.NewUserService(a.userStore, a.jwtService)

	a.walletService = service.NewWalletService(a.walletStore, a.keyManager, a.clientManager, a.cfg)
}

func (a *App) initControllers() {
	a.authController = controller.NewAuthController(a.userService, a.jwtService)
	a.userController = controller.NewUserController(a.userService)
	a.walletController = controller.NewWalletController(a.walletService)
}

// InitRouter 初始化并返回配置好的 Gin Engine
func (a *App) InitRouter() *gin.Engine {
	if a.cfg == nil {
		panic("Application config (a.cfg) is nil!")
	}
	// 创建配置对象，将所有控制器和服务注入
	routerCfg := &router.RouterConfig{
		ServerCfg:        &a.cfg.Server,
		CORSConfig:       &a.cfg.CORS,
		LimitConfig:      &a.cfg.Limit,
		JWTService:       a.jwtService,
		AuthController:   a.authController,
		UserController:   a.userController,
		WalletController: a.walletController,
	}
	// 调用 router 包中的函数来构建 Engine
	return router.NewRouter(routerCfg)
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

// WalletController 返回钱包控制器
func (a *App) WalletController() *controller.WalletController {
	return a.walletController
}
