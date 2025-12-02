package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/bwmspring/go-web3-wallet-backend/config"
	"github.com/bwmspring/go-web3-wallet-backend/pkg/logger"
)

// IPLimiterStore 存储每个 IP 的限流器实例
type IPLimiterStore struct {
	limiters        *sync.Map // key: string (IP Address), value: *rate.Limiter
	rate            rate.Limit
	bucket          int
	cleanupInterval time.Duration // 清理过期限流器的时间间隔
}

// NewIPLimiterStore 创建一个新的 IPLimiterStore 实例
func NewIPLimiterStore(r float64, b int) *IPLimiterStore {
	store := &IPLimiterStore{
		limiters:        &sync.Map{},
		rate:            rate.Limit(r),
		bucket:          b,
		cleanupInterval: 5 * time.Minute, // 默认每 5 分钟清理一次
	}
	// 启动后台清理协程
	go store.cleanupExpiredLimiters()
	return store
}

// getLimiter 获取或创建给定 IP 的限流器
func (s *IPLimiterStore) getLimiter(ip string) *rate.Limiter {
	// 尝试加载已有的限流器
	limiter, ok := s.limiters.Load(ip)
	if ok {
		return limiter.(*rate.Limiter)
	}

	// 如果不存在，则创建一个新的限流器
	newLimiter := rate.NewLimiter(s.rate, s.bucket)

	// 尝试存储新的限流器。LoadOrStore 保证并发安全，如果其他协程同时创建了，则使用已存在的。
	actual, loaded := s.limiters.LoadOrStore(ip, newLimiter)

	// 如果是刚创建的，则返回新创建的；如果是已存在的，则返回已存在的。
	if loaded {
		return actual.(*rate.Limiter)
	}
	return newLimiter
}

// cleanupExpiredLimiters 定期清理不活跃的限流器，防止内存泄漏
func (s *IPLimiterStore) cleanupExpiredLimiters() {
	// TODO: 真正的生产环境清理逻辑应该基于 LRU 缓存或设置访问时间戳。
	// 这里的简单实现仅作为占位符，演示后台任务。
	logger.L().Info("Rate Limiter Cleanup Goroutine Started.")
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		// 简单起见，我们暂时不移除，但在高流量下这可能会导致内存不断增长。
		// 生产环境应该追踪上次访问时间，并移除长时间不活跃的 IP。
		// 示例：可以遍历 sync.Map，检查自定义包装结构中的 LastAccessTime。
		logger.L().Debugw("Performing rate limiter cleanup (Not removing yet)", "interval", s.cleanupInterval.String())
	}
}

// Limit 是频率限制 Gin 中间件
func Limit(cfg config.LimitConfig) gin.HandlerFunc {
	// 如果配置未启用，则返回一个空操作的中间件
	if !cfg.Enable || cfg.Rate <= 0 || cfg.Bucket <= 0 {
		logger.L().Warn("Rate Limiting is disabled or misconfigured. Skipping.")
		return func(c *gin.Context) { c.Next() }
	}

	// 初始化 IP 限流存储
	store := NewIPLimiterStore(cfg.Rate, cfg.Bucket)
	logger.L().Infow("Rate Limiting enabled.", "rate", cfg.Rate, "bucket", cfg.Bucket)

	return func(c *gin.Context) {
		// 1. 获取客户端 IP
		ip := c.ClientIP()

		// 2. 获取该 IP 对应的限流器
		limiter := store.getLimiter(ip)

		// 3. 尝试获取一个令牌。如果失败，表示请求超速。
		if !limiter.Allow() {
			// 限流日志
			logger.L().Warnw("Rate limit exceeded", "ip", ip, "path", c.Request.URL.Path)

			// 4. 返回 429 Too Many Requests
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": fmt.Sprintf("Too many requests from IP %s. Limit: %.2f requests/sec.", ip, cfg.Rate),
			})
			return
		}

		// 5. 允许请求继续
		c.Next()
	}
}
