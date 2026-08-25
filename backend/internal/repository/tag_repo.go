package repository

import (
	"context"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// TagRepo 标签数据访问对象
type TagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签 Repository
func NewTagRepo(db *gorm.DB) *TagRepo {
	return &TagRepo{db: db}
}

// Create 创建标签
func (r *TagRepo) Create(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// Delete 删除标签（事务内先清理中间表关联，再删除标签本身）
// 注意：friend_tags.tag_id 外键约束指向 tags.id，
// 若不先清理关联直接删除标签，会触发外键约束错误
// （Error 1451: Cannot delete or update a parent row）。
func (r *TagRepo) Delete(ctx context.Context, userID, tagID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先清理好友-标签中间表关联
		if err := tx.Exec("DELETE FROM friend_tags WHERE tag_id = ?", tagID).Error; err != nil {
			return err
		}
		// 2. 再删除标签本身
		result := tx.Where("id = ? AND user_id = ?", tagID, userID).
			Delete(&model.Tag{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// List 查询用户全部标签
func (r *TagRepo) List(ctx context.Context, userID uint) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tags).Error
	return tags, err
}
