package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// GroupService 分组业务逻辑
type GroupService struct {
	groupRepo *repository.GroupRepo
}

// NewGroupService 创建分组 Service
func NewGroupService(groupRepo *repository.GroupRepo) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

// Create 创建分组（名称必填）
func (s *GroupService) Create(ctx context.Context, userID uint, name string, sortOrder int) (*model.Group, error) {
	if name == "" {
		return nil, errors.New("分组名称不能为空")
	}
	group := &model.Group{UserID: userID, Name: name, SortOrder: sortOrder}
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// Update 更新分组
func (s *GroupService) Update(ctx context.Context, userID, groupID uint, name string, sortOrder int) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	updates["sort_order"] = sortOrder
	err := s.groupRepo.Update(ctx, userID, groupID, updates)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("分组不存在")
	}
	return err
}

// Delete 删除分组
func (s *GroupService) Delete(ctx context.Context, userID, groupID uint) error {
	err := s.groupRepo.Delete(ctx, userID, groupID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("分组不存在")
	}
	return err
}

// List 查询用户全部分组
func (s *GroupService) List(ctx context.Context, userID uint) ([]model.Group, error) {
	return s.groupRepo.List(ctx, userID)
}
