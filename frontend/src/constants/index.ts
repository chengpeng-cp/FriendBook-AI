/**
 * 业务常量与后端 model/enums.go 保持一致，
 * 前端所有枚举展示统一从这里取，禁止散落魔法数字。
 */
import type { Gender, InteractionType, ReminderType, Temperature } from '@/types'

/** 关系温度元信息（核心视觉语言） */
export const TEMP_META: Record<Temperature, { label: string; emoji: string; color: string; desc: string }> = {
  1: { label: '夯', emoji: '🔥', color: '#FF5C39', desc: '7 天内有互动 / 好感度 ≥ 90' },
  2: { label: '热', emoji: '😊', color: '#FFA940', desc: '30 天内有互动 / 好感度 75-89' },
  3: { label: '温', emoji: '🙂', color: '#F6C453', desc: '90 天内有互动 / 好感度 60-74' },
  4: { label: '凉', emoji: '😐', color: '#63B3ED', desc: '180 天内有互动 / 好感度 40-59' },
  5: { label: '拉', emoji: '🥶', color: '#8A94A3', desc: '超 180 天未互动 / 好感度 < 40' },
}

/** 好感度 → 温度档位（与后端 Service 联动规则一致，用于表单实时预览） */
export function favorToTemp(favor: number): Temperature {
  if (favor >= 90) return 1
  if (favor >= 75) return 2
  if (favor >= 60) return 3
  if (favor >= 40) return 4
  return 5
}

/** 性别元信息 */
export const GENDER_META: Record<Gender, string> = { 0: '未知', 1: '男', 2: '女' }

/** 互动类型元信息 */
export const INTERACT_META: Record<InteractionType, { label: string; emoji: string; color: string }> = {
  1: { label: '聊天', emoji: '💬', color: '#63B3ED' },
  2: { label: '见面', emoji: '🤝', color: '#FFA940' },
  3: { label: '通话', emoji: '📞', color: '#9F7AEA' },
  4: { label: '聚餐', emoji: '🍜', color: '#F6C453' },
  5: { label: '送礼', emoji: '🎁', color: '#F687B3' },
  6: { label: '其他', emoji: '✨', color: '#8A94A3' },
}

/** 提醒类型元信息 */
export const REMINDER_META: Record<ReminderType, { label: string; emoji: string }> = {
  1: { label: '生日', emoji: '🎂' },
  2: { label: '纪念日', emoji: '💌' },
}

/** 平台元信息（与后端 Account.PlatformLabel 一致） */
export const PLATFORM_META: Record<string, { label: string; emoji: string }> = {
  wechat: { label: '微信', emoji: '💚' },
  qq: { label: 'QQ', emoji: '🐧' },
  douyin: { label: '抖音', emoji: '🎵' },
  xhs: { label: '小红书', emoji: '📕' },
  bilibili: { label: 'B站', emoji: '📺' },
  phone: { label: '手机', emoji: '📱' },
  email: { label: '邮箱', emoji: '📮' },
  custom: { label: '自定义', emoji: '🔗' },
}

/** 根据用户 ID 生成稳定的头像渐变色（同一好友颜色恒定） */
const AVATAR_GRADIENTS = [
  'from-orange-400 to-rose-500',
  'from-amber-400 to-orange-600',
  'from-sky-400 to-indigo-500',
  'from-emerald-400 to-teal-600',
  'from-fuchsia-400 to-purple-600',
  'from-cyan-400 to-blue-500',
]
export function avatarGradient(id: number): string {
  return AVATAR_GRADIENTS[id % AVATAR_GRADIENTS.length]
}

/** 姓名首字符（头像占位） */
export function nameInitial(name: string): string {
  return (name.trim()[0] ?? '友').toUpperCase()
}
