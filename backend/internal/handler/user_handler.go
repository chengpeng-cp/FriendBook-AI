// Package handler HTTP 处理层（Controller）：
// 职责单一——参数绑定与校验 → 调用 Service → 组装统一响应。
// 不写业务逻辑，业务逻辑全部在 Service 层。
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"golang-learn/internal/pkg/errcode"
	"golang-learn/internal/pkg/response"
	"golang-learn/internal/service"
)

// UserHandler 用户相关接口
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户 Handler
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// RegisterReq 注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`  // 用户名
	Password string `json:"password" binding:"required,min=6,max=128"` // 密码
	Nickname string `json:"nickname" binding:"max=64"`                 // 昵称（可选）
}

// Register 用户注册
// @Summary 用户注册
// @Tags 用户
// @Accept json
// @Produce json
// @Param body body RegisterReq true "注册信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	user, err := h.userSvc.Register(c, &service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		if err.Error() == "用户名已存在" {
			response.Fail(c, errcode.UserAlreadyExist)
			return
		}
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}

	response.OKWithMsg(c, gin.H{"id": user.ID, "username": user.Username}, "注册成功")
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 用户登录
// @Summary 用户登录
// @Tags 用户
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录信息"
// @Success 200 {object} response.Response{data=service.LoginOutput}
// @Router /api/v1/users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	out, err := h.userSvc.Login(c, req.Username, req.Password)
	if err != nil {
		if err.Error() == "用户不存在" {
			response.Fail(c, errcode.UserNotFound)
			return
		}
		if err.Error() == "用户名或密码错误" {
			response.Fail(c, errcode.PasswordIncorrect)
			return
		}
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, out, "登录成功")
}

// GetProfile 获取个人信息
// @Summary 获取个人信息
// @Tags 用户
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.userSvc.GetProfile(c, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, errcode.UserNotFound)
			return
		}
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OK(c, user)
}

// UpdateProfileReq 修改个人信息请求
type UpdateProfileReq struct {
	Nickname string `json:"nickname" binding:"max=64"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Bio      string `json:"bio" binding:"max=255"`
}

// UpdateProfile 修改个人信息
// @Summary 修改个人信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body UpdateProfileReq true "个人信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")
	if err := h.userSvc.UpdateProfile(c, userID, req.Nickname, req.Avatar, req.Bio); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "更新成功")
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=128"` // 新密码
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 用户
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <token>"
// @Param body body ChangePasswordReq true "密码信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, errcode.InvalidParams, err.Error())
		return
	}
	userID := c.GetUint("user_id")
	if err := h.userSvc.ChangePassword(c, userID, req.OldPassword, req.NewPassword); err != nil {
		response.FailWithMsg(c, errcode.InternalError, err.Error())
		return
	}
	response.OKWithMsg(c, nil, "密码修改成功")
}
