// FriendBook AI —— 智能好友管理系统后端
//
// 技术栈：Gin + GORM + MySQL + Redis + JWT + Eino (AI Agent) + MCP
// 启动流程：配置 → 日志 → MySQL → Redis → 依赖容器 → AI Agent → MCP Server → 定时任务 → HTTP
//
//	@title			FriendBook AI API
//	@version		1.0.0
//	@description	智能好友管理系统后端接口文档
//	@BasePath		/
package main

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	_ "golang-learn/docs" // Swagger 文档（make swag 重新生成）

	"golang-learn/internal/ai"
	"golang-learn/internal/config"
	"golang-learn/internal/initialize"
	"golang-learn/internal/mcpserver"
	"golang-learn/internal/pkg/logger"
	"golang-learn/internal/router"
)

func main() {
	ctx := context.Background()

	// 1. 加载配置（配置文件路径可按环境覆盖）
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 2. 初始化日志（后续所有组件用它打日志）
	appLogger := logger.New(&cfg.Log)
	defer appLogger.Sync() // 退出前刷缓冲

	// 3. 初始化 MySQL（含连接重试 + AutoMigrate）
	db, err := initialize.InitDB(&cfg.MySQL)
	if err != nil {
		appLogger.Fatal("初始化 MySQL 失败", zap.Error(err))
	}

	// 4. 初始化 Redis
	rdb, err := initialize.InitRedis(&cfg.Redis)
	if err != nil {
		appLogger.Fatal("初始化 Redis 失败", zap.Error(err))
	}

	// 5. 依赖注入容器（Repository → Service）
	container := initialize.InitContainer(db, rdb, cfg, appLogger)

	// 6. Eino Tools（复用 Service 层，与 HTTP Handler 共享业务逻辑）
	tools := ai.NewTools(
		container.FriendService,
		container.InteractionService,
		container.StatsService,
		container.ReminderService,
		appLogger,
	)

	// 7. AI Agent（ReAct：ChatModel + Tools）
	agent, err := ai.NewAgent(ctx, ai.AgentConfig{
		OpenAIKey:     cfg.OpenAI.APIKey,
		OpenAIModel:   cfg.OpenAI.Model,
		OpenAIBaseURL: cfg.OpenAI.BaseURL,
	}, tools, appLogger)
	if err != nil {
		appLogger.Fatal("初始化 AI Agent 失败", zap.Error(err))
	}

	// 8. MCP Server（对外暴露，供 Claude Desktop 等外部 AI 客户端调用）
	mcp := mcpserver.New(
		container.FriendService,
		container.StatsService,
		appLogger,
	)

	// 9. 定时任务（每天凌晨 1 点：刷新关系温度 + 生成生日提醒）
	initialize.StartCronJobs(
		container.UserRepo,
		container.FriendService,
		container.ReminderService,
		appLogger,
	)

	// 10. 注册路由并启动 HTTP 服务
	r := router.Setup(cfg, container, agent, mcp, rdb, appLogger)
	addr := ":" + strconv.Itoa(cfg.Server.Port)
	appLogger.Info("服务启动", zap.String("addr", addr), zap.String("mode", cfg.Server.Mode))
	if err := r.Run(addr); err != nil {
		appLogger.Fatal("服务启动失败", zap.Error(err))
	}
}
