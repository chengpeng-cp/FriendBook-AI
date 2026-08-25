// Package errcode 统一错误码定义。
// 约定：0 成功；1xxxx 通用错误；2xxxx 用户相关；3xxxx 好友相关；4xxxx AI 相关。
package errcode

const (
	Success = 0 // 成功

	// 通用错误 1xxxx
	InvalidParams = 10001 // 参数错误
	Unauthorized  = 10002 // 未授权（未登录 / Token 无效）
	Forbidden     = 10003 // 禁止访问
	NotFound      = 10004 // 资源不存在
	RateLimited   = 10005 // 请求过于频繁（触发限流）
	InternalError = 10006 // 服务器内部错误

	// 用户相关 2xxxx
	UserAlreadyExist  = 20001 // 用户已存在
	UserNotFound      = 20002 // 用户不存在
	PasswordIncorrect = 20003 // 密码错误
	TokenExpired      = 20004 // Token 过期

	// 好友相关 3xxxx
	FriendNotFound = 30001 // 好友不存在
	GroupNotFound  = 30002 // 分组不存在

	// AI 相关 4xxxx
	AIChatError = 40001 // AI 对话失败
)
