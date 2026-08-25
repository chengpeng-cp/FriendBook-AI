package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件。
// 教学简化：Allow-Origin 放开为 *；生产环境应改为配置化的域名白名单。
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		// 暴露限流响应头，便于前端感知限流状态
		c.Header("Access-Control-Expose-Headers", "X-RateLimit-Limit, X-RateLimit-Remaining")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 预检请求缓存 24 小时

		// 预检请求直接放行
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
