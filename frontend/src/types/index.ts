/**
 * 全局类型定义 —— 与后端 model/handler 层 JSON 契约一一对应。
 * 后端统一响应：{ code, message, data }（HTTP 恒 200，code=0 成功）
 */

/** 统一响应结构 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

/** 分页列表结构（后端 FriendListResult / gin.H 两种返回均为此形态） */
export interface Paged<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

/** 简单列表结构（list + count） */
export interface ListWrap<T> {
  list: T[]
  count: number
}

// ==================== 用户 ====================

export interface User {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  created_at: string
  updated_at: string
}

export interface LoginOutput {
  token: string
  expires_at: number // 秒级时间戳
}

// ==================== 好友 ====================

/** 关系温度枚举：1夯 2热 3温 4凉 5拉（值越小关系越热） */
export type Temperature = 1 | 2 | 3 | 4 | 5

/** 性别：0未知 1男 2女 */
export type Gender = 0 | 1 | 2

export interface Friend {
  id: number
  user_id: number
  name: string
  remark: string
  avatar: string
  gender: Gender
  birthday: string | null
  phone: string
  email: string
  location: string
  occupation: string
  bio: string
  meet_scene: string
  favorability: number // 0-100
  temperature: Temperature
  mbti: string
  last_interact: string | null
  accounts?: Account[] // 详情/创建时返回
  groups?: Group[]
  tags?: Tag[]
  created_at: string
  updated_at: string
}

/** 平台账号 */
export interface Account {
  id: number
  friend_id: number
  platform: string // wechat/qq/douyin/xhs/bilibili/phone/email/custom
  account: string
  remark: string
  created_at: string
}

/** 创建好友入参（accounts 随主记录一并创建） */
export interface AccountParam {
  platform: string
  account: string
  remark?: string
}

export interface CreateFriendPayload {
  name: string
  remark?: string
  gender?: Gender
  birthday?: string // yyyy-MM-dd
  phone?: string
  email?: string
  location?: string
  occupation?: string
  bio?: string
  meet_scene?: string
  favorability?: number
  mbti?: string
  accounts?: AccountParam[]
  group_ids?: number[]
  tag_ids?: number[]
}

/** 更新好友入参（指针语义：不传的字段不更新） */
export interface UpdateFriendPayload {
  name?: string
  remark?: string
  phone?: string
  email?: string
  location?: string
  occupation?: string
  bio?: string
  mbti?: string
  group_ids?: number[]
  tag_ids?: number[]
}

/** 好友列表查询参数 */
export interface FriendQuery {
  page?: number
  page_size?: number
  keyword?: string
  group_id?: number
  tag_id?: number
  temperature?: number
  min_favor?: number
  max_favor?: number
  sort?: 'favorability_desc' | 'favorability_asc' | 'last_interact_asc' | 'last_interact_desc'
}

/** 调整好感度返回（联动温度结果） */
export interface AdjustFavorResult {
  favorability: number
  temperature: Temperature
  temp_label: string
}

/** CSV 导入结果 */
export interface ImportResult {
  success_count: number
  failed_rows: string[]
}

// ==================== 分组 / 标签 ====================

export interface Group {
  id: number
  user_id: number
  name: string
  sort_order: number
  created_at: string
}

export interface Tag {
  id: number
  user_id: number
  name: string
  color: string
  created_at: string
}

// ==================== 互动记录 ====================

export type InteractionType = 1 | 2 | 3 | 4 | 5 | 6 // 聊天 见面 通话 聚餐 送礼 其他

export interface Interaction {
  id: number
  friend_id: number
  user_id: number
  type: InteractionType
  content: string
  location: string
  mood: string
  photo_urls: string // JSON 数组字符串
  is_memory: boolean
  memory_title: string
  interact_at: string
  created_at: string
}

export interface InteractionPayload {
  friend_id: number
  type: InteractionType
  content?: string
  location?: string
  mood?: string
  photo_urls?: string[]
  is_memory?: boolean
  memory_title?: string
  interact_at?: string // yyyy-MM-dd HH:mm:ss（缺省为当前）
}

// ==================== 提醒 ====================

/** 提醒类型：1生日 2纪念日 */
export type ReminderType = 1 | 2

export interface Reminder {
  id: number
  friend_id: number
  user_id: number
  type: ReminderType
  title: string
  date: string
  days_in_advance: number
  created_at: string
}

// ==================== 统计 ====================

export interface StatsOverview {
  total: number
  new_this_month: number
  today_reminders: number
  cold_count: number
  avg_favor: number
  temp_dist: Record<string, number> // key 为 "1"~"5"
  platform_dist: Record<string, number>
  interact_trend: { date: string; count: number }[]
}

export interface TemperatureDist {
  hot: number
  warm: number
  mild: number
  cool: number
  cold: number
}

export interface PlatformDistItem {
  platform: string
  count: number
}

export interface TrendPoint {
  date: string
  count: number
}

// ==================== AI 对话 ====================

export type ChatRole = 'system' | 'user' | 'assistant' | 'tool'

export interface ChatMessage {
  role: ChatRole
  content: string
}

/** SSE 事件帧：{"content":"文本","tool_call":"工具名","error":"错误信息","done":bool} */
export interface AiStreamEvent {
  content: string
  tool_call?: string
  tool_args?: string
  /** 流中断时后端透传的错误信息 */
  error?: string
  done: boolean
}

/** 持久化的对话记录（后端 chat_messages 表） */
export interface ChatHistoryItem extends ChatMessage {
  id: number
  user_id: number
  created_at: string
}
