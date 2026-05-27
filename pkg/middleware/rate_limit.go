package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count    int
	resetAt  time.Time
	mu       sync.Mutex
}

var (
	ipBuckets   = map[string]*bucket{}
	bucketsMu   sync.RWMutex
)

// RateLimit 简单的滑动窗口限流：每个 IP 在 window 时间内最多 max 次请求
// 生产环境建议换成 Redis + 令牌桶
func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	// 后台定期清理过期桶，防止内存泄漏
	go func() {
		for range time.Tick(5 * time.Minute) {
			bucketsMu.Lock()
			now := time.Now()
			for ip, b := range ipBuckets {
				if now.After(b.resetAt) {
					delete(ipBuckets, ip)
				}
			}
			bucketsMu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		bucketsMu.RLock()
		b, ok := ipBuckets[ip]
		bucketsMu.RUnlock()

		if !ok {
			b = &bucket{}
			bucketsMu.Lock()
			ipBuckets[ip] = b
			bucketsMu.Unlock()
		}

		b.mu.Lock()
		now := time.Now()
		if now.After(b.resetAt) {
			b.count = 0
			b.resetAt = now.Add(window)
		}
		b.count++
		count := b.count
		b.mu.Unlock()

		if count > max {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}

		// 响应头告知客户端剩余配额
		remaining := max - count
		if remaining < 0 { remaining = 0 }
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	}
}
