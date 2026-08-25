package model

import "time"

// User 用户表模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64;uniqueIndex:uk_username;not null" json:"username"` // 用户名（唯一）
	Password  string    `gorm:"size:128;not null" json:"-"`                              // 密码（bcrypt 哈希，序列化时隐藏）
	Nickname  string    `gorm:"size:64;default:''" json:"nickname"`                       // 昵称
	Avatar    string    `gorm:"size:255;default:''" json:"avatar"`                        // 头像 URL
	Bio       string    `gorm:"size:255;default:''" json:"bio"`                            // 个人简介
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 显式指定表名
func (User) TableName() string { return "users" }
