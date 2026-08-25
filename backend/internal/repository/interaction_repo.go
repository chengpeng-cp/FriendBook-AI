package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// InteractionRepo 互动记录数据访问对象
type InteractionRepo struct {
	db *gorm.DB
}

// NewInteractionRepo 创建互动记录 Repository
func NewInteractionRepo(db *gorm.DB) *InteractionRepo {
	return &InteractionRepo{db: db}
}

// DB 暴露底层 *gorm.DB（统计服务做日期聚合查询用）
func (r *InteractionRepo) DB() *gorm.DB {
	return r.db
}

// Create 创建互动记录
func (r *InteractionRepo) Create(ctx context.Context, interaction *model.Interaction) error {
	return r.db.WithContext(ctx).Create(interaction).Error
}

// Update 按字段更新互动记录
func (r *InteractionRepo) Update(ctx context.Context, userID, interactionID uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Interaction{}).
		Where("id = ? AND user_id = ?", interactionID, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除互动记录
func (r *InteractionRepo) Delete(ctx context.Context, userID, interactionID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", interactionID, userID).
		Delete(&model.Interaction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 分页查询互动记录（可按好友筛选，按互动时间倒序）
func (r *InteractionRepo) List(ctx context.Context, userID, friendID uint, page, pageSize int) ([]model.Interaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Interaction{}).
		Where("user_id = ?", userID)
	if friendID > 0 {
		query = query.Where("friend_id = ?", friendID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Interaction
	err := query.Order("interact_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListByFriendID 查询好友全部互动记录（AI 分析"共同记忆"用，不分页）
func (r *InteractionRepo) ListByFriendID(ctx context.Context, userID, friendID uint) ([]model.Interaction, error) {
	var list []model.Interaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Order("interact_at DESC").
		Find(&list).Error
	return list, err
}
