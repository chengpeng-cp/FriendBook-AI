package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// GroupHandler 分组相关接口
type GroupHandler struct {
	groupSvc *service.GroupService
}

// NewGroupHandler 创建分组 Handler
func NewGroupHandler(groupSvc *service.GroupService) *GroupHandler {
	return &GroupHandler{groupSvc: groupSvc}
}

// CreateGroupReq 创建分组请求
type CreateGroupReq struct {
	Name      string `json:"name" binding:"required,min=1,max=64"` // 分组名
	SortOrder int    `json:"sort_order"`                            // 排序值
}

// Create 创建分组
// @Summary 创建分组
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body CreateGroupReq true "分组信息"
// @Success 200 {object} response.Response{data=model.Group}
// @Router /api/v1/groups [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var req CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	group, err := h.groupSvc.Create(c, userID, req.Name, req.SortOrder)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, group, "创建成功")
}

// UpdateGroupReq 更新分组请求
type UpdateGroupReq struct {
	Name      string `json:"name" binding:"omitempty,min=1,max=64"`
	SortOrder *int   `json:"sort_order"` // 指针区分是否传入
}

// Update 更新分组
// @Summary 更新分组
// @Tags 分组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "分组ID"
// @Param body body UpdateGroupReq true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/v1/groups/{id} [put]
func (h *GroupHandler) Update(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	var req UpdateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	if err := h.groupSvc.Update(c, userID, uint(groupID), req.Name, sortOrder); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "更新成功")
}

// Delete 删除分组
// @Summary 删除分组
// @Tags 分组管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "分组ID"
// @Success 200 {object} response.Response
// @Router /api/v1/groups/{id} [delete]
func (h *GroupHandler) Delete(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	if err := h.groupSvc.Delete(c, userID, uint(groupID)); err != nil {
		response.Fail(c, errcode.GroupNotFound)
		return
	}
	response.OKWithMsg(c, nil, "删除成功")
}

// List 分组列表
// @Summary 分组列表
// @Tags 分组管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response
// @Router /api/v1/groups [get]
func (h *GroupHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	groups, err := h.groupSvc.List(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": groups, "count": len(groups)})
}
