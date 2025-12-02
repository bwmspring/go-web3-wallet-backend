package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是全局导出的 Zap Logger 实例
var Logger *zap.Logger

// SLogger 是 SugaredLogger 实例，用于更方便的结构化日志
var SLogger *zap.SugaredLogger

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
	// 创建 SugaredLogger 实例
	SLogger = Logger.Sugar()

	SLogger.Infow("Zap Logger initialized successfully", "environment", env)

	// 确保标准库 log 的输出也被重定向到 Zap
	zap.RedirectStdLog(Logger)
}

// L 返回全局的 SugaredLogger 实例，用于记录结构化日志
func L() *zap.SugaredLogger {
	// 防止在未初始化时调用
	if SLogger == nil {
		// 返回一个空的 Logger 以避免运行时崩溃，但会丢失日志
		return zap.NewNop().Sugar()
	}
	return SLogger
}
