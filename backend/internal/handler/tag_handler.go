package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// TagHandler 标签相关接口
type TagHandler struct {
	tagSvc *service.TagService
}

// NewTagHandler 创建标签 Handler
func NewTagHandler(tagSvc *service.TagService) *TagHandler {
	return &TagHandler{tagSvc: tagSvc}
}

// CreateTagReq 创建标签请求
type CreateTagReq struct {
	Name  string `json:"name" binding:"required,min=1,max=32"` // 标签名
	Color string `json:"color" binding:"max=16"`                // 颜色（如 #409EFF）
}

// Create 创建标签
// @Summary 创建标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body CreateTagReq true "标签信息"
// @Success 200 {object} response.Response{data=model.Tag}
// @Router /api/v1/tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	var req CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	tag, err := h.tagSvc.Create(c, userID, req.Name, req.Color)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, tag, "创建成功")
}

// Delete 删除标签
// @Summary 删除标签
// @Tags 标签管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "标签ID"
// @Success 200 {object} response.Response
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	tagID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	if err := h.tagSvc.Delete(c, userID, uint(tagID)); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "删除成功")
}

// List 标签列表
// @Summary 标签列表
// @Tags 标签管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response
// @Router /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	tags, err := h.tagSvc.List(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": tags, "count": len(tags)})
}
