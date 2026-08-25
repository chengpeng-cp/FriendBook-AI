/** 好友管理 API（含 CSV 导入导出、好感度调整、平台账号） */
import client, { del, get, post, put } from './client'
import type {
  Account,
  AccountParam,
  AdjustFavorResult,
  CreateFriendPayload,
  Friend,
  FriendQuery,
  ImportResult,
  ListWrap,
  Paged,
  UpdateFriendPayload,
} from '@/types'

/** 构造查询串（过滤 undefined，保持 URL 干净） */
function qs(query: FriendQuery): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [k, v] of Object.entries(query)) {
    if (v !== undefined && v !== '' && v !== 0) out[k] = v
  }
  return out
}

export const friendApi = {
  /** 好友列表（分页 + 多条件筛选） */
  list: (query: FriendQuery) => get<Paged<Friend>>('/api/v1/friends', qs(query)),

  /** 关键词搜索 */
  search: (keyword: string, page = 1, pageSize = 10) =>
    get<Paged<Friend>>('/api/v1/friends/search', { keyword, page, page_size: pageSize }),

  /** 关系变凉列表（最久未互动优先） */
  cold: (page = 1, pageSize = 10) =>
    get<Paged<Friend>>('/api/v1/friends/cold', { page, page_size: pageSize }),

  /** 好友详情（含 accounts/groups/tags 关联） */
  detail: (id: number) => get<Friend>(`/api/v1/friends/${id}`),

  /** 创建好友 */
  create: (payload: CreateFriendPayload) => post<Friend>('/api/v1/friends', payload),

  /** 更新好友 */
  update: (id: number, payload: UpdateFriendPayload) => put<null>(`/api/v1/friends/${id}`, payload),

  /** 删除好友 */
  remove: (id: number) => del<null>(`/api/v1/friends/${id}`),

  /** 调整好感度（后端联动关系温度） */
  adjustFavor: (id: number, favorability: number) =>
    put<AdjustFavorResult>(`/api/v1/friends/${id}/favorability`, { favorability }),

  /** 好友的平台账号列表 */
  accounts: (id: number) => get<ListWrap<Account>>(`/api/v1/friends/${id}/accounts`),

  /** 添加平台账号 */
  addAccount: (id: number, payload: AccountParam) =>
    post<null>(`/api/v1/friends/${id}/accounts`, payload),

  /**
   * 导出 CSV（blob 下载：手动构造 <a> 触发浏览器保存）
   */
  async exportCsv(): Promise<void> {
    const res = await client.get<Blob>('/api/v1/friends/export', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `friends-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  },

  /** 导入 CSV（multipart 上传） */
  importCsv: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return client
      .post<{ code: number; message: string; data: ImportResult }>('/api/v1/friends/import', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((res) => {
        if (res.data.code !== 0) throw new Error(res.data.message)
        return res.data.data
      })
  },
}
