package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
)

// Recovery 全局 panic 恢复中间件：
// 捕获 Handler 链路中的 panic，记录堆栈并返回统一 500 业务响应，
// 避免单个请求的 panic 打挂整个进程。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.Stack("stack"),
				)
				response.Fail(c, errcode.InternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
