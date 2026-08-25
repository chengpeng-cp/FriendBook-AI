package handler

import (
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"golang-learn/internal/model"
	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/repository"
	"golang-learn/internal/service"
)

// FriendHandler 好友相关接口
type FriendHandler struct {
	friendSvc *service.FriendService
}

// NewFriendHandler 创建好友 Handler
func NewFriendHandler(friendSvc *service.FriendService) *FriendHandler {
	return &FriendHandler{friendSvc: friendSvc}
}

// CreateFriendReq 创建好友请求
type CreateFriendReq struct {
	Name         string                `json:"name" binding:"required,min=1,max=64"` // 姓名
	Remark       string                `json:"remark" binding:"max=64"`              // 备注
	Gender       int8                  `json:"gender" binding:"oneof=0 1 2"`         // 性别
	Birthday     string                `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // 生日
	Phone        string                `json:"phone" binding:"omitempty,max=20"`      // 手机号（宽松校验：兼容 138xxx / +86138xxx 两种写法）
	Email        string                `json:"email" binding:"omitempty,email"`      // 邮箱
	Location     string                `json:"location" binding:"max=64"`            // 城市
	Occupation   string                `json:"occupation" binding:"max=64"`          // 职业
	Bio          string                `json:"bio" binding:"max=255"`                // 简介
	MeetScene    string                `json:"meet_scene" binding:"max=128"`         // 认识场景
	Favorability int                   `json:"favorability" binding:"gte=0,lte=100"` // 好感度
	MBTI         string                `json:"mbti" binding:"omitempty,len=4"`       // MBTI
	Accounts     []service.AccountParam `json:"accounts"`                           // 平台账号
	GroupIDs     []uint                `json:"group_ids" binding:"omitempty,dive,required"` // 分组
	TagIDs       []uint                `json:"tag_ids" binding:"omitempty,dive,required"`   // 标签
}

// Create 创建好友
// @Summary 创建好友
// @Tags 好友管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body CreateFriendReq true "好友信息"
// @Success 200 {object} response.Response{data=model.Friend}
// @Router /api/v1/friends [post]
func (h *FriendHandler) Create(c *gin.Context) {
	var req CreateFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	// DTO → Service 参数
	p := &service.CreateFriendParams{
		Name:         req.Name,
		Remark:       req.Remark,
		Gender:       req.Gender,
		Phone:        req.Phone,
		Email:        req.Email,
		Location:     req.Location,
		Occupation:   req.Occupation,
		Bio:          req.Bio,
		MeetScene:    req.MeetScene,
		Favorability: req.Favorability,
		MBTI:         req.MBTI,
		Accounts:     req.Accounts,
		GroupIDs:     req.GroupIDs,
		TagIDs:       req.TagIDs,
		UserID:       userID,
	}
	// 生日字符串转 *time.Time
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			p.Birthday = &t
		}
	}

	friend, err := h.friendSvc.Create(c, p)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, friend, "创建成功")
}

// GetByID 查询好友详情
// @Summary 查询好友详情
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Success 200 {object} response.Response{data=model.Friend}
// @Router /api/v1/friends/{id} [get]
func (h *FriendHandler) GetByID(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	friend, err := h.friendSvc.GetByID(c, userID, uint(friendID))
	if err != nil {
		response.Fail(c, errcode.FriendNotFound)
		return
	}
	response.OK(c, friend)
}

// List 好友列表（分页 + 多条件筛选）
// @Summary 好友列表
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(10)
// @Param keyword query string false "关键词（姓名/备注/手机号）"
// @Param group_id query int false "分组ID"
// @Param tag_id query int false "标签ID"
// @Param temperature query int false "关系温度 1夯 2热 3温 4凉 5拉"
// @Param min_favor query int false "最低好感度"
// @Param max_favor query int false "最高好感度"
// @Param sort query string false "排序" Enums(favorability_desc,favorability_asc,last_interact_asc,last_interact_desc)
// @Success 200 {object} response.Response{data=service.FriendListResult}
// @Router /api/v1/friends [get]
func (h *FriendHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 分页参数（默认 1 / 10）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	q := &repository.FriendQuery{
		Keyword:     c.Query("keyword"),
		Page:        page,
		PageSize:    pageSize,
		Sort:        c.Query("sort"),
		Temperature: queryInt(c, "temperature"),
		MinFavor:    queryInt(c, "min_favor"),
		MaxFavor:    queryInt(c, "max_favor"),
	}
	if v := queryInt(c, "group_id"); v > 0 {
		q.GroupID = uint(v)
	}
	if v := queryInt(c, "tag_id"); v > 0 {
		q.TagID = uint(v)
	}

	result, err := h.friendSvc.List(c, userID, q)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, result)
}

// Search 搜索好友
// @Summary 搜索好友
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param keyword query string true "关键词"
// @Success 200 {object} response.Response{data=service.FriendListResult}
// @Router /api/v1/friends/search [get]
func (h *FriendHandler) Search(c *gin.Context) {
	userID := c.GetUint("user_id")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.friendSvc.Search(c, userID, keyword, page, pageSize)
	if err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	response.OK(c, result)
}

// UpdateFriendReq 更新好友请求（指针类型区分"不更新"与"更新为零值"）
type UpdateFriendReq struct {
	Name       *string `json:"name" binding:"omitempty,min=1,max=64"`
	Remark     *string `json:"remark" binding:"omitempty,max=64"`
	Phone      *string `json:"phone"`
	Email      *string `json:"email"`
	Location   *string `json:"location"`
	Occupation *string `json:"occupation"`
	Bio        *string `json:"bio"`
	MBTI       *string `json:"mbti" binding:"omitempty,len=4"`
	GroupIDs   []uint  `json:"group_ids"`
	TagIDs     []uint  `json:"tag_ids"`
}

// Update 更新好友
// @Summary 更新好友
// @Tags 好友管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Param body body UpdateFriendReq true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/v1/friends/{id} [put]
func (h *FriendHandler) Update(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	var req UpdateFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	p := &service.UpdateFriendParams{
		Name:       req.Name,
		Remark:     req.Remark,
		Phone:      req.Phone,
		Email:      req.Email,
		Location:   req.Location,
		Occupation: req.Occupation,
		Bio:        req.Bio,
		MBTI:       req.MBTI,
		GroupIDs:   req.GroupIDs,
		TagIDs:     req.TagIDs,
	}
	if err := h.friendSvc.Update(c, userID, uint(friendID), p); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "更新成功")
}

// Delete 删除好友
// @Summary 删除好友
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Success 200 {object} response.Response
// @Router /api/v1/friends/{id} [delete]
func (h *FriendHandler) Delete(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	if err := h.friendSvc.Delete(c, userID, uint(friendID)); err != nil {
		response.Fail(c, errcode.FriendNotFound)
		return
	}
	response.OKWithMsg(c, nil, "删除成功")
}

// AdjustFavorabilityReq 调整好感度请求
type AdjustFavorabilityReq struct {
	Favorability int `json:"favorability" binding:"required,gte=0,lte=100"` // 目标好感度
}

// AdjustFavorability 调整好感度（联动关系温度）
// @Summary 调整好感度
// @Tags 好友管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Param body body AdjustFavorabilityReq true "好感度"
// @Success 200 {object} response.Response
// @Router /api/v1/friends/{id}/favorability [put]
func (h *FriendHandler) AdjustFavorability(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	var req AdjustFavorabilityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	if err := h.friendSvc.AdjustFavorability(c, userID, uint(friendID), req.Favorability); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}

	// 回读好友拿最新温度，给前端展示联动结果
	friend, err := h.friendSvc.GetByID(c, userID, uint(friendID))
	if err != nil {
		response.OKWithMsg(c, nil, "好感度已调整")
		return
	}
	response.OK(c, gin.H{
		"favorability": friend.Favorability,
		"temperature":  friend.Temperature,
		"temp_label":   model.Temperature(friend.Temperature).Label(),
	})
}

// ListCold 关系变凉好友列表（最久未互动优先）
// @Summary 关系变凉好友列表
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(10)
// @Success 200 {object} response.Response
// @Router /api/v1/friends/cold [get]
func (h *FriendHandler) ListCold(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	friends, total, err := h.friendSvc.ListColdFriends(c, userID, page, pageSize)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"list":      friends,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddAccountReq 添加平台账号请求
type AddAccountReq struct {
	Platform string `json:"platform" binding:"required"` // 平台
	Account  string `json:"account" binding:"required"`  // 账号
	Remark   string `json:"remark"`                      // 备注
}

// AddAccount 添加平台账号
// @Summary 添加平台账号
// @Tags 好友管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Param body body AddAccountReq true "账号信息"
// @Success 200 {object} response.Response
// @Router /api/v1/friends/{id}/accounts [post]
func (h *FriendHandler) AddAccount(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	var req AddAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")

	if err := h.friendSvc.AddAccount(c, userID, uint(friendID), req.Platform, req.Account, req.Remark); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "添加成功")
}

// ListAccounts 获取好友平台账号
// @Summary 获取好友平台账号
// @Tags 好友管理
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param id path int true "好友ID"
// @Success 200 {object} response.Response
// @Router /api/v1/friends/{id}/accounts [get]
func (h *FriendHandler) ListAccounts(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}
	userID := c.GetUint("user_id")

	accounts, err := h.friendSvc.ListAccounts(c, userID, uint(friendID))
	if err != nil {
		response.Fail(c, errcode.FriendNotFound)
		return
	}
	response.OK(c, gin.H{"list": accounts, "count": len(accounts)})
}

// ExportCSV 导出好友 CSV
// @Summary 导出好友 CSV
// @Tags 好友管理
// @Produce text/csv
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {file} file
// @Router /api/v1/friends/export [get]
func (h *FriendHandler) ExportCSV(c *gin.Context) {
	userID := c.GetUint("user_id")

	data, err := h.friendSvc.ExportFriendsCSV(c, userID)
	if err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	// 以文件下载形式返回（带 BOM 的 CSV，Excel 可直接打开）
	c.Header("Content-Disposition", "attachment; filename=friends.csv")
	c.Data(200, "text/csv; charset=utf-8", data)
}

// ImportCSV 导入好友 CSV
// @Summary 导入好友 CSV
// @Tags 好友管理
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param file formData file true "CSV 文件"
// @Success 200 {object} response.Response{data=service.ImportResult}
// @Router /api/v1/friends/import [post]
func (h *FriendHandler) ImportCSV(c *gin.Context) {
	userID := c.GetUint("user_id")

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, "请上传 CSV 文件")
		return
	}
	defer file.Close()

	// 读取全部内容（教学项目单文件体积小，直接 ReadAll）
	data, err := io.ReadAll(file)
	if err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, "读取文件失败")
		return
	}

	result, err := h.friendSvc.ImportFriendsCSV(c, userID, data)
	if err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	response.OKWithMsg(c, result, "导入完成")
}

// queryInt 从 query 中读取 int 参数（非法值返回 0）
func queryInt(c *gin.Context, key string) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return 0
	}
	return v
}
