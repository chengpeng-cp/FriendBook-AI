package model

// ==================== 枚举常量定义 ====================
// 集中定义业务枚举，避免代码里出现魔法数字。

// Gender 性别
type Gender int8

const (
	GenderUnknown Gender = 0 // 未知
	GenderMale    Gender = 1 // 男
	GenderFemale  Gender = 2 // 女
)

// String 性别中文文案
func (g Gender) String() string {
	switch g {
	case GenderMale:
		return "男"
	case GenderFemale:
		return "女"
	default:
		return "未知"
	}
}

// Temperature 关系温度（趣味档位：从"夯"到"拉"）
// 值越小关系越热，值越大关系越冷。
type Temperature int8

const (
	TempHot  Temperature = 1 // 夯：7 天内有互动 / 好感度 >= 90
	TempWarm Temperature = 2 // 热：30 天内有互动 / 好感度 75-89
	TempMild Temperature = 3 // 温：90 天内有互动 / 好感度 60-74
	TempCool Temperature = 4 // 凉：180 天内有互动 / 好感度 40-59
	TempCold Temperature = 5 // 拉：超过 180 天未互动 / 好感度 < 40
)

// String 返回温度的纯中文文案
func (t Temperature) String() string {
	switch t {
	case TempHot:
		return "夯"
	case TempWarm:
		return "热"
	case TempMild:
		return "温"
	case TempCool:
		return "凉"
	case TempCold:
		return "拉"
	default:
		return "未知"
	}
}

// Label 返回温度的展示标签（带 emoji，供前端/AI 输出使用）
func (t Temperature) Label() string {
	switch t {
	case TempHot:
		return "🔥 夯"
	case TempWarm:
		return "😊 热"
	case TempMild:
		return "🙂 温"
	case TempCool:
		return "😐 凉"
	case TempCold:
		return "🥶 拉"
	default:
		return "未知"
	}
}

// InteractionType 互动类型
type InteractionType int8

const (
	InteractChat  InteractionType = 1 // 聊天
	InteractMeet  InteractionType = 2 // 见面
	InteractCall  InteractionType = 3 // 通话
	InteractDine  InteractionType = 4 // 聚餐
	InteractGift  InteractionType = 5 // 送礼
	InteractOther InteractionType = 6 // 其他
)

// String 互动类型中文文案
func (i InteractionType) String() string {
	switch i {
	case InteractChat:
		return "聊天"
	case InteractMeet:
		return "见面"
	case InteractCall:
		return "通话"
	case InteractDine:
		return "聚餐"
	case InteractGift:
		return "送礼"
	case InteractOther:
		return "其他"
	default:
		return "未知"
	}
}

// ReminderType 提醒类型
type ReminderType int8

const (
	ReminderBirthday   ReminderType = 1 // 生日
	ReminderAnniversary ReminderType = 2 // 纪念日
)

// String 提醒类型中文文案
func (r ReminderType) String() string {
	switch r {
	case ReminderBirthday:
		return "生日"
	case ReminderAnniversary:
		return "纪念日"
	default:
		return "自定义"
	}
}
