// Package repository 数据访问层：封装所有 GORM 数据库操作。
// 规范：所有涉及用户数据的查询必须带 user_id 条件，保证数据隔离。
package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// UserRepo 用户数据访问对象
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户 Repository
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetByUsername 按用户名查询（注册查重 / 登录使用）
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 按主键查询
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 按字段更新用户信息（map 形式，只更新传入的字段）
func (r *UserRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// ListAllUserIDs 返回所有用户 ID（定时任务遍历使用）
func (r *UserRepo) ListAllUserIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.User{}).Pluck("id", &ids).Error
	return ids, err
}
