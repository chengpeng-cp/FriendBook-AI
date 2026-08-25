package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// TagService 标签业务逻辑
type TagService struct {
	tagRepo *repository.TagRepo
}

// NewTagService 创建标签 Service
func NewTagService(tagRepo *repository.TagRepo) *TagService {
	return &TagService{tagRepo: tagRepo}
}

// Create 创建标签（名称必填，颜色可选默认蓝色）
func (s *TagService) Create(ctx context.Context, userID uint, name, color string) (*model.Tag, error) {
	if name == "" {
		return nil, errors.New("标签名称不能为空")
	}
	if color == "" {
		color = "#409EFF"
	}
	tag := &model.Tag{UserID: userID, Name: name, Color: color}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// Delete 删除标签
func (s *TagService) Delete(ctx context.Context, userID, tagID uint) error {
	err := s.tagRepo.Delete(ctx, userID, tagID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("标签不存在")
	}
	return err
}

// List 查询用户全部标签
func (s *TagService) List(ctx context.Context, userID uint) ([]model.Tag, error) {
	return s.tagRepo.List(ctx, userID)
}
