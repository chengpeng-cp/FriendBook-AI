/**
 * 好友详情页
 * - 顶部档案卡：头像、基本信息、好感度滑块（联动温度展示）
 * - 平台账号管理：列表 + 添加
 * - 互动时间线：按时间倒序 + 新增互动
 */
import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { friendApi } from '@/api/friends'
import { interactionApi } from '@/api/data'
import type { Account, Friend, Interaction, Temperature } from '@/types'
import { usePageEnter } from '@/hooks/useGsap'
import { Avatar, Empty, FavorBar, PageLoading, TempBadge } from '@/components/ui/Basics'
import { toast } from '@/stores/toast'
import { GENDER_META, INTERACT_META, PLATFORM_META, TEMP_META, favorToTemp } from '@/constants'
import { fmtDate, fmtDateTime, fmtRelative } from '@/utils/format'

/** 档案字段小卡 */
function Field({ label, value }: { label: string; value: string }) {
  return (
    <div className="p-3 rounded-xl bg-ink-900/60 border border-ink-700/50">
      <p className="text-[11px] text-fog-500">{label}</p>
      <p className="mt-0.5 text-sm truncate">{value || '—'}</p>
    </div>
  )
}

export default function FriendDetail() {
  const { id } = useParams<{ id: string }>()
  const friendId = Number(id)
  const navigate = useNavigate()
  const pageRef = usePageEnter<HTMLDivElement>([friendId])

  const [friend, setFriend] = useState<Friend | null>(null)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [interactions, setInteractions] = useState<Interaction[] | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [favor, setFavor] = useState(50)

  // 新账号表单
  const [accPlatform, setAccPlatform] = useState('wechat')
  const [accAccount, setAccAccount] = useState('')
  const [accRemark, setAccRemark] = useState('')

  // 新互动表单
  const [itType, setItType] = useState<1 | 2 | 3 | 4 | 5 | 6>(1)
  const [itContent, setItContent] = useState('')
  const [itMood, setItMood] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(() => {
    friendApi
      .detail(friendId)
      .then((f) => {
        setFriend(f)
        setFavor(f.favorability)
        setAccounts(f.accounts ?? [])
        return interactionApi.list(friendId, 1, 50)
      })
      .then((res) => setInteractions(res.list))
      .catch((e) => {
        setNotFound(true)
        toast.error((e as Error).message)
      })
  }, [friendId])

  useEffect(load, [load])

  if (notFound) {
    return (
      <div className="fb-card p-12 text-center">
        <Empty icon="🫥" title="好友不存在或已被删除" hint="可能已被其他端删除" />
        <Link to="/friends" className="fb-btn-primary mt-2">返回好友列表</Link>
      </div>
    )
  }
  if (!friend) return <PageLoading text="加载好友档案…" />

  /** 调整好感度 */
  const adjustFavor = async (v: number) => {
    try {
      const res = await friendApi.adjustFavor(friendId, v)
      setFriend((f) => (f ? { ...f, favorability: res.favorability, temperature: res.temperature } : f))
      toast.success(`好感度 ${res.favorability} · 温度 ${res.temp_label}`)
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  /** 添加平台账号 */
  const addAccount = async () => {
    if (!accAccount.trim()) return toast.error('账号不能为空')
    try {
      await friendApi.addAccount(friendId, { platform: accPlatform, account: accAccount.trim(), remark: accRemark })
      const res = await friendApi.accounts(friendId)
      setAccounts(res.list)
      setAccAccount('')
      setAccRemark('')
      toast.success('账号已添加')
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  /** 记录互动（联动温度刷新） */
  const addInteraction = async () => {
    if (!itContent.trim()) return toast.error('写点什么吧')
    setSubmitting(true)
    try {
      await interactionApi.create({ friend_id: friendId, type: itType, content: itContent.trim(), mood: itMood.trim() })
      toast.success('互动已记录')
      setItContent('')
      setItMood('')
      load() // 好感度/温度/最近互动都会联动，整页刷新
    } catch (e) {
      toast.error((e as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  const previewTemp: Temperature = favorToTemp(favor)

  return (
    <div ref={pageRef} className="page-enter max-w-5xl mx-auto">
      <button className="text-xs text-fog-500 hover:text-fog-900 mb-4 transition-colors" onClick={() => navigate(-1)}>
        ← 返回
      </button>

      {/* ===== 档案头卡 ===== */}
      <div className="fb-card p-6 relative overflow-hidden" data-stagger>
        <div
          className="absolute -right-16 -top-16 h-48 w-48 rounded-full blur-3xl opacity-15"
          style={{ background: TEMP_META[friend.temperature].color }}
        />
        <div className="relative flex items-start gap-5">
          <Avatar id={friend.id} name={friend.name} size={72} />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-bold">{friend.name}</h1>
              {friend.remark && <span className="text-sm text-fog-500">「{friend.remark}」</span>}
              <TempBadge temp={friend.temperature} />
              {friend.mbti && (
                <span className="px-2 py-0.5 rounded-lg bg-violet-500/15 border border-violet-500/30 text-violet-300 text-xs font-semibold">
                  {friend.mbti}
                </span>
              )}
            </div>
            {friend.bio && <p className="mt-2 text-sm text-fog-700">{friend.bio}</p>}
            <p className="mt-2 text-xs text-fog-500">
              {GENDER_META[friend.gender]} · 认识于 {friend.meet_scene || '未知场景'} · 最近互动 {fmtRelative(friend.last_interact)}
            </p>
          </div>
        </div>

        {/* 好感度调整区 */}
        <div className="relative mt-6 p-4 rounded-xl bg-ink-900/70 border border-ink-600/50">
          <div className="flex items-center justify-between mb-3">
            <span className="text-xs text-fog-500">好感度 · 拖动后自动保存并联动温度</span>
            <div className="flex items-center gap-2">
              <span className="text-lg font-bold tabular-nums" style={{ color: TEMP_META[previewTemp].color }}>{favor}</span>
              <TempBadge temp={previewTemp} size="sm" />
            </div>
          </div>
          <input
            type="range"
            min={0}
            max={100}
            value={favor}
            onChange={(e) => setFavor(Number(e.target.value))}
            onMouseUp={(e) => adjustFavor(Number((e.target as HTMLInputElement).value))}
            onTouchEnd={(e) => adjustFavor(Number((e.target as HTMLInputElement).value))}
            className="w-full accent-brand-500 cursor-pointer"
            aria-label="好感度"
          />
          <div className="mt-2"><FavorBar favor={favor} /></div>
        </div>

        {/* 字段网格 */}
        <div className="relative mt-4 grid grid-cols-2 md:grid-cols-4 gap-2.5">
          <Field label="生日" value={friend.birthday ? fmtDate(friend.birthday) : ''} />
          <Field label="手机号" value={friend.phone} />
          <Field label="邮箱" value={friend.email} />
          <Field label="城市" value={friend.location} />
          <Field label="职业" value={friend.occupation} />
          <Field label="分组" value={(friend.groups ?? []).map((g) => g.name).join('、')} />
          <Field label="标签" value={(friend.tags ?? []).map((t) => `#${t.name}`).join(' ')} />
          <Field label="加入时间" value={fmtDate(friend.created_at)} />
        </div>
      </div>

      {/* ===== 两栏：账号 + 互动 ===== */}
      <div className="mt-5 grid lg:grid-cols-5 gap-5">
        {/* 平台账号 */}
        <div className="fb-card p-5 lg:col-span-2" data-stagger>
          <h2 className="text-sm font-semibold mb-4">平台账号（{accounts.length}）</h2>
          <ul className="flex flex-col gap-2">
            {accounts.length === 0 && <Empty icon="🔗" title="暂无账号" hint="在下方添加 TA 的社交账号" />}
            {accounts.map((a) => (
              <li key={a.id} className="flex items-center gap-3 p-3 rounded-xl bg-ink-900/60 border border-ink-700/50">
                <span className="text-lg">{PLATFORM_META[a.platform]?.emoji ?? '🔗'}</span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm truncate">{a.account}</p>
                  <p className="text-[11px] text-fog-500">
                    {PLATFORM_META[a.platform]?.label ?? a.platform}
                    {a.remark && ` · ${a.remark}`}
                  </p>
                </div>
              </li>
            ))}
          </ul>

          {/* 添加账号 */}
          <div className="mt-4 pt-4 border-t border-ink-700/50 flex flex-col gap-2">
            <div className="flex gap-2">
              <select className="fb-input w-28 shrink-0" value={accPlatform} onChange={(e) => setAccPlatform(e.target.value)}>
                {Object.entries(PLATFORM_META).map(([k, v]) => (
                  <option key={k} value={k}>{v.emoji} {v.label}</option>
                ))}
              </select>
              <input className="fb-input flex-1" placeholder="账号" value={accAccount} onChange={(e) => setAccAccount(e.target.value)} />
            </div>
            <div className="flex gap-2">
              <input className="fb-input flex-1" placeholder="备注（可选）" value={accRemark} onChange={(e) => setAccRemark(e.target.value)} />
              <button className="fb-btn-primary shrink-0" onClick={addAccount}>添加</button>
            </div>
          </div>
        </div>

        {/* 互动时间线 */}
        <div className="fb-card p-5 lg:col-span-3" data-stagger>
          <h2 className="text-sm font-semibold mb-4">互动时间线</h2>

          {/* 新增互动 */}
          <div className="p-4 rounded-xl bg-ink-900/70 border border-ink-600/50 mb-5">
            <div className="flex gap-2 mb-2.5 flex-wrap">
              {Object.entries(INTERACT_META).map(([k, v]) => (
                <button
                  key={k}
                  className={`px-3 py-1.5 rounded-xl text-xs border transition-all ${
                    itType === Number(k)
                      ? 'bg-brand-500/20 border-brand-500/50 text-brand-400 font-medium'
                      : 'bg-ink-700/40 border-ink-600/50 text-fog-500 hover:border-fog-300/30'
                  }`}
                  onClick={() => setItType(Number(k) as 1 | 2 | 3 | 4 | 5 | 6)}
                >
                  {v.emoji} {v.label}
                </button>
              ))}
            </div>
            <div className="flex gap-2">
              <input
                className="fb-input flex-1"
                placeholder={`这次和 ${friend.name} 做了什么？`}
                value={itContent}
                onChange={(e) => setItContent(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addInteraction()}
              />
              <input
                className="fb-input w-28 shrink-0"
                placeholder="心情 😊"
                value={itMood}
                onChange={(e) => setItMood(e.target.value)}
              />
              <button className="fb-btn-primary shrink-0" onClick={addInteraction} disabled={submitting}>
                {submitting ? '…' : '记录'}
              </button>
            </div>
          </div>

          {/* 时间线列表 */}
          {interactions === null ? (
            <PageLoading />
          ) : interactions.length === 0 ? (
            <Empty icon="✦" title="还没有互动记录" hint="记录每一次聊天、见面，让关系有迹可循" />
          ) : (
            <ol className="relative border-l border-ink-600/50 ml-2 flex flex-col gap-5">
              {interactions.map((it) => {
                const meta = INTERACT_META[it.type]
                return (
                  <li key={it.id} className="ml-5 relative">
                    {/* 节点 */}
                    <span
                      className="absolute -left-[30px] top-0.5 h-4 w-4 rounded-full border-2 bg-ink-950"
                      style={{ borderColor: meta.color, background: `${meta.color}33` }}
                    />
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-xs font-semibold" style={{ color: meta.color }}>
                        {meta.emoji} {meta.label}
                      </span>
                      <span className="text-[11px] text-fog-500">{fmtDateTime(it.interact_at)}</span>
                      {it.is_memory && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded-md bg-violet-500/15 border border-violet-500/30 text-violet-300">
                          记忆卡片 {it.memory_title && `· ${it.memory_title}`}
                        </span>
                      )}
                    </div>
                    <p className="mt-1 text-sm text-fog-700">{it.content}</p>
                    {(it.location || it.mood) && (
                      <p className="mt-0.5 text-[11px] text-fog-300">
                        {it.location && `📍 ${it.location}`}
                        {it.mood && ` · ${it.mood}`}
                      </p>
                    )}
                  </li>
                )
              })}
            </ol>
          )}
        </div>
      </div>
    </div>
  )
}
