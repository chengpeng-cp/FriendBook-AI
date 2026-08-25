package handler

import (
	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// StatsHandler 统计看板相关接口（全部带 Redis 缓存，TTL 5 分钟）
type StatsHandler struct {
	statsSvc *service.StatsService
}

// NewStatsHandler 创建统计 Handler
func NewStatsHandler(statsSvc *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

// Overview 统计概览
// @Summary 统计概览
// @Tags 数据统计
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response{data=service.StatsOverview}
// @Router /api/v1/stats/overview [get]
func (h *StatsHandler) Overview(c *gin.Context) {
	userID := c.GetUint("user_id")

	data, err := h.statsSvc.GetOverview(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, data)
}

// PlatformDistribution 平台分布
// @Summary 好友平台分布
// @Tags 数据统计
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response}
// @Router /api/v1/stats/platform-distribution [get]
func (h *StatsHandler) PlatformDistribution(c *gin.Context) {
	userID := c.GetUint("user_id")

	result, err := h.statsSvc.PlatformDistribution(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	// 空数据兜底：Go nil slice 会序列化为 "list":null，
	// 前端拿到 null 会误判为"仍在加载"而永久转圈，这里统一补成空数组
	if result == nil {
		result = []service.PlatformDistItem{}
	}
	response.OK(c, gin.H{"list": result})
}

// TemperatureDistribution 关系温度分布
// @Summary 关系温度分布
// @Tags 数据统计
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response{data=service.TemperatureDist}
// @Router /api/v1/stats/temperature-distribution [get]
func (h *StatsHandler) TemperatureDistribution(c *gin.Context) {
	userID := c.GetUint("user_id")

	result, err := h.statsSvc.TemperatureDistribution(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, result)
}

// InteractionTrend 近 30 天互动趋势
// @Summary 互动趋势
// @Tags 数据统计
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response}
// @Router /api/v1/stats/interaction-trend [get]
func (h *StatsHandler) InteractionTrend(c *gin.Context) {
	userID := c.GetUint("user_id")

	result, err := h.statsSvc.InteractionTrend(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	// 空数据兜底：同 PlatformDistribution，避免 "list":null 让前端永久转圈
	if result == nil {
		result = []service.TrendPoint{}
	}
	response.OK(c, gin.H{"list": result})
}
