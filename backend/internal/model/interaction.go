package model

import "time"

// Interaction 互动记录表（与好友的每次聊天/见面/通话/聚餐/送礼记录）
type Interaction struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	FriendID    uint            `gorm:"index:idx_friend_id;not null" json:"friend_id"` // 好友 ID
	UserID      uint            `gorm:"index:idx_user_interact;not null" json:"user_id"` // 用户 ID
	Type        InteractionType `gorm:"type:tinyint;default:1" json:"type"`           // 1聊天 2见面 3通话 4聚餐 5送礼 6其他
	Content     string          `gorm:"type:text" json:"content"`                     // 互动内容摘要
	Location    string          `gorm:"size:128;default:''" json:"location"`          // 地点
	Mood        string          `gorm:"size:32;default:''" json:"mood"`               // 心情
	PhotoURLs   string          `gorm:"type:json" json:"photo_urls"`                  // 照片 URL 列表（JSON 数组字符串）
	IsMemory    bool            `gorm:"default:false" json:"is_memory"`              // 是否标记为"记忆卡片"
	MemoryTitle string          `gorm:"size:64;default:''" json:"memory_title"`       // 记忆卡片标题
	InteractAt  time.Time       `gorm:"not null" json:"interact_at"`                  // 互动发生时间
	CreatedAt   time.Time       `json:"created_at"`
}

// TableName 显式指定表名
func (Interaction) TableName() string { return "interactions" }
