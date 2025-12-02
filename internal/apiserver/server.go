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

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	store "github.com/bwmspring/go-web3-wallet-backend/pkg/db"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// cfgFile 用于存储命令行传入的配置文件路径
var cfgFile string

// NewAPIServerCommand 创建 Apiserver 启动命令的工厂函数
func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "Start the Go Web3 Wallet API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIServer()
		},
	}

	// 将配置文件 flag 绑定到 apiserver 子命令上
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	return cmd
}

// runAPIServer 包含实际的服务器初始化和运行逻辑
func runAPIServer() error {
	// 加载配置
	Cfg, err := config.LoadConfigFromFile(cfgFile)
	if err != nil {
		// 返回错误，由 Cobra 处理退出码
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 日志初始化
	logger.InitLogger(Cfg.Server.Environment)
	defer logger.Logger.Sync()

	// 初始化数据库连接
	database, err := store.NewDatabase(Cfg)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// 初始化应用容器（包含所有依赖）
	application, err := NewApp(Cfg, database.DB())
	if err != nil {
		logger.Logger.Fatal("Failed to initialize APIServer", zap.Error(err))
	}

	// 获取配置好的路由
	r := application.InitRouter()

	// 配置 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器的 Goroutine
	go func() {
		logger.Logger.Info("APIServer is starting", zap.Int("port", Cfg.Server.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("APIServer listen error", zap.Error(err))
		}
	}()

	// 监听操作系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Logger.Info("Received signal. Starting graceful shutdown", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("APIServer forced to shutdown (timeout or error)", zap.Error(err))
	}

	logger.Logger.Info("APIServer exiting gracefully")
	return nil
}
