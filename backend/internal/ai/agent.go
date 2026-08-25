package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// AgentConfig 大模型配置
type AgentConfig struct {
	OpenAIKey     string
	OpenAIModel   string // 如 "gpt-4o-mini"
	OpenAIBaseURL string // 兼容 OpenAI 协议的中转/国产模型地址
}

// Agent Eino ReAct Agent 封装：
// ChatModel 负责推理，Tools 负责落库，ReAct 循环负责"思考→调用工具→观察→再思考"
type Agent struct {
	reactAgent *react.Agent
	tools      *Tools
	logger     *zap.Logger
}

// systemPrompt 人设与行为约束（Persona）。
// 包含社交分析的推荐工作流与破冰话术要点，让模型稳定输出高质量建议。
const systemPrompt = `你是 FriendBook AI，一个智能好友管理助手兼社交分析师。

【你的核心能力】
1. 好友管理：创建、查询、更新、删除好友
2. 互动记录：记录与好友的每次互动
3. 提醒查询：查询今日/近期生日提醒
4. 社交分析：分析用户的社交圈、给出维护建议、生成破冰话术

【分析社交圈的推荐流程】
当用户请求"分析社交圈""给建议""该联系谁"时，按以下顺序：
1. 调用 get_friend_stats 获取整体统计
2. 如果发现"关系变凉"数量较多，调用 list_friends 筛选 temperature=4（凉）或 5（拉）的好友
3. 针对最值得联系的好友，调用 get_friend 拿到详细画像（MBTI、职业、认识场景等）
4. 综合给出分析报告 + 具体行动建议

【生成破冰话术的要点】
- 结合好友的 MBTI、职业、爱好、认识场景
- 引用共同记忆（如果有互动记录，可用 list_interactions 查询）
- 给出 2-3 个具体可执行的开场白示例
- 话术要自然、有温度、不像模板

【回复风格】
- 中文回复，自然口语化
- 数据要具体（不要说"很多好友"，要说"50 个好友里有 5 个变凉"）
- 建议要可执行（不要说"多联系"，要说"本周给小刚发条消息问问他最近的项目"）
- 控制篇幅，重要结论先行`

// NewAgent 创建 ReAct Agent：
//  1. 初始化 OpenAI 兼容 ChatModel
//  2. 注册全部 Eino Tools
//  3. 注入 Persona（系统提示词）
func NewAgent(ctx context.Context, cfg AgentConfig, tools *Tools, logger *zap.Logger) (*Agent, error) {
	// 1. ChatModel（OpenAI 兼容协议）
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.OpenAIBaseURL,
		APIKey:  cfg.OpenAIKey,
		Model:   cfg.OpenAIModel,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 打印已注册工具清单（便于排查"AI 不会用某工具"问题）
	logger.Info("已加载 Eino Tools", zap.Int("count", len(tools.GetTools())))
	for _, t := range tools.GetTools() {
		if info, err := t.Info(ctx); err == nil {
			logger.Info("  工具", zap.String("name", info.Name), zap.String("desc", info.Desc))
		}
	}

	// 2. ReAct Agent（MaxStep 默认 10 不够多工具串联，调到 20）
	reactAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools.GetTools(), // 直接注入 Eino 原生 Tools
		},
		MessageModifier: react.NewPersonaModifier(systemPrompt),
		MaxStep:         20,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Agent 失败: %w", err)
	}

	return &Agent{
		reactAgent: reactAgent,
		tools:      tools,
		logger:     logger,
	}, nil
}

// Chat 同步对话（一次性返回完整回复）
func (a *Agent) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return a.reactAgent.Generate(ctx, messages)
}

// StreamChat 流式对话（SSE 逐 token 返回，用户体验更好）
func (a *Agent) StreamChat(ctx context.Context, messages []*schema.Message) (
	*schema.StreamReader[*schema.Message], error,
) {
	return a.reactAgent.Stream(ctx, messages)
}
