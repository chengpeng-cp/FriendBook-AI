/**
 * 个人中心：资料编辑 + 密码修改 + 分组/标签管理
 */
import { useEffect, useState, type FormEvent } from 'react'
import { authApi } from '@/api/auth'
import { groupApi, tagApi } from '@/api/meta'
import { useAuthStore } from '@/stores/auth'
import { usePageEnter } from '@/hooks/useGsap'
import { Avatar, SectionTitle } from '@/components/ui/Basics'
import Modal from '@/components/ui/Modal'
import { toast } from '@/stores/toast'
import type { Group, Tag } from '@/types'

export default function Profile() {
  const pageRef = usePageEnter<HTMLDivElement>()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)

  // ---- 资料表单 ----
  const [nickname, setNickname] = useState('')
  const [bio, setBio] = useState('')
  const [savingProfile, setSavingProfile] = useState(false)

  // ---- 密码表单 ----
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [confirmPwd, setConfirmPwd] = useState('')
  const [savingPwd, setSavingPwd] = useState(false)

  // ---- 分组 / 标签 ----
  const [groups, setGroups] = useState<Group[]>([])
  const [tags, setTags] = useState<Tag[]>([])
  const [newGroupName, setNewGroupName] = useState('')
  const [newTagName, setNewTagName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<{ kind: 'group' | 'tag'; id: number; name: string } | null>(null)

  useEffect(() => {
    if (user) {
      setNickname(user.nickname)
      setBio(user.bio)
    }
  }, [user])

  const loadMeta = () => {
    groupApi.list().then((r) => setGroups(r.list)).catch(() => {})
    tagApi.list().then((r) => setTags(r.list)).catch(() => {})
  }
  useEffect(loadMeta, [])

  /** 保存资料 */
  const saveProfile = async (e: FormEvent) => {
    e.preventDefault()
    setSavingProfile(true)
    try {
      await authApi.updateProfile({ nickname: nickname.trim(), bio: bio.trim() })
      const fresh = await authApi.profile()
      setUser(fresh)
      toast.success('资料已更新')
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setSavingProfile(false)
    }
  }

  /** 修改密码 */
  const savePassword = async (e: FormEvent) => {
    e.preventDefault()
    if (newPwd.length < 6) return toast.error('新密码至少 6 位')
    if (newPwd !== confirmPwd) return toast.error('两次输入的新密码不一致')
    setSavingPwd(true)
    try {
      await authApi.changePassword({ old_password: oldPwd, new_password: newPwd })
      toast.success('密码修改成功，下次登录请使用新密码')
      setOldPwd('')
      setNewPwd('')
      setConfirmPwd('')
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setSavingPwd(false)
    }
  }

  return (
    <div ref={pageRef} className="page-enter max-w-3xl mx-auto">
      <h1 className="text-xl font-bold mb-7">个人中心</h1>

      <div className="flex flex-col gap-5">
        {/* ===== 基本资料 ===== */}
        <form onSubmit={saveProfile} className="fb-card p-6" data-stagger>
          <SectionTitle>基本资料</SectionTitle>
          <div className="flex items-center gap-4 mb-5">
            <Avatar id={user?.id ?? 0} name={nickname || user?.username || '友'} size={56} />
            <div className="text-xs text-fog-500">
              <p className="text-sm text-fog-900 font-medium">{user?.username}</p>
              <p className="mt-0.5">加入于 {user?.created_at?.slice(0, 10) ?? '—'}</p>
            </div>
          </div>
          <div className="flex flex-col gap-4">
            <div>
              <label className="fb-label">昵称</label>
              <input className="fb-input" value={nickname} onChange={(e) => setNickname(e.target.value)} maxLength={64} />
            </div>
            <div>
              <label className="fb-label">个人简介</label>
              <textarea
                className="fb-input !h-20 py-2.5 resize-none"
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                maxLength={255}
                placeholder="一句话介绍自己"
              />
            </div>
            <div className="flex justify-end">
              <button type="submit" className="fb-btn-primary" disabled={savingProfile}>
                {savingProfile ? '保存中…' : '保存资料'}
              </button>
            </div>
          </div>
        </form>

        {/* ===== 分组管理 ===== */}
        <div className="fb-card p-6" data-stagger>
          <SectionTitle>好友分组（{groups.length}）</SectionTitle>
          <div className="flex flex-wrap gap-2 mb-4">
            {groups.length === 0 && <p className="text-xs text-fog-300">暂无分组，如：同事 / 大学同学 / 家人</p>}
            {groups.map((g) => (
              <span
                key={g.id}
                className="group inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs bg-brand-500/10 border border-brand-500/30 text-brand-400"
              >
                {g.name}
                <button
                  className="opacity-40 group-hover:opacity-100 hover:text-rose-400 transition-all"
                  onClick={() => setDeleteTarget({ kind: 'group', id: g.id, name: g.name })}
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
          <div className="flex gap-2">
            <input
              className="fb-input flex-1"
              placeholder="新分组名称"
              value={newGroupName}
              onChange={(e) => setNewGroupName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && newGroupName.trim()) {
                  e.preventDefault()
                  groupApi
                    .create(newGroupName.trim())
                    .then(() => {
                      toast.success('分组已创建')
                      setNewGroupName('')
                      loadMeta()
                    })
                    .catch((err) => toast.error((err as Error).message))
                }
              }}
            />
            <button
              className="fb-btn-ghost"
              onClick={() => {
                if (!newGroupName.trim()) return
                groupApi
                  .create(newGroupName.trim())
                  .then(() => {
                    toast.success('分组已创建')
                    setNewGroupName('')
                    loadMeta()
                  })
                  .catch((err) => toast.error((err as Error).message))
              }}
            >
              添加
            </button>
          </div>
        </div>

        {/* ===== 标签管理 ===== */}
        <div className="fb-card p-6" data-stagger>
          <SectionTitle>好友标签（{tags.length}）</SectionTitle>
          <div className="flex flex-wrap gap-2 mb-4">
            {tags.length === 0 && <p className="text-xs text-fog-300">暂无标签，如：#铁哥们 #大佬 #酒友</p>}
            {tags.map((t) => (
              <span
                key={t.id}
                className="group inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs bg-violet-500/10 border border-violet-500/30 text-violet-300"
              >
                #{t.name}
                <button
                  className="opacity-40 group-hover:opacity-100 hover:text-rose-400 transition-all"
                  onClick={() => setDeleteTarget({ kind: 'tag', id: t.id, name: t.name })}
                >
                  ✕
                </button>
              </span>
            ))}
          </div>
          <div className="flex gap-2">
            <input
              className="fb-input flex-1"
              placeholder="新标签名称"
              value={newTagName}
              onChange={(e) => setNewTagName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && newTagName.trim()) {
                  e.preventDefault()
                  tagApi
                    .create(newTagName.trim())
                    .then(() => {
                      toast.success('标签已创建')
                      setNewTagName('')
                      loadMeta()
                    })
                    .catch((err) => toast.error((err as Error).message))
                }
              }}
            />
            <button
              className="fb-btn-ghost"
              onClick={() => {
                if (!newTagName.trim()) return
                tagApi
                  .create(newTagName.trim())
                  .then(() => {
                    toast.success('标签已创建')
                    setNewTagName('')
                    loadMeta()
                  })
                  .catch((err) => toast.error((err as Error).message))
              }}
            >
              添加
            </button>
          </div>
        </div>

        {/* ===== 修改密码 ===== */}
        <form onSubmit={savePassword} className="fb-card p-6" data-stagger>
          <SectionTitle>修改密码</SectionTitle>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="fb-label">旧密码</label>
              <input type="password" className="fb-input" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} autoComplete="current-password" />
            </div>
            <div>
              <label className="fb-label">新密码</label>
              <input type="password" className="fb-input" value={newPwd} onChange={(e) => setNewPwd(e.target.value)} autoComplete="new-password" />
            </div>
            <div>
              <label className="fb-label">确认新密码</label>
              <input type="password" className="fb-input" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)} autoComplete="new-password" />
            </div>
          </div>
          <div className="flex justify-end mt-4">
            <button type="submit" className="fb-btn-primary" disabled={savingPwd || !oldPwd || !newPwd}>
              {savingPwd ? '提交中…' : '修改密码'}
            </button>
          </div>
        </form>
      </div>

      {/* 删除确认弹窗 */}
      <Modal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title={`删除${deleteTarget?.kind === 'group' ? '分组' : '标签'}`}
        width="max-w-sm"
        footer={
          <>
            <button className="fb-btn-ghost" onClick={() => setDeleteTarget(null)}>再想想</button>
            <button
              className="fb-btn-danger"
              onClick={async () => {
                if (!deleteTarget) return
                try {
                  const fn = deleteTarget.kind === 'group' ? groupApi.remove : tagApi.remove
                  await fn(deleteTarget.id)
                  toast.success('已删除')
                  setDeleteTarget(null)
                  loadMeta()
                } catch (e) {
                  toast.error((e as Error).message)
                }
              }}
            >
              确认删除
            </button>
          </>
        }
      >
        <p className="text-sm text-fog-700">
          确定删除{deleteTarget?.kind === 'group' ? '分组' : '标签'}「{deleteTarget?.name}」吗？
          <span className="block mt-1 text-xs text-fog-500">好友与该分组的关联关系将被解除。</span>
        </p>
      </Modal>
    </div>
  )
}
