package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 定义Context Key，用于存取值
type contextKey string

const (
	// RequestIDKey 是存储请求ID的Context键
	RequestIDKey contextKey = "X-Request-ID"
	// UsernameKey 是存储用户名的Context键（未来认证后使用）
	UsernameKey contextKey = "username"
	// RequestStartTimeKey 是存储请求开始时间的Context键
	RequestStartTimeKey contextKey = "request-start-time"
)

// Context 用于处理请求上下文，注入RequestID、起始时间等信息。
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 生成 Request ID
		requestID := c.Request.Header.Get(string(RequestIDKey))
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 2. 将 Request ID 和 Start Time 存入 Gin.Context
		c.Set(string(RequestIDKey), requestID)
		c.Set(string(RequestStartTimeKey), time.Now())

		// 3. 将 Request ID 注入到 Go 的 context.Context 中
		// 这使得在业务逻辑深处也可以通过 c.Request.Context() 获取 Request ID
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 4. 设置 Response Header
		c.Writer.Header().Set(string(RequestIDKey), requestID)

		c.Next()
	}
}

// GetRequestID 从 Go context.Context 中获取 Request ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
