package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// ReminderRepo 提醒数据访问对象
type ReminderRepo struct {
	db *gorm.DB
}

// NewReminderRepo 创建提醒 Repository
func NewReminderRepo(db *gorm.DB) *ReminderRepo {
	return &ReminderRepo{db: db}
}

// Upsert 按好友+类型去重的写入：
// 不存在则创建；已存在则更新日期/标题/提前天数（生日可能被用户改过，需同步刷新）
func (r *ReminderRepo) Upsert(ctx context.Context, reminder *model.Reminder) error {
	var existing model.Reminder
	err := r.db.WithContext(ctx).
		Where("friend_id = ? AND type = ?", reminder.FriendID, reminder.Type).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(reminder).Error
	}
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"date":            reminder.Date,
		"days_in_advance": reminder.DaysInAdvance,
		"title":           reminder.Title,
		"user_id":         reminder.UserID,
	}).Error
}

// GetByDateRange 查询用户在 [start, end] 日期范围内的提醒（含当天）
func (r *ReminderRepo) GetByDateRange(ctx context.Context, userID uint, start, end time.Time) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?",
			userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date ASC").
		Find(&reminders).Error
	return reminders, err
}
