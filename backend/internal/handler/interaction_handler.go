package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// InteractionHandler 互动记录相关接口
type InteractionHandler struct {
	interactSvc *service.InteractionService
}

// NewInteractionHandler 创建互动记录 Handler
func NewInteractionHandler(interactSvc *service.InteractionService) *InteractionHandler {
	return &InteractionHandler{interactSvc: interactSvc}
}

// CreateInteractionReq 创建互动记录请求
type CreateInteractionReq struct {
	FriendID    uint     `json:"friend_id" binding:"required"`     // 好友 ID
	Type        int      `json:"type" binding:"gte=0,lte=6"`       // 互动类型 1-6
	Content     string   `json:"content"`                          // 内容摘要
	Location    string   `json:"location" binding:"max=128"`       // 地点
	Mood        string   `json:"mood" binding:"max=32"`            // 心情
	PhotoURLs   []string `json:"photo_urls"`                       // 照片列表
	IsMemory    bool     `json:"is_memory"`                        // 是否记忆卡片
	MemoryTitle string   `json:"memory_title" binding:"max=64"`    // 记忆卡片标题
	InteractAt  string   `json:"interact_at" binding:"omitempty,datetime=2006-01-02 15:04:05"` // 互动时间（缺省为当前）
}

// Create 创建互动记录（联动刷新好友最近互动时间与关系温度）
// @Summary 创建互动记录
// @Tags 互动记录
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body CreateInteractionReq true "互动信息"
// @Success 200 {object} response.Response{data=model.Interaction}
// @Router /api/v1/interactions [post]
func (h *InteractionHandler) Create(c *gin.Context) {
	var req CreateInteractionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	p := &service.InteractionParams{
		FriendID:    int(req.FriendID),
		Type:        req.Type,
		Content:     req.Content,
		Location:    req.Location,
		Mood:        req.Mood,
		PhotoURLs:   req.PhotoURLs,
		IsMemory:    req.IsMemory,
		MemoryTitle: req.MemoryTitle,
	}
	// 互动时间字符串 → time.Time（缺省由 Service 填当前时间）
	if req.InteractAt != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.InteractAt); err == nil {
			p.InteractAt = t
		}
	}

	interaction, err := h.interactSvc.Create(c, userID, p)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, interaction, "记录成功")
}

// UpdateInteractionReq 更新互动记录请求
type UpdateInteractionReq struct {
	Type       int    `json:"type" binding:"gte=0,lte=6"`
	Content    string `json:"content"`
	Location   string `json:"location"`
	Mood       string `json:"mood"`
	InteractAt string `json:"interact_at" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}

// Update 更新互动记录
// @Summary 更新互动记录
// @Tags 互动记录
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "互动ID"
// @Param body body UpdateInteractionReq true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/v1/interactions/{id} [put]
func (h *InteractionHandler) Update(c *gin.Context) {
	interactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	var req UpdateInteractionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	p := &service.InteractionParams{
		Type:     req.Type,
		Content:  req.Content,
		Location: req.Location,
		Mood:     req.Mood,
	}
	if req.InteractAt != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.InteractAt); err == nil {
			p.InteractAt = t
		}
	}

	if err := h.interactSvc.Update(c, userID, uint(interactionID), p); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "更新成功")
}

// Delete 删除互动记录
// @Summary 删除互动记录
// @Tags 互动记录
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "互动ID"
// @Success 200 {object} response.Response
// @Router /api/v1/interactions/{id} [delete]
func (h *InteractionHandler) Delete(c *gin.Context) {
	interactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	if err := h.interactSvc.Delete(c, userID, uint(interactionID)); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "删除成功")
}

// List 互动记录列表（可按好友筛选）
// @Summary 互动记录列表
// @Tags 互动记录
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param friend_id query int false "好友ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/interactions [get]
func (h *InteractionHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	friendID := uint(queryInt(c, "friend_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := h.interactSvc.List(c, userID, friendID, page, pageSize)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
