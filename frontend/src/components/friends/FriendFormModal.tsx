/**
 * 好友创建/编辑表单弹窗
 * - 好感度滑块：拖动实时预览温度档位（与后端联动规则一致）
 * - 平台账号：动态增删行
 * - 分组/标签： chips 多选
 */
import { useEffect, useState, type FormEvent } from 'react'
import Modal from '@/components/ui/Modal'
import { friendApi } from '@/api/friends'
import { groupApi, tagApi } from '@/api/meta'
import { toast } from '@/stores/toast'
import { PLATFORM_META, TEMP_META, favorToTemp, GENDER_META } from '@/constants'
import type { Friend, Gender, Group, Tag, Temperature } from '@/types'

interface Props {
  open: boolean
  onClose: () => void
  /** 传入则为编辑模式（预填数据） */
  friend?: Friend | null
  onSaved: () => void
}

/** 账号编辑行 */
interface AccountRow {
  platform: string
  account: string
  remark: string
}

export default function FriendFormModal({ open, onClose, friend, onSaved }: Props) {
  const isEdit = !!friend

  // ---- 表单状态 ----
  const [name, setName] = useState('')
  const [remark, setRemark] = useState('')
  const [gender, setGender] = useState<Gender>(0)
  const [birthday, setBirthday] = useState('')
  const [phone, setPhone] = useState('')
  const [email, setEmail] = useState('')
  const [location, setLocation] = useState('')
  const [occupation, setOccupation] = useState('')
  const [bio, setBio] = useState('')
  const [meetScene, setMeetScene] = useState('')
  const [favor, setFavor] = useState(50)
  const [mbti, setMbti] = useState('')
  const [accounts, setAccounts] = useState<AccountRow[]>([])
  const [groupIds, setGroupIds] = useState<number[]>([])
  const [tagIds, setTagIds] = useState<number[]>([])

  const [groups, setGroups] = useState<Group[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  // 好感度 → 温度预览
  const previewTemp: Temperature = favorToTemp(favor)
  const previewMeta = TEMP_META[previewTemp]

  // 打开时：编辑模式预填 / 新建模式重置；并行拉分组标签
  useEffect(() => {
    if (!open) return
    setError('')
    groupApi.list().then((r) => setGroups(r.list)).catch(() => {})
    tagApi.list().then((r) => setTags(r.list)).catch(() => {})

    if (friend) {
      setName(friend.name)
      setRemark(friend.remark)
      setGender(friend.gender)
      setBirthday(friend.birthday ? friend.birthday.slice(0, 10) : '')
      setPhone(friend.phone)
      setEmail(friend.email)
      setLocation(friend.location)
      setOccupation(friend.occupation)
      setBio(friend.bio)
      setMeetScene(friend.meet_scene)
      setFavor(friend.favorability)
      setMbti(friend.mbti)
      setAccounts((friend.accounts ?? []).map((a) => ({ platform: a.platform, account: a.account, remark: a.remark })))
      setGroupIds((friend.groups ?? []).map((g) => g.id))
      setTagIds((friend.tags ?? []).map((t) => t.id))
    } else {
      setName(''); setRemark(''); setGender(0); setBirthday(''); setPhone(''); setEmail('')
      setLocation(''); setOccupation(''); setBio(''); setMeetScene(''); setFavor(50); setMbti('')
      setAccounts([]); setGroupIds([]); setTagIds([])
    }
  }, [open, friend])

  /** chips 切换 */
  const toggleId = (list: number[], setList: (v: number[]) => void, id: number) =>
    setList(list.includes(id) ? list.filter((i) => i !== id) : [...list, id])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return setError('姓名不能为空')
    if (mbti && mbti.length !== 4) return setError('MBTI 必须是 4 个字符，如 INFP')
    if (email && !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email)) return setError('邮箱格式不正确')

    setSubmitting(true)
    setError('')
    try {
      if (isEdit) {
        // 更新：仅传有值字段（后端指针语义，未传不更新）
        await friendApi.update(friend!.id, {
          name: name.trim() || undefined,
          remark: remark.trim() || undefined,
          phone: phone.trim() || undefined,
          email: email.trim() || undefined,
          location: location.trim() || undefined,
          occupation: occupation.trim() || undefined,
          bio: bio.trim() || undefined,
          mbti: mbti.trim() || undefined,
          group_ids: groupIds,
          tag_ids: tagIds,
        })
        // 好感度单独接口（联动温度）
        if (favor !== friend!.favorability) {
          await friendApi.adjustFavor(friend!.id, favor)
        }
        toast.success('好友信息已更新')
      } else {
        await friendApi.create({
          name: name.trim(),
          remark: remark.trim(),
          gender,
          birthday: birthday || undefined,
          phone: phone.trim(),
          email: email.trim(),
          location: location.trim(),
          occupation: occupation.trim(),
          bio: bio.trim(),
          meet_scene: meetScene.trim(),
          favorability: favor,
          mbti: mbti.trim(),
          accounts: accounts
            .filter((a) => a.platform && a.account.trim())
            .map((a) => ({ platform: a.platform, account: a.account.trim(), remark: a.remark })),
          group_ids: groupIds,
          tag_ids: tagIds,
        })
        toast.success(`好友「${name.trim()}」已创建`)
      }
      onSaved()
      onClose()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      width="max-w-2xl"
      title={isEdit ? `编辑 ${friend!.name}` : '添加好友'}
      subtitle={isEdit ? '修改资料或调整好感度' : '记录一段新关系的开始'}
      footer={
        <>
          <button className="fb-btn-ghost" onClick={onClose}>取消</button>
          <button className="fb-btn-primary" onClick={submit} disabled={submitting}>
            {submitting ? '保存中…' : isEdit ? '保存修改' : '创建好友'}
          </button>
        </>
      }
    >
      <form onSubmit={submit} className="flex flex-col gap-4" noValidate>
        {error && (
          <p className="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/30 rounded-lg px-3 py-2">
            {error}
          </p>
        )}

        {/* 基本信息 */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="fb-label">姓名 *</label>
            <input className="fb-input" value={name} onChange={(e) => setName(e.target.value)} placeholder="TA 是谁" />
          </div>
          <div>
            <label className="fb-label">备注名</label>
            <input className="fb-input" value={remark} onChange={(e) => setRemark(e.target.value)} placeholder="怎么称呼 TA" />
          </div>
          <div>
            <label className="fb-label">性别</label>
            <select className="fb-input" value={gender} onChange={(e) => setGender(Number(e.target.value) as Gender)}>
              {Object.entries(GENDER_META).map(([v, label]) => (
                <option key={v} value={v}>{label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="fb-label">生日</label>
            <input type="date" className="fb-input" value={birthday} onChange={(e) => setBirthday(e.target.value)} />
          </div>
          <div>
            <label className="fb-label">手机号</label>
            <input className="fb-input" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="选填" />
          </div>
          <div>
            <label className="fb-label">邮箱</label>
            <input className="fb-input" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="选填" />
          </div>
          <div>
            <label className="fb-label">所在城市</label>
            <input className="fb-input" value={location} onChange={(e) => setLocation(e.target.value)} placeholder="如：广州" />
          </div>
          <div>
            <label className="fb-label">职业</label>
            <input className="fb-input" value={occupation} onChange={(e) => setOccupation(e.target.value)} placeholder="如：产品经理" />
          </div>
          <div>
            <label className="fb-label">MBTI</label>
            <input
              className="fb-input uppercase"
              maxLength={4}
              value={mbti}
              onChange={(e) => setMbti(e.target.value.toUpperCase())}
              placeholder="如 ENFP"
            />
          </div>
          <div>
            <label className="fb-label">认识场景</label>
            <input className="fb-input" value={meetScene} onChange={(e) => setMeetScene(e.target.value)} placeholder="在哪认识的" />
          </div>
        </div>

        <div>
          <label className="fb-label">一句话简介</label>
          <input className="fb-input" value={bio} onChange={(e) => setBio(e.target.value)} placeholder="用一句话描述 TA" />
        </div>

        {/* 好感度滑块（温度实时联动预览） */}
        <div className="p-4 rounded-xl bg-ink-900/70 border border-ink-600/50">
          <div className="flex items-center justify-between mb-3">
            <label className="fb-label mb-0">好感度</label>
            <span className="flex items-center gap-2 text-xs">
              <span className="font-bold text-lg tabular-nums" style={{ color: previewMeta.color }}>{favor}</span>
              <span
                className="px-2 py-0.5 rounded-full text-[11px] font-semibold"
                style={{ color: previewMeta.color, background: `${previewMeta.color}1a`, border: `1px solid ${previewMeta.color}44` }}
              >
                {previewMeta.emoji} {previewMeta.label}
              </span>
            </span>
          </div>
          <input
            type="range"
            min={0}
            max={100}
            value={favor}
            onChange={(e) => setFavor(Number(e.target.value))}
            className="w-full accent-brand-500 cursor-pointer"
            aria-label="好感度"
          />
          <div className="flex justify-between text-[10px] text-fog-300 mt-1">
            <span>0 · 拉 🥶</span>
            <span>50</span>
            <span>100 · 夯 🔥</span>
          </div>
        </div>

        {/* 平台账号 */}
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className="fb-label mb-0">平台账号</label>
            <button
              type="button"
              className="text-xs text-brand-400 hover:text-brand-500"
              onClick={() => setAccounts([...accounts, { platform: 'wechat', account: '', remark: '' }])}
            >
              + 添加账号
            </button>
          </div>
          {accounts.length === 0 ? (
            <p className="text-xs text-fog-300">微信号、QQ、抖音… 都可以记录（选填）</p>
          ) : (
            <div className="flex flex-col gap-2">
              {accounts.map((row, i) => (
                <div key={i} className="flex gap-2">
                  <select
                    className="fb-input w-32 shrink-0"
                    value={row.platform}
                    onChange={(e) => {
                      const next = [...accounts]
                      next[i] = { ...row, platform: e.target.value }
                      setAccounts(next)
                    }}
                  >
                    {Object.entries(PLATFORM_META).map(([k, v]) => (
                      <option key={k} value={k}>{v.emoji} {v.label}</option>
                    ))}
                  </select>
                  <input
                    className="fb-input flex-1"
                    placeholder="账号"
                    value={row.account}
                    onChange={(e) => {
                      const next = [...accounts]
                      next[i] = { ...row, account: e.target.value }
                      setAccounts(next)
                    }}
                  />
                  <input
                    className="fb-input w-28 shrink-0"
                    placeholder="备注"
                    value={row.remark}
                    onChange={(e) => {
                      const next = [...accounts]
                      next[i] = { ...row, remark: e.target.value }
                      setAccounts(next)
                    }}
                  />
                  <button
                    type="button"
                    className="shrink-0 w-10 rounded-xl border border-ink-600/60 text-fog-500 hover:text-rose-400 hover:border-rose-500/40 transition-colors"
                    onClick={() => setAccounts(accounts.filter((_, j) => j !== i))}
                    aria-label="删除账号"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* 分组 / 标签 chips */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="fb-label">分组</label>
            <div className="flex flex-wrap gap-1.5">
              {groups.length === 0 && <span className="text-xs text-fog-300">暂无分组</span>}
              {groups.map((g) => (
                <button
                  type="button"
                  key={g.id}
                  onClick={() => toggleId(groupIds, setGroupIds, g.id)}
                  className={`px-2.5 py-1 rounded-lg text-xs border transition-all ${
                    groupIds.includes(g.id)
                      ? 'bg-brand-500/20 border-brand-500/50 text-brand-400'
                      : 'bg-ink-700/40 border-ink-600/50 text-fog-500 hover:border-fog-300/30'
                  }`}
                >
                  {g.name}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="fb-label">标签</label>
            <div className="flex flex-wrap gap-1.5">
              {tags.length === 0 && <span className="text-xs text-fog-300">暂无标签</span>}
              {tags.map((t) => (
                <button
                  type="button"
                  key={t.id}
                  onClick={() => toggleId(tagIds, setTagIds, t.id)}
                  className={`px-2.5 py-1 rounded-lg text-xs border transition-all ${
                    tagIds.includes(t.id)
                      ? 'bg-brand-500/20 border-brand-500/50 text-brand-400'
                      : 'bg-ink-700/40 border-ink-600/50 text-fog-500 hover:border-fog-300/30'
                  }`}
                >
                  #{t.name}
                </button>
              ))}
            </div>
          </div>
        </div>
      </form>
    </Modal>
  )
}
