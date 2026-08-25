// Package middleware HTTP 中间件：鉴权、跨域、限流、日志、恢复。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/pkg/utils"
)

// JWTAuth JWT 鉴权中间件：
//  1. 从 Authorization Header 提取 "Bearer <token>"
//  2. 验签 + 过期校验
//  3. 把 user_id / username 注入 gin.Context，后续 Handler 通过 c.GetUint("user_id") 获取
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, errcode.Unauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.FailWithMsg(c, errcode.Unauthorized, "Token 格式错误，应为 Bearer <token>")
			c.Abort()
			return
		}

		// 2. 解析校验
		claims, err := utils.ParseToken(secret, parts[1])
		if err != nil {
			if err == utils.ErrTokenExpired {
				response.Fail(c, errcode.TokenExpired)
			} else {
				response.Fail(c, errcode.Unauthorized)
			}
			c.Abort()
			return
		}

		// 3. 注入用户身份，供后续 Handler / Service 使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
