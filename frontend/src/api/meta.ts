/** 分组 / 标签 API */
import { del, get, post, put } from './client'
import type { Group, ListWrap, Tag } from '@/types'

export const groupApi = {
  list: () => get<ListWrap<Group>>('/api/v1/groups'),
  create: (name: string, sortOrder = 0) =>
    post<Group>('/api/v1/groups', { name, sort_order: sortOrder }),
  update: (id: number, payload: { name?: string; sort_order?: number }) =>
    put<null>(`/api/v1/groups/${id}`, payload),
  remove: (id: number) => del<null>(`/api/v1/groups/${id}`),
}

export const tagApi = {
  list: () => get<ListWrap<Tag>>('/api/v1/tags'),
  create: (name: string, color = '') => post<Tag>('/api/v1/tags', { name, color }),
  remove: (id: number) => del<null>(`/api/v1/tags/${id}`),
}
