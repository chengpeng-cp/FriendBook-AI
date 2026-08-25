// Package mcpserver MCP Server 模块（Phase 3）：
// 将项目的好友管理能力通过 MCP 协议对外暴露，
// 供外部 AI 客户端（Claude Desktop / Cursor 等）直接调用。
//
// 与 Eino Tools 的区别（面试必考）：
//   Eino Tools —— 项目内置 ReAct Agent 的"内线电话"（进程内函数调用）
//   MCP        —— 对部暴露的"门牌号"（SSE 网络协议，跨进程跨机器）
// 两者复用同一套 Service 层，业务逻辑只有一份。
package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"golang-learn/internal/repository"
	"golang-learn/internal/service"
)

// MCPUserID MCP 会话默认操作的用户 ID。
// 教学简化：MCP 不走 JWT，固定操作用户 1 的数据；
// 生产环境应实现 MCP 会话鉴权（API Key → user_id 映射）。
const MCPUserID uint = 1

// MCPServer 持有 MCP Server 实例和复用的 Service
type MCPServer struct {
	server    *server.MCPServer
	friendSvc *service.FriendService
	statsSvc  *service.StatsService
	logger    *zap.Logger
}

// New 创建 MCP Server 并注册所有 Tools
func New(
	friendSvc *service.FriendService,
	statsSvc *service.StatsService,
	logger *zap.Logger,
) *MCPServer {
	s := server.NewMCPServer(
		"FriendBook AI",                  // 服务名（外部客户端可见）
		"1.0.0",                          // 版本
		server.WithToolCapabilities(true), // 声明支持工具调用
	)

	m := &MCPServer{
		server:    s,
		friendSvc: friendSvc,
		statsSvc:  statsSvc,
		logger:    logger,
	}

	m.registerFriendTools()
	m.registerStatsTools()

	logger.Info("MCP Server 已就绪")
	return m
}

// GetServer 返回底层 MCP Server，供启动 SSE 端点用
func (m *MCPServer) GetServer() *server.MCPServer {
	return m.server
}

// ==================== 好友工具 ====================

func (m *MCPServer) registerFriendTools() {
	// ---- 创建好友：链式调用声明参数 schema（与 Eino 的 struct tag 推导不同）----
	createTool := mcp.NewTool("create_friend",
		mcp.WithDescription("创建一个新的好友记录"),
		mcp.WithString("name", mcp.Required(), mcp.Description("好友姓名")),
		mcp.WithString("remark", mcp.Description("备注名")),
		mcp.WithString("phone", mcp.Description("手机号")),
		mcp.WithString("wechat", mcp.Description("微信号")),
		mcp.WithString("location", mcp.Description("所在城市")),
		mcp.WithString("occupation", mcp.Description("职业")),
		mcp.WithNumber("favorability", mcp.Description("好感度 0-100，默认 50")),
	)
	m.server.AddTool(createTool, m.handleCreateFriend)

	// ---- 查询好友详情 ----
	getTool := mcp.NewTool("get_friend",
		mcp.WithDescription("根据好友ID查询好友详细信息"),
		mcp.WithNumber("friend_id", mcp.Required(), mcp.Description("好友ID")),
	)
	m.server.AddTool(getTool, m.handleGetFriend)

	// ---- 好友列表 ----
	listTool := mcp.NewTool("list_friends",
		mcp.WithDescription("列出好友，支持按关键词搜索、按温度/好感度筛选"),
		mcp.WithString("keyword", mcp.Description("搜索关键词（姓名/备注/手机号）")),
		mcp.WithNumber("temperature", mcp.Description("关系温度筛选 1夯 2热 3温 4凉 5拉")),
		mcp.WithNumber("min_favor", mcp.Description("最低好感度")),
		mcp.WithNumber("max_favor", mcp.Description("最高好感度")),
		mcp.WithNumber("page", mcp.Description("页码，默认1")),
		mcp.WithNumber("page_size", mcp.Description("每页条数，默认10")),
	)
	m.server.AddTool(listTool, m.handleListFriends)
}

// getString / getInt / getFloat 参数取值辅助（MCP 参数是 map[string]any）
func getString(args map[string]any, key string) string {
	v, _ := args[key].(string)
	return v
}

func getFloat(args map[string]any, key string) float64 {
	v, _ := args[key].(float64) // JSON 数字统一反序列化为 float64
	return v
}

func getInt(args map[string]any, key string) int {
	return int(getFloat(args, key))
}

// handleCreateFriend create_friend 工具处理函数
func (m *MCPServer) handleCreateFriend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Params.Arguments 本身就是 map[string]interface{} 类型，直接使用即可
	args := req.Params.Arguments
	wechat := getString(args, "wechat")

	favorability := 50
	if v := getInt(args, "favorability"); v > 0 {
		favorability = v
	}

	p := &service.CreateFriendParams{
		Name:         getString(args, "name"),
		Remark:       getString(args, "remark"),
		Phone:        getString(args, "phone"),
		Location:     getString(args, "location"),
		Occupation:   getString(args, "occupation"),
		Favorability: favorability,
		UserID:       MCPUserID, // TODO: 生产环境从 MCP 会话身份认证中获取
	}
	if wechat != "" {
		p.Accounts = append(p.Accounts, service.AccountParam{Platform: "wechat", Account: wechat})
	}

	friend, err := m.friendSvc.Create(ctx, p)
	if err != nil {
		m.logger.Error("MCP create_friend 失败", zap.Error(err))
		// 工具执行错误以 ToolResultError 返回（协议层仍 200，错误信息给到客户端）
		return mcp.NewToolResultError(fmt.Sprintf("创建失败: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf(
		"已创建好友「%s」，ID: %d，好感度: %d",
		friend.Name, friend.ID, friend.Favorability,
	)), nil
}

// handleGetFriend get_friend 工具处理函数
func (m *MCPServer) handleGetFriend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	friendID := getInt(args, "friend_id")

	friend, err := m.friendSvc.GetByID(ctx, MCPUserID, uint(friendID))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(friend.ToSummary()), nil
}

// handleListFriends list_friends 工具处理函数
func (m *MCPServer) handleListFriends(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments

	result, err := m.friendSvc.List(ctx, MCPUserID, &repository.FriendQuery{
		Keyword:     getString(args, "keyword"),
		Temperature: getInt(args, "temperature"),
		MinFavor:    getInt(args, "min_favor"),
		MaxFavor:    getInt(args, "max_favor"),
		Page:        getInt(args, "page"),
		PageSize:    getInt(args, "page_size"),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询失败: %v", err)), nil
	}
	return mcp.NewToolResultText(result.ToSummary()), nil
}

// ==================== 统计工具 ====================

func (m *MCPServer) registerStatsTools() {
	statsTool := mcp.NewTool(
		"get_friend_stats",
		mcp.WithDescription("获取社交圈统计：好友总数、温度分布、关系变凉数、平均好感度"),
	)
	m.server.AddTool(statsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stats, err := m.statsSvc.GetOverview(ctx, MCPUserID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("统计失败: %v", err)), nil
		}
		return mcp.NewToolResultText(stats.ToSummary()), nil
	})
}
