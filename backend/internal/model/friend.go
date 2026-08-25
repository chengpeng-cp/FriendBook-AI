package model

import (
	"fmt"
	"time"
)

// Friend 好友表模型（核心业务表）
type Friend struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	UserID       uint          `gorm:"index:idx_user_id;not null" json:"user_id"`     // 所属用户（数据隔离关键字段）
	Name         string        `gorm:"size:64;not null" json:"name"`                  // 姓名
	Remark       string        `gorm:"size:64;default:''" json:"remark"`              // 备注名
	Avatar       string        `gorm:"size:255;default:''" json:"avatar"`             // 头像 URL
	Gender       Gender        `gorm:"type:tinyint;default:0" json:"gender"`          // 0未知 1男 2女
	Birthday     *time.Time    `gorm:"type:date" json:"birthday"`                     // 生日（可为空）
	Phone        string        `gorm:"size:20;default:''" json:"phone"`               // 手机号
	Email        string        `gorm:"size:128;default:''" json:"email"`              // 邮箱
	Location     string        `gorm:"size:64;default:''" json:"location"`            // 所在城市
	Occupation   string        `gorm:"size:64;default:''" json:"occupation"`          // 职业
	Bio          string        `gorm:"size:255;default:''" json:"bio"`                // 一句话简介
	MeetScene    string        `gorm:"size:128;default:''" json:"meet_scene"`         // 认识场景
	Favorability int           `gorm:"default:50" json:"favorability"`                // 好感度 0-100
	Temperature  Temperature   `gorm:"type:tinyint;default:2;index:idx_user_temperature" json:"temperature"` // 关系温度 1夯 2热 3温 4凉 5拉
	MBTI         string        `gorm:"size:4;default:''" json:"mbti"`                 // MBTI 人格
	LastInteract *time.Time    `json:"last_interact"`                                 // 最近互动时间（可为空）

	// 关联：多平台账号（一对多）
	Accounts []Account `gorm:"foreignKey:FriendID" json:"accounts,omitempty"`
	// 关联：分组（多对多，中间表 friend_groups）
	Groups []Group `gorm:"many2many:friend_groups;" json:"groups,omitempty"`
	// 关联：标签（多对多，中间表 friend_tags）
	Tags []Tag `gorm:"many2many:friend_tags;" json:"tags,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 显式指定表名
func (Friend) TableName() string { return "friends" }

// ToSummary 生成好友信息摘要（供 AI 工具返回给大模型使用，
// 控制字段数量避免撑爆上下文窗口）
func (f *Friend) ToSummary() string {
	birthday := "未填写"
	if f.Birthday != nil {
		birthday = f.Birthday.Format("2006-01-02")
	}
	last := "从未互动"
	if f.LastInteract != nil {
		last = f.LastInteract.Format("2006-01-02")
	}
	return fmt.Sprintf(
		"ID:%d 姓名:%s 备注:%s 性别:%s 生日:%s 城市:%s 职业:%s 好感度:%d 温度:%s(%d) MBTI:%s 最近互动:%s",
		f.ID, f.Name, f.Remark, f.Gender.String(), birthday,
		f.Location, f.Occupation, f.Favorability,
		f.Temperature.String(), int8(f.Temperature), f.MBTI, last,
	)
}
