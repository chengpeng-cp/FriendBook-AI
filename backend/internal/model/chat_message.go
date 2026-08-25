package model

import "time"

// ChatMessage AI 对话记录表（持久化多轮对话历史，刷新页面后可恢复）
type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`         // 归属用户（数据隔离）
	Role      string    `gorm:"size:16;not null" json:"role"`          // user / assistant
	Content   string    `gorm:"type:text;not null" json:"content"`     // 消息内容
	CreatedAt time.Time `json:"created_at"`                            // 创建时间（按序展示）
}

// TableName 显式指定表名
func (ChatMessage) TableName() string { return "chat_messages" }
