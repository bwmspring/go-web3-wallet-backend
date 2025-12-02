package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	store "github.com/bwmspring/go-web3-wallet-backend/pkg/db"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// Run 启动 API 服务器
// 该函数包含服务器的完整生命周期：配置加载、依赖初始化、HTTP 服务启动和优雅关闭
func Run(configPath string) error {
	// 1. 加载配置
	cfg, err := config.LoadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 2. 初始化日志
	logger.InitLogger(cfg.Server.Environment)
	defer logger.Logger.Sync()

	// 3. 初始化数据库连接
	database, err := store.NewDatabase(cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// 4. 初始化应用容器（包含所有依赖注入）
	application, err := NewApp(cfg, database.DB())
	if err != nil {
		logger.Logger.Fatal("Failed to initialize APIServer", zap.Error(err))
	}

	// 5. 获取配置好的路由
	router := application.InitRouter()

	// 6. 配置 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 7. 在独立的 goroutine 中启动服务器
	go func() {
		logger.Logger.Info(
			"APIServer is starting",
			zap.Int("port", cfg.Server.Port),
			zap.String("environment", cfg.Server.Environment),
		)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("APIServer listen error", zap.Error(err))
		}
	}()

	// 8. 监听操作系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Logger.Info("Received signal. Starting graceful shutdown", zap.String("signal", sig.String()))

	// 9. 执行优雅关闭（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("APIServer forced to shutdown (timeout or error)", zap.Error(err))
	}

	logger.Logger.Info("APIServer exiting gracefully")
	return nil
}
