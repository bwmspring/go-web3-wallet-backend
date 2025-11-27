package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是全局导出的 Zap Logger 实例
var Logger *zap.Logger

// InitLogger 根据运行环境初始化 Zap Logger
func InitLogger(env string) {
	var config zap.Config
	var err error

	if env == "production" {
		config = zap.NewProductionConfig()
		config.Encoding = "json"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.LevelKey = "level"
		config.DisableStacktrace = true
	} else {
		config = zap.NewDevelopmentConfig()
		config.Encoding = "console"
	}

	zLogger, err := config.Build()
	if err != nil {
		log.Fatalf("Fatal: 无法初始化 Zap Logger: %v", err)
	}

	// 导出原生 Logger 实例
	Logger = zLogger
	// 使用原生 Logger 记录初始化信息
	Logger.Info("Zap Logger initialized successfully", zap.String("environment", env))

	// 确保标准库 log 的输出也被重定向到 Zap
	zap.RedirectStdLog(Logger)
}
