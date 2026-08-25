// Package ai AI 助手模块：Eino ReAct Agent + 原生 Tools。
// 核心设计：工具执行函数直接调用 Service 层（与 HTTP Handler 共享业务逻辑），
// user_id 通过 context.Context 透传，保证数据隔离。
package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"go.uber.org/zap"

	"golang-learn/internal/repository"
	"golang-learn/internal/service"
)

// ToolCtxKey context key：Handler 层把 user_id 注入 context，
// 工具执行函数再取出来调用 Service（key 用空结构体，避免冲突）
type ToolCtxKey struct{}

// Tools 工具集：持有所有 Service 引用，注册 Eino 原生工具
type Tools struct {
	tools       []tool.BaseTool
	friendSvc   *service.FriendService
	interactSvc *service.InteractionService
	statsSvc    *service.StatsService
	reminderSvc *service.ReminderService
	logger      *zap.Logger
}

// NewTools 创建工具集并注册全部工具
func NewTools(
	friendSvc *service.FriendService,
	interactSvc *service.InteractionService,
	statsSvc *service.StatsService,
	reminderSvc *service.ReminderService,
	logger *zap.Logger,
) *Tools {
	t := &Tools{
		friendSvc:   friendSvc,
		interactSvc: interactSvc,
		statsSvc:    statsSvc,
		reminderSvc: reminderSvc,
		logger:      logger,
	}
	t.registerFriendTools()
	t.registerInteractionTools()
	t.registerStatsTools()
	t.registerReminderTools()
	return t
}

// GetTools 返回所有工具，供 Agent 注册
func (t *Tools) GetTools() []tool.BaseTool {
	return t.tools
}

// getUserID 从 context 获取当前用户 ID（未注入时返回 0，
// Service 层查询会因 user_id=0 查不到数据，天然安全）
func (t *Tools) getUserID(ctx context.Context) uint {
	if v, ok := ctx.Value(ToolCtxKey{}).(uint); ok {
		return v
	}
	return 0
}

// ==================== 好友相关工具 ====================

func (t *Tools) registerFriendTools() {
	// ---- 创建好友：struct tag 自动推导 JSON Schema ----
	type createFriendArgs struct {
		Name         string `json:"name" jsonschema:"required" jsonschema_description:"好友姓名"`
		Remark       string `json:"remark" jsonschema_description:"备注名"`
		Phone        string `json:"phone" jsonschema_description:"手机号"`
		Wechat       string `json:"wechat" jsonschema_description:"微信号"`
		Location     string `json:"location" jsonschema_description:"所在城市"`
		Occupation   string `json:"occupation" jsonschema_description:"职业"`
		Favorability int    `json:"favorability" jsonschema_description:"好感度 0-100，默认50"`
	}
	toolCreate, _ := utils.InferTool(
		"create_friend",
		"创建一个新的好友记录",
		func(ctx context.Context, args *createFriendArgs) (string, error) {
			params := &service.CreateFriendParams{
				Name:         args.Name,
				Remark:       args.Remark,
				Phone:        args.Phone,
				Location:     args.Location,
				Occupation:   args.Occupation,
				Favorability: args.Favorability,
				UserID:       t.getUserID(ctx),
			}
			// 微信号单独入参，转为平台账号
			if args.Wechat != "" {
				params.Accounts = append(params.Accounts, service.AccountParam{
					Platform: "wechat", Account: args.Wechat,
				})
			}
			friend, err := t.friendSvc.Create(ctx, params)
			if err != nil {
				t.logger.Error("tool create_friend failed", zap.Error(err))
				// 错误以文本返回给 LLM（而非中断），让它能向用户解释原因
				return fmt.Sprintf("创建失败: %v", err), nil
			}
			return fmt.Sprintf("已创建好友「%s」，ID: %d，好感度: %d",
				friend.Name, friend.ID, friend.Favorability), nil
		},
	)
	t.tools = append(t.tools, toolCreate)

	// ---- 查询好友详情 ----
	type getFriendArgs struct {
		FriendID int `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
	}
	toolGet, _ := utils.InferTool(
		"get_friend",
		"根据好友ID查询好友详细信息",
		func(ctx context.Context, args *getFriendArgs) (string, error) {
			friend, err := t.friendSvc.GetByID(ctx, t.getUserID(ctx), uint(args.FriendID))
			if err != nil {
				return fmt.Sprintf("查询失败: %v", err), nil
			}
			return friend.ToSummary(), nil // 摘要格式，控制 token
		},
	)
	t.tools = append(t.tools, toolGet)

	// ---- 列出好友（多条件筛选）----
	type listFriendsArgs struct {
		Keyword     string `json:"keyword" jsonschema_description:"搜索关键词（姓名/备注/手机号）"`
		Temperature int    `json:"temperature" jsonschema_description:"关系温度筛选 1夯 2热 3温 4凉 5拉"`
		MinFavor    int    `json:"min_favor" jsonschema_description:"最低好感度"`
		MaxFavor    int    `json:"max_favor" jsonschema_description:"最高好感度"`
		Page        int    `json:"page" jsonschema_description:"页码，默认1"`
		PageSize    int    `json:"page_size" jsonschema_description:"每页条数，默认10"`
	}
	toolList, _ := utils.InferTool(
		"list_friends",
		"列出好友，支持按关键词搜索、按温度/好感度筛选",
		func(ctx context.Context, args *listFriendsArgs) (string, error) {
			result, err := t.friendSvc.List(ctx, t.getUserID(ctx), &repository.FriendQuery{
				Keyword:     args.Keyword,
				Temperature: args.Temperature,
				MinFavor:    args.MinFavor,
				MaxFavor:    args.MaxFavor,
				Page:        args.Page,
				PageSize:    args.PageSize,
			})
			if err != nil {
				return fmt.Sprintf("查询失败: %v", err), nil
			}
			return result.ToSummary(), nil
		},
	)
	t.tools = append(t.tools, toolList)

	// ---- 更新好友 ----
	type updateFriendArgs struct {
		FriendID   int    `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
		Name       string `json:"name" jsonschema_description:"新姓名"`
		Phone      string `json:"phone" jsonschema_description:"新手机号"`
		Location   string `json:"location" jsonschema_description:"新城市"`
		Occupation string `json:"occupation" jsonschema_description:"新职业"`
	}
	toolUpdate, _ := utils.InferTool(
		"update_friend",
		"更新好友的基本信息（只传需要修改的字段）",
		func(ctx context.Context, args *updateFriendArgs) (string, error) {
			p := &service.UpdateFriendParams{}
			if args.Name != "" {
				p.Name = &args.Name
			}
			if args.Phone != "" {
				p.Phone = &args.Phone
			}
			if args.Location != "" {
				p.Location = &args.Location
			}
			if args.Occupation != "" {
				p.Occupation = &args.Occupation
			}
			if err := t.friendSvc.Update(ctx, t.getUserID(ctx), uint(args.FriendID), p); err != nil {
				return fmt.Sprintf("更新失败: %v", err), nil
			}
			return fmt.Sprintf("已更新好友 ID: %d", args.FriendID), nil
		},
	)
	t.tools = append(t.tools, toolUpdate)

	// ---- 删除好友 ----
	type deleteFriendArgs struct {
		FriendID int `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
	}
	toolDelete, _ := utils.InferTool(
		"delete_friend",
		"根据好友ID删除好友（危险操作，需向用户确认后调用）",
		func(ctx context.Context, args *deleteFriendArgs) (string, error) {
			if err := t.friendSvc.Delete(ctx, t.getUserID(ctx), uint(args.FriendID)); err != nil {
				return fmt.Sprintf("删除失败: %v", err), nil
			}
			return fmt.Sprintf("已删除好友 ID: %d", args.FriendID), nil
		},
	)
	t.tools = append(t.tools, toolDelete)

	// ---- 调整好感度 ----
	type adjustFavorArgs struct {
		FriendID     int `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
		Favorability int `json:"favorability" jsonschema:"required" jsonschema_description:"新的好感度值 0-100"`
	}
	toolAdjust, _ := utils.InferTool(
		"adjust_favorability",
		"调整好友的好感度（0-100），会联动更新关系温度",
		func(ctx context.Context, args *adjustFavorArgs) (string, error) {
			if err := t.friendSvc.AdjustFavorability(ctx, t.getUserID(ctx),
				uint(args.FriendID), args.Favorability); err != nil {
				return fmt.Sprintf("调整失败: %v", err), nil
			}
			return fmt.Sprintf("已将好友 ID %d 的好感度调整为 %d",
				args.FriendID, args.Favorability), nil
		},
	)
	t.tools = append(t.tools, toolAdjust)
}

// ==================== 互动相关工具 ====================

func (t *Tools) registerInteractionTools() {
	// ---- 记录互动 ----
	type addInteractionArgs struct {
		FriendID int    `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
		Type     int    `json:"type" jsonschema_description:"互动类型 1聊天 2见面 3通话 4聚餐 5送礼 6其他"`
		Content  string `json:"content" jsonschema_description:"互动内容摘要"`
		Location string `json:"location" jsonschema_description:"地点"`
		Mood     string `json:"mood" jsonschema_description:"心情"`
	}
	toolAdd, _ := utils.InferTool(
		"add_interaction",
		"记录一次与好友的互动",
		func(ctx context.Context, args *addInteractionArgs) (string, error) {
			_, err := t.interactSvc.Create(ctx, t.getUserID(ctx), &service.InteractionParams{
				FriendID: args.FriendID, Type: args.Type,
				Content: args.Content, Location: args.Location, Mood: args.Mood,
			})
			if err != nil {
				return fmt.Sprintf("记录失败: %v", err), nil
			}
			return "互动记录已添加", nil
		},
	)
	t.tools = append(t.tools, toolAdd)

	// ---- 查看互动历史（供 AI 引用"共同记忆"生成破冰话术）----
	type listInteractionsArgs struct {
		FriendID int `json:"friend_id" jsonschema:"required" jsonschema_description:"好友ID"`
	}
	toolListInter, _ := utils.InferTool(
		"list_interactions",
		"查看与某位好友的历史互动记录（最近10条）",
		func(ctx context.Context, args *listInteractionsArgs) (string, error) {
			list, err := t.interactSvc.ListByFriendID(ctx, t.getUserID(ctx), uint(args.FriendID))
			if err != nil {
				return fmt.Sprintf("查询失败: %v", err), nil
			}
			if len(list) == 0 {
				return "暂无互动记录", nil
			}
			result := "互动记录：\n"
			max := 10
			if len(list) < max {
				max = len(list)
			}
			for i := 0; i < max; i++ {
				it := list[i]
				result += fmt.Sprintf("%d. [%s] %s %s\n",
					i+1, it.InteractAt.Format("2006-01-02"),
					it.Type.String(), it.Content)
			}
			return result, nil
		},
	)
	t.tools = append(t.tools, toolListInter)
}

// ==================== 统计相关工具 ====================

func (t *Tools) registerStatsTools() {
	// ---- 社交圈统计（AI 分析的核心数据源）----
	toolStats, _ := utils.InferTool(
		"get_friend_stats",
		"获取当前用户的社交圈统计数据，包括好友总数、温度分布、关系变凉数等。用于社交圈整体分析。",
		func(ctx context.Context, args *struct{}) (string, error) {
			stats, err := t.statsSvc.GetOverview(ctx, t.getUserID(ctx))
			if err != nil {
				return fmt.Sprintf("统计失败: %v", err), nil
			}
			return stats.ToSummary(), nil
		},
	)
	t.tools = append(t.tools, toolStats)
}

// ==================== 提醒相关工具 ====================

func (t *Tools) registerReminderTools() {
	type getRemindersArgs struct {
		Type string `json:"type" jsonschema_description:"today今日 upcoming近期，默认today"`
	}
	toolReminders, _ := utils.InferTool(
		"get_reminders",
		"获取生日/纪念日提醒，type 可选 today 或 upcoming",
		func(ctx context.Context, args *getRemindersArgs) (string, error) {
			reminders, err := t.reminderSvc.GetReminders(ctx, t.getUserID(ctx), args.Type)
			if err != nil {
				return fmt.Sprintf("查询失败: %v", err), nil
			}
			if len(reminders) == 0 {
				return "暂无提醒", nil
			}
			result := "提醒列表:\n"
			for i, r := range reminders {
				result += fmt.Sprintf("%d. %s（%s 提前%d天）\n",
					i+1, r.Title, r.Date.Format("2006-01-02"), r.DaysInAdvance)
			}
			return result, nil
		},
	)
	t.tools = append(t.tools, toolReminders)
}
