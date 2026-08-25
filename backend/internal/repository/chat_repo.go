package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// ChatRepo AI 对话记录数据访问对象
type ChatRepo struct {
	db *gorm.DB
}

// NewChatRepo 创建对话记录 Repository
func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// CreateBatch 批量写入一轮对话（用户消息 + AI 回复）
func (r *ChatRepo) CreateBatch(ctx context.Context, msgs []*model.ChatMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(msgs).Error
}

// ListRecent 查询最近 limit 条对话（按时间正序返回，便于前端直接渲染）
func (r *ChatRepo) ListRecent(ctx context.Context, userID uint, limit int) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	// 子查询取最新 N 条（倒序），再正序排回时间线
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	// 反转为时间正序
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// Clear 清空用户全部对话记录
func (r *ChatRepo) Clear(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.ChatMessage{}).Error
}
