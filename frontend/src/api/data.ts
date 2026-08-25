/** 互动记录 / 提醒 / 统计 API */
import { del, get, post, put } from './client'
import type {
  Interaction,
  InteractionPayload,
  InteractionType,
  ListWrap,
  Paged,
  PlatformDistItem,
  Reminder,
  StatsOverview,
  TemperatureDist,
  TrendPoint,
} from '@/types'

export const interactionApi = {
  /** 互动列表（可按好友筛选） */
  list: (friendId?: number, page = 1, pageSize = 10) =>
    get<Paged<Interaction>>('/api/v1/interactions', {
      friend_id: friendId || undefined,
      page,
      page_size: pageSize,
    }),

  /** 创建互动记录（联动刷新好友温度） */
  create: (payload: InteractionPayload) => post<Interaction>('/api/v1/interactions', payload),

  update: (id: number, payload: Partial<InteractionPayload> & { type?: InteractionType }) =>
    put<null>(`/api/v1/interactions/${id}`, payload),

  remove: (id: number) => del<null>(`/api/v1/interactions/${id}`),
}

export const reminderApi = {
  /** 今日提醒（含未来 3 天） */
  today: () => get<ListWrap<Reminder>>('/api/v1/reminders/today'),
  /** 近期提醒（未来 7 天） */
  upcoming: () => get<ListWrap<Reminder>>('/api/v1/reminders/upcoming'),
}

export const statsApi = {
  /** 统计概览（含全部维度） */
  overview: () => get<StatsOverview>('/api/v1/stats/overview'),
  /** 平台分布 */
  platformDist: () => get<ListWrap<PlatformDistItem>>('/api/v1/stats/platform-distribution'),
  /** 温度分布 */
  tempDist: () => get<TemperatureDist>('/api/v1/stats/temperature-distribution'),
  /** 近 30 天互动趋势 */
  trend: () => get<ListWrap<TrendPoint>>('/api/v1/stats/interaction-trend'),
}
