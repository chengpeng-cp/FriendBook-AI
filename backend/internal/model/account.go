package model

import "time"

// Account 社交平台账号表（一个好友可绑定多个平台账号：微信/QQ/抖音/小红书/B站等）
type Account struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FriendID  uint      `gorm:"index:idx_friend_id;not null" json:"friend_id"` // 所属好友
	Platform  string    `gorm:"size:32;index:idx_platform;not null" json:"platform"` // 平台标识：wechat/qq/douyin/xhs/bilibili/phone/email/custom
	Account   string    `gorm:"size:128;not null" json:"account"`            // 账号/号码
	Remark    string    `gorm:"size:64;default:''" json:"remark"`            // 备注
	CreatedAt time.Time `json:"created_at"`
}

// TableName 显式指定表名
func (Account) TableName() string { return "accounts" }

// PlatformLabel 平台中文名（供展示与 AI 理解）
func (a *Account) PlatformLabel() string {
	switch a.Platform {
	case "wechat":
		return "微信"
	case "qq":
		return "QQ"
	case "douyin":
		return "抖音"
	case "xhs":
		return "小红书"
	case "bilibili":
		return "B站"
	case "phone":
		return "手机"
	case "email":
		return "邮箱"
	default:
		return a.Platform
	}
}
