package model

import "time"

// Reminder 提醒表（生日/纪念日提醒，由定时任务自动生成或用户手动创建）
type Reminder struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	FriendID      uint         `gorm:"index;not null" json:"friend_id"`        // 关联好友
	UserID        uint         `gorm:"index:idx_user_date;not null" json:"user_id"` // 归属用户
	Type          ReminderType `gorm:"type:tinyint;not null" json:"type"`     // 1生日 2纪念日
	Title         string       `gorm:"size:64;not null" json:"title"`         // 提醒标题
	Date          time.Time    `gorm:"type:date;not null" json:"date"`        // 提醒日期
	DaysInAdvance int          `gorm:"default:3" json:"days_in_advance"`      // 提前几天开始提醒
	CreatedAt     time.Time    `json:"created_at"`
}

// TableName 显式指定表名
func (Reminder) TableName() string { return "reminders" }
