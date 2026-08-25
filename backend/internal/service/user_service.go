// Package service 业务逻辑层：封装核心业务规则、事务编排。
// HTTP Handler 与 AI Tools 都调用本层，保证业务逻辑只有一份实现。
package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"golang-learn/internal/config"
	"golang-learn/internal/model"
	"golang-learn/internal/pkg/utils"
	"golang-learn/internal/repository"
)

// UserService 用户业务逻辑
type UserService struct {
	userRepo *repository.UserRepo
	jwtCfg   *config.JWTConfig
}

// NewUserService 创建用户 Service
func NewUserService(userRepo *repository.UserRepo, jwtCfg *config.JWTConfig) *UserService {
	return &UserService{userRepo: userRepo, jwtCfg: jwtCfg}
}

// RegisterInput 注册入参
type RegisterInput struct {
	Username string // 用户名（3-64 字符）
	Password string // 明文密码（6-128 字符，Service 层做兜底校验）
	Nickname string // 昵称（可选）
}

// Register 注册：用户名查重 → bcrypt 加密 → 落库
func (s *UserService) Register(ctx context.Context, in *RegisterInput) (*model.User, error) {
	if len(in.Username) < 3 || len(in.Password) < 6 {
		return nil, errors.New("用户名至少3位，密码至少6位")
	}

	// 1. 用户名查重
	if _, err := s.userRepo.GetByUsername(ctx, in.Username); err == nil {
		return nil, errors.New("用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. 密码加密（绝不存明文）
	hashed, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	// 3. 落库
	user := &model.User{
		Username: in.Username,
		Password: hashed,
		Nickname: in.Nickname,
	}
	if user.Nickname == "" {
		user.Nickname = in.Username
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// LoginOutput 登录返回
type LoginOutput struct {
	Token     string `json:"token"`      // JWT
	ExpiresAt int64  `json:"expires_at"` // 过期时间戳（秒）
}

// Login 登录：查询用户 → 校验密码 → 签发 Token
func (s *UserService) Login(ctx context.Context, username, password string) (*LoginOutput, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// bcrypt 校验（密码错误统一提示，避免向攻击者泄露"用户存在"）
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, expiresAt, err := utils.GenerateToken(
		s.jwtCfg.Secret, s.jwtCfg.Issuer, user.ID, user.Username, s.jwtCfg.ExpireHours)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{Token: token, ExpiresAt: expiresAt}, nil
}

// GetProfile 查询个人信息
func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile 修改个人信息（昵称/头像/简介）
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, nickname, avatar, bio string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if bio != "" {
		updates["bio"] = bio
	}
	if len(updates) == 0 {
		return nil // 无可更新字段，直接成功
	}
	return s.userRepo.Update(ctx, userID, updates)
}

// ChangePassword 修改密码：校验旧密码 → 加密新密码 → 更新
func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("新密码至少6位")
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("旧密码错误")
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.userRepo.Update(ctx, userID, map[string]interface{}{"password": hashed})
}
