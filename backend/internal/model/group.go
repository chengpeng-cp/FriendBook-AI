package model

import "time"

// Group 好友分组表（如"大学同学""同事""家人"）
type Group struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_id;not null" json:"user_id"` // 所属用户
	Name      string    `gorm:"size:64;not null" json:"name"`              // 分组名称
	SortOrder int       `gorm:"default:0" json:"sort_order"`               // 排序值（越小越靠前）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 显式指定表名
func (Group) TableName() string { return "groups" }
