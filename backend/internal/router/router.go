// Package router 路由注册中心：集中注册全部路由与中间件。
// 挂载顺序：Recovery → Logger → CORS → （/health、/swagger）→ RateLimit → JWT。
package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
	"github.com/redis/go-redis/v9"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"golang-learn/internal/ai"
	"golang-learn/internal/config"
	"golang-learn/internal/handler"
	"golang-learn/internal/initialize"
	"golang-learn/internal/mcpserver"
	"golang-learn/internal/middleware"
	"golang-learn/internal/pkg/response"
)

// Setup 组装路由树。
// 注意：/friends/cold、/friends/search 等具体路径必须注册在 /friends/:id 之前，
// 否则 Gin 会把 "cold" 当作 :id 解析（Gin 路由树的特性）。
func Setup(
	cfg *config.Config,
	container *initialize.Container,
	agent *ai.Agent,
	mcp *mcpserver.MCPServer,
	rdb *redis.Client,
	logger *zap.Logger,
) *gin.Engine {
	// release 模式关闭 gin 调试输出
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ===== 全局中间件（顺序即执行顺序）=====
	r.Use(middleware.Recovery(logger)) // panic 兜底
	r.Use(middleware.Logger(logger))   // 请求日志（脱敏）
	r.Use(middleware.Cors())           // 跨域

	// ===== 健康检查（Docker / K8s 探活用，不限流）=====
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// ===== Swagger 文档 =====
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ===== MCP Server SSE 端点（供 Claude Desktop 等外部客户端接入）=====
	if mcp != nil {
		sseServer := server.NewSSEServer(mcp.GetServer(),
			server.WithBaseURL("http://localhost:"+strconv.Itoa(cfg.Server.Port)))
		// v0.26.0 API：SSEHandler 处理 SSE 长连接，MessageHandler 处理 JSON-RPC 消息
		// 两者返回 http.Handler，需用 gin.WrapH 适配为 gin.HandlerFunc
		r.GET("/mcp/sse", gin.WrapH(sseServer.SSEHandler()))
		r.POST("/mcp/message", gin.WrapH(sseServer.MessageHandler()))
	}

	// ===== API v1 =====
	api := r.Group("/api/v1")
	// 限流：每 IP 每路径每分钟 N 次（可配置开关）
	if cfg.RateLimit.Enabled {
		api.Use(middleware.RateLimit(rdb, cfg.RateLimit.LimitPerMinute, time.Minute))
	}
	{
		// ---- 用户（无需鉴权）----
		userHandler := handler.NewUserHandler(container.UserService)
		api.POST("/users/register", userHandler.Register)
		api.POST("/users/login", userHandler.Login)

		// ---- 需要鉴权的接口 ----
		auth := api.Group("")
		auth.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 用户信息
			auth.GET("/users/profile", userHandler.GetProfile)
			auth.PUT("/users/profile", userHandler.UpdateProfile)
			auth.PUT("/users/password", userHandler.ChangePassword)

			// 好友管理（具体路径放前，:id 路由放后）
			friendHandler := handler.NewFriendHandler(container.FriendService)
			auth.GET("/friends/search", friendHandler.Search)
			auth.GET("/friends/cold", friendHandler.ListCold)
			auth.GET("/friends/export", friendHandler.ExportCSV)
			auth.POST("/friends/import", friendHandler.ImportCSV)
			auth.POST("/friends", friendHandler.Create)
			auth.DELETE("/friends/:id", friendHandler.Delete)
			auth.PUT("/friends/:id", friendHandler.Update)
			auth.GET("/friends/:id", friendHandler.GetByID)
			auth.GET("/friends", friendHandler.List)
			auth.PUT("/friends/:id/favorability", friendHandler.AdjustFavorability)
			auth.POST("/friends/:id/accounts", friendHandler.AddAccount)
			auth.GET("/friends/:id/accounts", friendHandler.ListAccounts)

			// 分组
			groupHandler := handler.NewGroupHandler(container.GroupService)
			auth.POST("/groups", groupHandler.Create)
			auth.DELETE("/groups/:id", groupHandler.Delete)
			auth.PUT("/groups/:id", groupHandler.Update)
			auth.GET("/groups", groupHandler.List)

			// 标签
			tagHandler := handler.NewTagHandler(container.TagService)
			auth.POST("/tags", tagHandler.Create)
			auth.DELETE("/tags/:id", tagHandler.Delete)
			auth.GET("/tags", tagHandler.List)

			// 互动记录
			interactHandler := handler.NewInteractionHandler(container.InteractionService)
			auth.POST("/interactions", interactHandler.Create)
			auth.DELETE("/interactions/:id", interactHandler.Delete)
			auth.PUT("/interactions/:id", interactHandler.Update)
			auth.GET("/interactions", interactHandler.List)

			// 提醒
			reminderHandler := handler.NewReminderHandler(container.ReminderService)
			auth.GET("/reminders/today", reminderHandler.GetToday)
			auth.GET("/reminders/upcoming", reminderHandler.GetUpcoming)

			// 数据统计
			statsHandler := handler.NewStatsHandler(container.StatsService)
			auth.GET("/stats/overview", statsHandler.Overview)
			auth.GET("/stats/platform-distribution", statsHandler.PlatformDistribution)
			auth.GET("/stats/temperature-distribution", statsHandler.TemperatureDistribution)
			auth.GET("/stats/interaction-trend", statsHandler.InteractionTrend)

			// AI 助手（SSE 流式 + 对话历史持久化）
			aiHandler := handler.NewAIHandler(agent, container.ChatRepo, logger)
			auth.POST("/ai/chat", aiHandler.Chat)
			auth.GET("/ai/history", aiHandler.History)
			auth.DELETE("/ai/history", aiHandler.ClearHistory)
		}
	}

	return r
}
