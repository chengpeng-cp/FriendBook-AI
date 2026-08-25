package model

import "time"

// Tag 标签表（好友的多对多标签，如"社牛""夜猫子""美食家"）
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_id;not null" json:"user_id"` // 所属用户
	Name      string    `gorm:"size:32;not null" json:"name"`              // 标签名称
	Color     string    `gorm:"size:16;default:'#409EFF'" json:"color"`    // 标签颜色（前端展示用）
	CreatedAt time.Time `json:"created_at"`
}

// TableName 显式指定表名
func (Tag) TableName() string { return "tags" }
