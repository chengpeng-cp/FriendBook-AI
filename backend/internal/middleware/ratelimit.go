package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
)

// RateLimit 基于 Redis 的固定窗口计数限流。
// 维度：IP + 路径（每个接口独立限流互不影响）。
// 实现：Redis INCR 计数，首次访问设置 EXPIRE 窗口期；
// 超过阈值的请求直接拒绝（错误码 10005）。
//
// 已知边界（教学简化）：INCR 后进程崩溃会导致 key 未设置过期；
// 生产建议用 Lua 脚本原子化 INCR+EXPIRE 或滑动窗口算法。
func RateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.Request.URL.Path)

		// 原子自增
		count, err := rdb.Incr(c, key).Result()
		if err != nil {
			// Redis 故障时降级放行：限流是保护措施，不能反过来弄挂业务
			c.Next()
			return
		}
		// 窗口内首次访问，设置过期时间
		if count == 1 {
			rdb.Expire(c, key, window)
		}

		// 响应头携带配额信息（对排查限流问题很有用）
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(limit)-count, 10))

		// 超限拒绝
		if count > int64(limit) {
			response.Fail(c, errcode.RateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
