/**
 * 好友管理页：列表 + 筛选 + 增删改 + 好感度快调 + CSV 导入导出
 * 数据加载后卡片瀑布式交错入场（GSAP）
 */
import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { friendApi } from '@/api/friends'
import type { Friend, FriendQuery } from '@/types'
import { usePageEnter, useStaggerIn } from '@/hooks/useGsap'
import { Avatar, Empty, FavorBar, PageLoading, TempBadge } from '@/components/ui/Basics'
import Pagination from '@/components/ui/Pagination'
import Modal from '@/components/ui/Modal'
import FriendFormModal from '@/components/friends/FriendFormModal'
import { toast } from '@/stores/toast'
import { TEMP_META, favorToTemp } from '@/constants'
import { fmtRelative } from '@/utils/format'

const SORTS: { value: string; label: string }[] = [
  { value: '', label: '默认排序' },
  { value: 'favorability_desc', label: '好感度 高→低' },
  { value: 'favorability_asc', label: '好感度 低→高' },
  { value: 'last_interact_asc', label: '最久未互动' },
  { value: 'last_interact_desc', label: '最近互动' },
]

/** 好友卡片 */
function FriendCard({
  friend,
  onEdit,
  onDelete,
  onAdjustFavor,
}: {
  friend: Friend
  onEdit: (f: Friend) => void
  onDelete: (f: Friend) => void
  onAdjustFavor: (f: Friend, v: number) => void
}) {
  const [hover, setHover] = useState(false)
  const [adjusting, setAdjusting] = useState(false)
  const [favor, setFavor] = useState(friend.favorability)
  const cardRef = useRef<HTMLDivElement>(null)

  // 快调面板开关时轻量动画
  useEffect(() => {
    if (!cardRef.current) return
    const panel = cardRef.current.querySelector('[data-favor-panel]')
    if (panel) panel.classList.toggle('hidden', !adjusting)
  }, [adjusting])

  return (
    <div
      ref={cardRef}
      className="fb-card p-4 group relative overflow-hidden hover:border-fog-300/20 hover:-translate-y-0.5 transition-all duration-200"
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => {
        setHover(false)
        setAdjusting(false)
      }}
    >
      <div className="flex items-start gap-3">
        <Link to={`/friends/${friend.id}`}>
          <Avatar id={friend.id} name={friend.name} size={44} />
        </Link>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Link
              to={`/friends/${friend.id}`}
              className="text-sm font-semibold truncate hover:text-brand-400 transition-colors"
            >
              {friend.name}
            </Link>
            {friend.remark && <span className="text-xs text-fog-500 truncate">{friend.remark}</span>}
          </div>
          <p className="mt-0.5 text-[11px] text-fog-500 truncate">
            {[friend.location, friend.occupation, friend.mbti].filter(Boolean).join(' · ') || '资料待完善'}
          </p>
        </div>
        <TempBadge temp={friend.temperature} size="sm" />
      </div>

      {/* 好感度 */}
      <div className="mt-3.5 flex items-center gap-2">
        <button
          className="text-xs shrink-0 heart"
          title="快调好感度"
          onClick={() => setAdjusting((v) => !v)}
        >
          {favor >= 75 ? '❤️' : favor >= 40 ? '💛' : '🤍'}
        </button>
        <FavorBar favor={favor} />
        <span className="text-xs tabular-nums text-fog-700 w-7 text-right">{favor}</span>
      </div>

      {/* 快调面板 */}
      <div data-favor-panel className="hidden mt-3">
        <div className="flex items-center gap-2">
          <input
            type="range"
            min={0}
            max={100}
            value={favor}
            onChange={(e) => setFavor(Number(e.target.value))}
            className="flex-1 accent-brand-500 cursor-pointer"
            aria-label="调整好感度"
          />
          <button
            className="fb-btn-primary !h-7 !px-3 text-xs"
            onClick={() => {
              onAdjustFavor(friend, favor)
              setAdjusting(false)
            }}
          >
            保存
          </button>
        </div>
        <p className="mt-1 text-[10px]" style={{ color: TEMP_META[favorToTemp(favor)].color }}>
          {TEMP_META[favorToTemp(favor)].emoji} 联动温度：{TEMP_META[favorToTemp(favor)].label} · {TEMP_META[favorToTemp(favor)].desc}
        </p>
      </div>

      {/* 底部：最近互动 + 操作 */}
      <div className="mt-3 pt-3 border-t border-ink-700/50 flex items-center justify-between">
        <span className="text-[11px] text-fog-500">最近互动 {fmtRelative(friend.last_interact)}</span>
        <div className={`flex gap-1 transition-opacity ${hover ? 'opacity-100' : 'opacity-0'}`}>
          <Link
            to={`/friends/${friend.id}`}
            className="px-2 py-1 rounded-lg text-[11px] text-fog-500 hover:text-sky-400 hover:bg-sky-500/10 transition-colors"
          >
            详情
          </Link>
          <button
            className="px-2 py-1 rounded-lg text-[11px] text-fog-500 hover:text-brand-400 hover:bg-brand-500/10 transition-colors"
            onClick={() => onEdit(friend)}
          >
            编辑
          </button>
          <button
            className="px-2 py-1 rounded-lg text-[11px] text-fog-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
            onClick={() => onDelete(friend)}
          >
            删除
          </button>
        </div>
      </div>
    </div>
  )
}

export default function Friends() {
  const pageRef = usePageEnter<HTMLDivElement>()

  // 列表数据
  const [friends, setFriends] = useState<Friend[] | null>(null)
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState<FriendQuery>({ page: 1, page_size: 12, keyword: '' })

  // 搜索防抖
  const [keyword, setKeyword] = useState('')
  const debounceRef = useRef<ReturnType<typeof setTimeout>>()
  useEffect(() => {
    clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      setQuery((q) => ({ ...q, keyword, page: 1 }))
    }, 350)
    return () => clearTimeout(debounceRef.current)
  }, [keyword])

  // 拉取列表
  const load = useCallback(() => {
    friendApi
      .list(query)
      .then((res) => {
        setFriends(res.list)
        setTotal(res.total)
      })
      .catch((e) => {
        toast.error((e as Error).message)
        setFriends([])
      })
  }, [query])

  useEffect(load, [load])

  // 卡片交错入场（数据变化时重播）
  const gridRef = useStaggerIn<HTMLDivElement>([friends])

  // 弹窗状态
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Friend | null>(null)
  const [deleting, setDeleting] = useState<Friend | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  /** 调整好感度（联动温度，卡片就地刷新） */
  const adjustFavor = async (f: Friend, v: number) => {
    try {
      const res = await friendApi.adjustFavor(f.id, v)
      setFriends((list) =>
        (list ?? []).map((it) =>
          it.id === f.id ? { ...it, favorability: res.favorability, temperature: res.temperature } : it,
        ),
      )
      toast.success(`已调整为 ${res.favorability} · 温度 ${res.temp_label}`)
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  /** 删除确认 */
  const confirmDelete = async () => {
    if (!deleting) return
    try {
      await friendApi.remove(deleting.id)
      toast.success(`已删除好友「${deleting.name}」`)
      setDeleting(null)
      load()
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  /** CSV 导入 */
  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // 允许重复选择同一文件
    if (!file) return
    try {
      const result = await friendApi.importCsv(file)
      toast.success(`导入完成：成功 ${result.success_count} 条` + (result.failed_rows.length > 0 ? `，失败 ${result.failed_rows.length} 条` : ''))
      if (result.failed_rows.length > 0) {
        console.warn('导入失败行：', result.failed_rows)
      }
      load()
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  const tempFilters: { value: number; label: string }[] = [
    { value: 0, label: '全部' },
    { value: 1, label: '🔥 夯' },
    { value: 2, label: '😊 热' },
    { value: 3, label: '🙂 温' },
    { value: 4, label: '😐 凉' },
    { value: 5, label: '🥶 拉' },
  ]

  return (
    <div ref={pageRef} className="page-enter max-w-6xl mx-auto">
      {/* 标题 + 主操作 */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-xl font-bold">好友管理</h1>
          <p className="mt-1 text-xs text-fog-500">共 {total} 位好友 · 点击 ❤️ 快调好感度</p>
        </div>
        <div className="flex items-center gap-2.5">
          <input ref={fileRef} type="file" accept=".csv" className="hidden" onChange={handleImport} />
          <button className="fb-btn-ghost" onClick={() => fileRef.current?.click()}>📥 导入</button>
          <button
            className="fb-btn-ghost"
            onClick={() =>
              friendApi.exportCsv().then(() => toast.success('CSV 已开始下载')).catch((e) => toast.error((e as Error).message))
            }
          >
            📤 导出
          </button>
          <button
            className="fb-btn-primary"
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            ＋ 添加好友
          </button>
        </div>
      </div>

      {/* 工具栏：搜索 + 温度筛选 + 排序 */}
      <div className="fb-card p-3 mb-5 flex flex-wrap items-center gap-3" data-stagger>
        <div className="relative flex-1 min-w-48">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-fog-300 text-sm">🔍</span>
          <input
            className="fb-input !pl-9"
            placeholder="搜索姓名 / 备注 / 手机号…"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
          />
        </div>
        <div className="flex gap-1">
          {tempFilters.map((tf) => (
            <button
              key={tf.value}
              className={`px-3 h-9 rounded-xl text-xs transition-all ${
                (query.temperature ?? 0) === tf.value
                  ? 'bg-brand-500/20 border border-brand-500/50 text-brand-400 font-medium'
                  : 'bg-ink-700/40 border border-ink-600/50 text-fog-500 hover:border-fog-300/30'
              }`}
              onClick={() => setQuery((q) => ({ ...q, temperature: tf.value || undefined, page: 1 }))}
            >
              {tf.label}
            </button>
          ))}
        </div>
        <select
          className="fb-input w-40"
          value={query.sort ?? ''}
          onChange={(e) => setQuery((q) => ({ ...q, sort: (e.target.value || undefined) as FriendQuery['sort'] ?? undefined, page: 1 }))}
        >
          {SORTS.map((s) => (
            <option key={s.label} value={s.value}>{s.label}</option>
          ))}
        </select>
      </div>

      {/* 卡片网格 */}
      {friends === null ? (
        <PageLoading text="加载好友列表…" />
      ) : friends.length === 0 ? (
        <div className="fb-card">
          <Empty
            icon={keyword || query.temperature ? '🔍' : '🌱'}
            title={keyword || query.temperature ? '没有符合条件的好友' : '还没有好友'}
            hint={keyword || query.temperature ? '换个关键词或温度档位试试' : '点击右上角「添加好友」开始记录你的社交圈'}
          />
        </div>
      ) : (
        <div ref={gridRef} className="grid sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {friends.map((f) => (
            <FriendCard
              key={f.id}
              friend={f}
              onEdit={(fr) => {
                setEditing(fr)
                setFormOpen(true)
              }}
              onDelete={setDeleting}
              onAdjustFavor={adjustFavor}
            />
          ))}
        </div>
      )}

      <Pagination
        page={query.page ?? 1}
        pageSize={query.page_size ?? 12}
        total={total}
        onChange={(p) => setQuery((q) => ({ ...q, page: p }))}
      />

      {/* 创建/编辑弹窗 */}
      <FriendFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        friend={editing}
        onSaved={load}
      />

      {/* 删除确认 */}
      <Modal
        open={!!deleting}
        onClose={() => setDeleting(null)}
        title="删除好友"
        width="max-w-sm"
        footer={
          <>
            <button className="fb-btn-ghost" onClick={() => setDeleting(null)}>再想想</button>
            <button className="fb-btn-danger" onClick={confirmDelete}>确认删除</button>
          </>
        }
      >
        <p className="text-sm text-fog-700 leading-relaxed">
          确定删除好友 <span className="text-rose-400 font-semibold">「{deleting?.name}」</span> 吗？
          <br />
          <span className="text-xs text-fog-500">TA 的平台账号与互动记录将一并删除，操作不可恢复。</span>
        </p>
      </Modal>
    </div>
  )
}
