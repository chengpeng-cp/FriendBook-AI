// Package response 统一 HTTP 响应封装。
// 所有接口统一返回 { code, message, data } 结构，HTTP 状态码恒为 200，
// 业务成败通过 code 字段区分（前端只需关心 code 是否为 0）。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务码：0 成功，非 0 失败
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据，失败时为 nil
}

// PageData 分页响应数据
type PageData struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总条数
	Page     int         `json:"page"`      // 当前页码（从 1 开始）
	PageSize int         `json:"page_size"` // 每页条数
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// OKWithMsg 成功响应（自定义提示语）
func OKWithMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: msg,
		Data:    data,
	})
}

// Fail 失败响应（指定错误码 + 默认提示）
func Fail(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msgOf(code),
		Data:    nil,
	})
}

// FailWithMsg 失败响应（指定错误码 + 自定义提示）
func FailWithMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}

// msgOf 错误码默认提示文案
func msgOf(code int) string {
	switch code {
	case 10001:
		return "参数错误"
	case 10002:
		return "未授权"
	case 10003:
		return "禁止访问"
	case 10004:
		return "资源不存在"
	case 10005:
		return "请求过于频繁"
	case 10006:
		return "服务器内部错误"
	case 20001:
		return "用户已存在"
	case 20002:
		return "用户不存在"
	case 20003:
		return "密码错误"
	case 20004:
		return "Token 已过期"
	case 30001:
		return "好友不存在"
	case 30002:
		return "分组不存在"
	case 40001:
		return "AI 对话失败"
	default:
		return "未知错误"
	}
}
