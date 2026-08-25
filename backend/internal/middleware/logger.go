package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// sensitiveFields 请求体日志脱敏字段：值一律替换为 ***
var sensitiveFields = map[string]bool{
	"password":      true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"api_key":       true,
	"secret":        true,
}

// Logger 请求日志中间件：
//   请求：method / path / client_ip / body（脱敏后）
//   响应：status / 耗时 / 响应字节数
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// ---- 读取并还原请求体（gin 的 body 是一次性流，读后必须塞回去）----
		var bodyStr string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				bodyStr = sanitizeBody(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		c.Next()

		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("cost", time.Since(start)),
			zap.Int("size", c.Writer.Size()),
			zap.String("body", bodyStr),
		)
	}
}

// sanitizeBody 请求体脱敏：JSON 则按字段名打码，非 JSON 截断到 500 字符
func sanitizeBody(bodyBytes []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &m); err == nil {
		sanitizeMap(m)
		if cleaned, err := json.Marshal(m); err == nil {
			return string(cleaned)
		}
	}
	// 非 JSON body：截断防日志爆炸
	s := string(bodyBytes)
	if len(s) > 500 {
		return s[:500] + "...(truncated)"
	}
	return s
}

// sanitizeMap 递归打敏嵌套 JSON 中的敏感字段
func sanitizeMap(m map[string]interface{}) {
	for k, v := range m {
		if sensitiveFields[k] {
			m[k] = "***"
			continue
		}
		if nested, ok := v.(map[string]interface{}); ok {
			sanitizeMap(nested)
		}
	}
}
