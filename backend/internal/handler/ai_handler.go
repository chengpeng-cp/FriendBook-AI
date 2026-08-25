package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"golang-learn/internal/ai"
	"golang-learn/internal/model"
	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/repository"
)

// AIHandler AI 助手接口
type AIHandler struct {
	agent   *ai.Agent
	chatRepo *repository.ChatRepo
	logger  *zap.Logger
}

// NewAIHandler 创建 AI Handler
func NewAIHandler(agent *ai.Agent, chatRepo *repository.ChatRepo, logger *zap.Logger) *AIHandler {
	return &AIHandler{agent: agent, chatRepo: chatRepo, logger: logger}
}

// ChatRequest AI 对话请求
type ChatRequest struct {
	Message string            `json:"message" binding:"required"` // 本次用户输入
	History []*schema.Message `json:"history"`                    // 历史对话（多轮上下文）
}

// Chat AI 对话（SSE 流式输出）
// @Summary AI 对话（SSE 流式）
// @Description 发送消息给 AI 助手，返回 SSE 流。事件格式：{"content":"文本","tool_call":"工具名","error":"错误信息","done":bool}
// @Tags AI助手
// @Accept json
// @Produce text/event-stream
// @Param Authorization header string true "Bearer <token>"
// @Param body body ChatRequest true "对话请求"
// @Success 200 {string} string "SSE 流"
// @Router /api/v1/ai/chat [post]
func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	userID := c.GetUint("user_id")

	// 组装消息：历史 + 当前输入（系统提示词由 Agent 的 Persona 注入，这里不用加）
	messages := append(req.History, schema.UserMessage(req.Message))

	// 把 user_id 注入 context：Eino Tools 执行时从这里取当前用户，保证数据隔离
	ctx := context.WithValue(c.Request.Context(), ai.ToolCtxKey{}, userID)

	streamReader, err := h.agent.StreamChat(ctx, messages)
	if err != nil {
		response.FailWithMsg(c, errcode.AIChatError, err.Error())
		return
	}
	defer streamReader.Close()

	// SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 拼接完整回复：流结束后连同用户消息一起落库（聊天记录持久化）
	var replyBuf []byte

	// 循环读取流，逐帧推送
	c.Stream(func(w io.Writer) bool {
		msg, err := streamReader.Recv()
		if err != nil {
			// 流结束。EOF 属正常收尾；其他错误必须以 error 帧透传给前端，
			// 否则前端只会看到"发出去没有回复"的静默失败
			if !errors.Is(err, io.EOF) {
				h.logger.Error("AI 流式输出中断", zap.Uint("user_id", userID), zap.Error(err))
				frame, _ := json.Marshal(map[string]any{
					"error": "AI 服务暂时不可用，请稍后重试",
					"done":  true,
				})
				c.SSEvent("message", string(frame))
			}
			return false
		}

		// 提取有效工具调用名（过滤空名调用，部分模型会产出 name 为空的 ToolCall）
		var validToolCalls []schema.ToolCall
		for _, tc := range msg.ToolCalls {
			if tc.Function.Name != "" {
				validToolCalls = append(validToolCalls, tc)
			}
		}
		// 过滤空帧：content 为空且无有效工具调用的思考帧，不推送（减少无效流量）
		if msg.Content == "" && len(validToolCalls) == 0 {
			return true
		}

		replyBuf = append(replyBuf, msg.Content...)

		event := map[string]any{
			"content": msg.Content,
			"done":    false,
		}
		// 工具调用帧：前端可展示"正在调用 xxx 工具"的提示
		for _, tc := range validToolCalls {
			event["tool_call"] = tc.Function.Name
			event["tool_args"] = tc.Function.Arguments
		}
		data, _ := json.Marshal(event)
		c.SSEvent("message", string(data))
		return true
	})

	// ======== 空回复兜底（必须在 done 帧之前推送，前端收到 done 即停止解析） ========
	// 免费模型偶发"只调工具不说话"或整体返回空白：此时给前端明确提示，
	// 避免用户看到"消息发出去了却没有回复"的静默失败
	reply := strings.TrimSpace(string(replyBuf))
	if reply == "" {
		h.logger.Warn("AI 返回空回复", zap.Uint("user_id", userID), zap.Int("raw_len", len(replyBuf)))
		c.SSEvent("message", `{"content":"（AI 本次没有生成有效回复，可能是模型服务波动，请重试一次）","done":false}`)
		c.SSEvent("message", `{"content":"","done":true}`)
		return // 空回复不落库，避免历史里出现无意义记录
	}

	// 发送结束标记（前端据此关闭 loading）
	c.SSEvent("message", `{"content":"","done":true}`)

	// ======== 聊天记录持久化（失败不影响对话主流程） ========
	msgs := []*model.ChatMessage{
		{UserID: userID, Role: "user", Content: req.Message},
		{UserID: userID, Role: "assistant", Content: reply},
	}
	if err := h.chatRepo.CreateBatch(context.Background(), msgs); err != nil {
		h.logger.Warn("保存聊天记录失败", zap.Uint("user_id", userID), zap.Error(err))
	}
}

// History 查询 AI 对话历史
// @Summary AI 对话历史
// @Description 返回当前用户最近的历史对话（时间正序），默认 50 条
// @Tags AI助手
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response
// @Router /api/v1/ai/history [get]
func (h *AIHandler) History(c *gin.Context) {
	userID := c.GetUint("user_id")
	msgs, err := h.chatRepo.ListRecent(c.Request.Context(), userID, 50)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": msgs})
}

// ClearHistory 清空 AI 对话历史
// @Summary 清空 AI 对话历史
// @Tags AI助手
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response
// @Router /api/v1/ai/history [delete]
func (h *AIHandler) ClearHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	if err := h.chatRepo.Clear(c.Request.Context(), userID); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "已清空对话记录")
}
