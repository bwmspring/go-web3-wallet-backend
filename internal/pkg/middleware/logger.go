package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// Logger 是一个 Gin 中间件，用于记录请求的结构化日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取请求的开始时间 (由 Context 中间件注入)
		var start time.Time
		if startTime, ok := c.Get(string(RequestStartTimeKey)); ok {
			start = startTime.(time.Time)
		} else {
			start = time.Now()
		}

		// 2. 执行后续的请求处理逻辑
		c.Next()

		// 3. 请求处理完毕后，计算延迟
		end := time.Now()
		latency := end.Sub(start)

		// 4. 从 Gin.Context 获取 Request ID (由 Context 中间件注入)
		requestID := ""
		if rid, ok := c.Get(string(RequestIDKey)); ok {
			requestID = rid.(string)
		}

		// 5. 结构化记录请求日志
		logFunc := logger.L().Infow

		statusCode := c.Writer.Status()

		if statusCode >= 500 {
			// 5xx 服务器错误：严重问题，使用 Errorw
			logFunc = logger.L().Errorw
		} else if statusCode == 404 || statusCode == 400 {
			// 404 Not Found 和 400 Bad Request：常见的客户端错误，保持 Infow 即可。
			// 如果在生产环境想彻底忽略 404，可以设置为 Debugw
			logFunc = logger.L().Infow
		} else if statusCode >= 401 && statusCode < 500 {
			// 401 Unauthorized, 403 Forbidden, 429 Too Many Requests 等：
			// 这些是需要注意的安全或限流问题，使用 Warnw
			logFunc = logger.L().Warnw
		}

		logFunc(
			"Request completed",
			"requestID", requestID,
			"statusCode", statusCode,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
			"userAgent", c.Request.UserAgent(),
			"latency", latency.String(),
			"errors", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}
