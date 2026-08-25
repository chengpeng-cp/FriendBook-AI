package handler

import (
	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// ReminderHandler 提醒相关接口
type ReminderHandler struct {
	reminderSvc *service.ReminderService
}

// NewReminderHandler 创建提醒 Handler
func NewReminderHandler(reminderSvc *service.ReminderService) *ReminderHandler {
	return &ReminderHandler{reminderSvc: reminderSvc}
}

// GetToday 获取今日提醒（含未来 3 天，提前提醒语义；带 Redis 缓存）
// @Summary 获取今日提醒
// @Tags 提醒
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response}
// @Router /api/v1/reminders/today [get]
func (h *ReminderHandler) GetToday(c *gin.Context) {
	userID := c.GetUint("user_id")

	reminders, err := h.reminderSvc.GetTodayReminders(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": reminders, "count": len(reminders)})
}

// GetUpcoming 获取近期提醒（未来 7 天）
// @Summary 获取近期提醒
// @Tags 提醒
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response
// @Router /api/v1/reminders/upcoming [get]
func (h *ReminderHandler) GetUpcoming(c *gin.Context) {
	userID := c.GetUint("user_id")

	reminders, err := h.reminderSvc.GetUpcomingReminders(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": reminders, "count": len(reminders)})
}
