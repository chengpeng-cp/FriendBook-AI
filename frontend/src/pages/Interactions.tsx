/**
 * 互动记录页：全部互动的时间线视图
 * - 按好友筛选（下拉）
 * - 新增互动（完整表单：类型/内容/地点/心情/记忆卡片）
 * - 记忆卡片特殊视觉（渐变描边）
 */
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { interactionApi } from '@/api/data'
import { friendApi } from '@/api/friends'
import type { Friend, Interaction, InteractionType } from '@/types'
import { usePageEnter } from '@/hooks/useGsap'
import { Empty, PageLoading } from '@/components/ui/Basics'
import Pagination from '@/components/ui/Pagination'
import Modal from '@/components/ui/Modal'
import { toast } from '@/stores/toast'
import { INTERACT_META } from '@/constants'
import { fmtDateTime } from '@/utils/format'

/** 按日期分组 */
function groupByDate(list: Interaction[]): { date: string; items: Interaction[] }[] {
  const groups: { date: string; items: Interaction[] }[] = []
  for (const it of list) {
    const date = it.interact_at.slice(0, 10)
    const last = groups[groups.length - 1]
    if (last && last.date === date) last.items.push(it)
    else groups.push({ date, items: [it] })
  }
  return groups
}

export default function Interactions() {
  const pageRef = usePageEnter<HTMLDivElement>()

  const [list, setList] = useState<Interaction[] | null>(null)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [friendId, setFriendId] = useState<number>(0)
  const [friends, setFriends] = useState<Friend[]>([])

  // 新增弹窗
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({
    friend_id: 0,
    type: 1 as InteractionType,
    content: '',
    location: '',
    mood: '',
    is_memory: false,
    memory_title: '',
  })
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(() => {
    interactionApi
      .list(friendId || undefined, page, 15)
      .then((res) => {
        setList(res.list)
        setTotal(res.total)
      })
      .catch((e) => {
        toast.error((e as Error).message)
        setList([])
      })
  }, [friendId, page])

  useEffect(load, [load])

  // 好友下拉数据（取前 100 个足够）
  useEffect(() => {
    friendApi
      .list({ page: 1, page_size: 100 })
      .then((res) => setFriends(res.list))
      .catch(() => {})
  }, [])

  /** 提交新互动 */
  const submit = async () => {
    if (!form.friend_id) return toast.error('请选择好友')
    if (!form.content.trim()) return toast.error('写点什么吧')
    setSubmitting(true)
    try {
      await interactionApi.create({
        friend_id: form.friend_id,
        type: form.type,
        content: form.content.trim(),
        location: form.location.trim(),
        mood: form.mood.trim(),
        is_memory: form.is_memory,
        memory_title: form.memory_title.trim(),
      })
      toast.success('互动已记录，好感度与温度已联动更新')
      setOpen(false)
      setForm((f) => ({ ...f, content: '', location: '', mood: '', is_memory: false, memory_title: '' }))
      load()
    } catch (e) {
      toast.error((e as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  const nameOf = (fid: number) => friends.find((f) => f.id === fid)?.name ?? `好友#${fid}`

  return (
    <div ref={pageRef} className="page-enter max-w-4xl mx-auto">
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold">互动记录</h1>
          <p className="mt-1 text-xs text-fog-500">共 {total} 条 · 每次互动都会联动刷新关系温度</p>
        </div>
        <button className="fb-btn-primary" onClick={() => setOpen(true)}>＋ 记录互动</button>
      </div>

      {/* 好友筛选 */}
      <div className="fb-card p-3 mb-5 flex items-center gap-3" data-stagger>
        <span className="text-xs text-fog-500 shrink-0">筛选好友</span>
        <select
          className="fb-input flex-1 max-w-60"
          value={friendId}
          onChange={(e) => {
            setFriendId(Number(e.target.value))
            setPage(1)
          }}
        >
          <option value={0}>全部好友</option>
          {friends.map((f) => (
            <option key={f.id} value={f.id}>{f.name}{f.remark && `（${f.remark}）`}</option>
          ))}
        </select>
      </div>

      {/* 时间线（按日期分组） */}
      {list === null ? (
        <PageLoading text="加载互动记录…" />
      ) : list.length === 0 ? (
        <div className="fb-card">
          <Empty icon="✦" title="暂无互动记录" hint="点击右上角「记录互动」，留下你们的第一个回忆" />
        </div>
      ) : (
        <div className="flex flex-col gap-6">
          {groupByDate(list).map(({ date, items }) => (
            <div key={date}>
              {/* 日期标头 */}
              <div className="flex items-center gap-3 mb-3" data-stagger>
                <span className="text-xs font-semibold text-fog-700">{date}</span>
                <span className="h-px flex-1 bg-ink-700/60" />
                <span className="text-[10px] text-fog-300">{items.length} 条</span>
              </div>
              <div className="flex flex-col gap-2.5">
                {items.map((it) => {
                  const meta = INTERACT_META[it.type]
                  return (
                    <div
                      key={it.id}
                      className={`fb-card p-4 flex items-start gap-3.5 ${
                        it.is_memory
                          ? '!border-violet-500/40 shadow-[0_0_20px_rgba(159,122,234,.12)]'
                          : ''
                      }`}
                    >
                      {/* 类型图标 */}
                      <span
                        className="h-9 w-9 shrink-0 rounded-xl flex items-center justify-center text-base"
                        style={{ background: `${meta.color}1a`, border: `1px solid ${meta.color}44` }}
                      >
                        {meta.emoji}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <Link to={`/friends/${it.friend_id}`} className="text-sm font-medium hover:text-brand-400 transition-colors">
                            {nameOf(it.friend_id)}
                          </Link>
                          <span className="text-xs" style={{ color: meta.color }}>{meta.label}</span>
                          <span className="text-[11px] text-fog-500">{fmtDateTime(it.interact_at).slice(11)}</span>
                          {it.is_memory && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded-md bg-violet-500/15 border border-violet-500/30 text-violet-300">
                              💜 记忆卡片{it.memory_title && ` · ${it.memory_title}`}
                            </span>
                          )}
                        </div>
                        <p className="mt-1 text-sm text-fog-700">{it.content}</p>
                        {(it.location || it.mood) && (
                          <p className="mt-1 text-[11px] text-fog-300">
                            {it.location && `📍 ${it.location}`}
                            {it.mood && ` · ${it.mood}`}
                          </p>
                        )}
                      </div>
                      {/* 删除 */}
                      <button
                        className="shrink-0 p-1.5 rounded-lg text-fog-300 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
                        title="删除记录"
                        onClick={async () => {
                          try {
                            await interactionApi.remove(it.id)
                            toast.success('已删除')
                            load()
                          } catch (e) {
                            toast.error((e as Error).message)
                          }
                        }}
                      >
                        <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                          <path d="M3 4h10M6.5 4V2.5h3V4M5 4l.5 9h5l.5-9" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
                        </svg>
                      </button>
                    </div>
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      )}

      <Pagination page={page} pageSize={15} total={total} onChange={setPage} />

      {/* 新增互动弹窗 */}
      <Modal
        open={open}
        onClose={() => setOpen(false)}
        title="记录互动"
        subtitle="记录后自动刷新好友的关系温度"
        footer={
          <>
            <button className="fb-btn-ghost" onClick={() => setOpen(false)}>取消</button>
            <button className="fb-btn-primary" onClick={submit} disabled={submitting}>
              {submitting ? '提交中…' : '保存记录'}
            </button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <div>
            <label className="fb-label">好友 *</label>
            <select
              className="fb-input"
              value={form.friend_id}
              onChange={(e) => setForm((f) => ({ ...f, friend_id: Number(e.target.value) }))}
            >
              <option value={0}>选择好友…</option>
              {friends.map((f) => (
                <option key={f.id} value={f.id}>{f.name}{f.remark && `（${f.remark}）`}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="fb-label">互动类型</label>
            <div className="flex gap-2 flex-wrap">
              {Object.entries(INTERACT_META).map(([k, v]) => (
                <button
                  key={k}
                  className={`px-3.5 py-2 rounded-xl text-xs border transition-all ${
                    form.type === Number(k)
                      ? 'bg-brand-500/20 border-brand-500/50 text-brand-400 font-medium'
                      : 'bg-ink-700/40 border-ink-600/50 text-fog-500 hover:border-fog-300/30'
                  }`}
                  onClick={() => setForm((f) => ({ ...f, type: Number(k) as InteractionType }))}
                >
                  {v.emoji} {v.label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="fb-label">内容 *</label>
            <textarea
              className="fb-input !h-24 py-2.5 resize-none"
              placeholder="聊了什么？去了哪里？"
              value={form.content}
              onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="fb-label">地点</label>
              <input
                className="fb-input"
                placeholder="选填"
                value={form.location}
                onChange={(e) => setForm((f) => ({ ...f, location: e.target.value }))}
              />
            </div>
            <div>
              <label className="fb-label">心情</label>
              <input
                className="fb-input"
                placeholder="如：开心 😊"
                value={form.mood}
                onChange={(e) => setForm((f) => ({ ...f, mood: e.target.value }))}
              />
            </div>
          </div>

          {/* 记忆卡片开关 */}
          <label className="flex items-center gap-3 p-3.5 rounded-xl bg-violet-500/5 border border-violet-500/25 cursor-pointer">
            <input
              type="checkbox"
              className="accent-violet-400 w-4 h-4"
              checked={form.is_memory}
              onChange={(e) => setForm((f) => ({ ...f, is_memory: e.target.checked }))}
            />
            <div>
              <p className="text-sm text-violet-300">💜 标记为记忆卡片</p>
              <p className="text-[11px] text-fog-500">值得铭记的时刻会以特殊样式展示</p>
            </div>
          </label>

          {form.is_memory && (
            <div>
              <label className="fb-label">记忆标题</label>
              <input
                className="fb-input"
                placeholder="给这个时刻起个名字"
                value={form.memory_title}
                onChange={(e) => setForm((f) => ({ ...f, memory_title: e.target.value }))}
              />
            </div>
          )}
        </div>
      </Modal>
    </div>
  )
}
