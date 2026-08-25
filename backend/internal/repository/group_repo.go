package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// GroupRepo 分组数据访问对象
type GroupRepo struct {
	db *gorm.DB
}

// NewGroupRepo 创建分组 Repository
func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

// Create 创建分组
func (r *GroupRepo) Create(ctx context.Context, group *model.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// Update 更新分组（名称/排序值）
func (r *GroupRepo) Update(ctx context.Context, userID, groupID uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Group{}).
		Where("id = ? AND user_id = ?", groupID, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除分组（事务内先清理中间表关联，再删除分组本身）
// 注意：friend_groups.group_id 外键约束指向 groups.id，
// 若不先清理关联直接删除分组，会触发外键约束错误
// （Error 1451: Cannot delete or update a parent row）。
func (r *GroupRepo) Delete(ctx context.Context, userID, groupID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先清理好友-分组中间表关联
		if err := tx.Exec("DELETE FROM friend_groups WHERE group_id = ?", groupID).Error; err != nil {
			return err
		}
		// 2. 再删除分组本身
		result := tx.Where("id = ? AND user_id = ?", groupID, userID).
			Delete(&model.Group{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// List 查询用户全部分组（按 sort_order 升序）
func (r *GroupRepo) List(ctx context.Context, userID uint) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("sort_order ASC, id ASC").
		Find(&groups).Error
	return groups, err
}
