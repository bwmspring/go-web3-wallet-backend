package cmd

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
	"github.com/bwmspring/go-web3-wallet-backend/internal/app"
	"github.com/bwmspring/go-web3-wallet-backend/internal/app/store"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

var cfgFile string

var Cfg *config.Config

// serveCmd 代表 'serve' 命令，用于启动 Web 服务器
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Go Web3 Wallet API server",
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
	// 注册 serve 命令
	rootCmd.AddCommand(serveCmd)

	// 为 serve 命令添加一个 flag 用于指定配置文件路径
	serveCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
}

func runServe() {
	// 加载配置
	var err error
	Cfg, err = config.LoadConfigFromFile(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
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
	application := app.NewApp(Cfg, database.DB())

	// 获取配置好的路由
	r := application.Router()

	// 配置 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器的 Goroutine
	go func() {
		logger.Logger.Info("Server is starting", zap.Int("port", Cfg.Server.Port))

		// srv.ListenAndServe() 是阻塞的，如果调用成功，会返回 http.ErrServerClosed
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Server listen error", zap.Error(err))
		}
	}()

	// 监听操作系统信号
	// 创建一个 channel 用于接收系统信号
	quit := make(chan os.Signal, 1)

	// 监听 SIGINT (Ctrl+C) 和 SIGTERM (一般用于容器或进程管理器关闭)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到接收到信号
	sig := <-quit
	logger.Logger.Info("Received signal. Starting graceful shutdown", zap.String("signal", sig.String()))

	// 设置关机超时
	// 给予 5 秒钟来完成正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行优雅关机
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown (timeout or error)", zap.Error(err))
	}

	logger.Logger.Info("Server exiting gracefully")
}
