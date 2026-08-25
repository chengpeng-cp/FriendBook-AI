/**
 * 登录 / 注册页
 * 左侧品牌区：渐变光斑漂浮（GSAP 循环动画）+ 温度光谱元素
 * 右侧表单区：登录/注册切换，成功后 GSAP 过场跳转
 */
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import gsap from 'gsap'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/stores/toast'
import { prefersReducedMotion } from '@/hooks/useGsap'
import { TEMP_META } from '@/constants'

type Mode = 'login' | 'register'

export default function Login() {
  const navigate = useNavigate()
  const location = useLocation()
  const setAuth = useAuthStore((s) => s.setAuth)

  const [mode, setMode] = useState<Mode>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [nickname, setNickname] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const brandRef = useRef<HTMLDivElement>(null)
  const formRef = useRef<HTMLDivElement>(null)

  /** 品牌区入场 + 光斑无限漂浮 */
  useEffect(() => {
    if (prefersReducedMotion()) return
    const ctx = gsap.context(() => {
      // 光斑漂浮（低振幅正弦感）
      gsap.to('[data-orb="1"]', {
        y: 'random(-30, 30)',
        x: 'random(-20, 20)',
        duration: 'random(4, 6)',
        repeat: -1,
        yoyo: true,
        ease: 'sine.inOut',
      })
      gsap.to('[data-orb="2"]', {
        y: 'random(-40, 20)',
        x: 'random(-30, 30)',
        duration: 'random(5, 8)',
        repeat: -1,
        yoyo: true,
        ease: 'sine.inOut',
        delay: 0.5,
      })
      // 温度徽标交错浮动
      gsap.to('[data-temp-chip]', {
        y: -8,
        duration: 1.8,
        repeat: -1,
        yoyo: true,
        ease: 'sine.inOut',
        stagger: 0.25,
      })
      // 表单入场
      gsap.fromTo(
        formRef.current,
        { opacity: 0, x: 30 },
        { opacity: 1, x: 0, duration: 0.6, ease: 'power3.out', delay: 0.15 },
      )
    }, brandRef)
    return () => ctx.revert()
  }, [])

  /** 切换登录/注册：表单轻微过场 */
  const switchMode = (next: Mode) => {
    if (next === mode) return
    setMode(next)
    setError('')
    if (!prefersReducedMotion() && formRef.current) {
      gsap.fromTo(
        formRef.current,
        { opacity: 0, y: 10 },
        { opacity: 1, y: 0, duration: 0.3, ease: 'power2.out' },
      )
    }
  }

  /** 提交（登录 / 注册→自动登录） */
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')

    if (username.trim().length < 3) return setError('用户名至少 3 位')
    if (password.length < 6) return setError('密码至少 6 位')

    setLoading(true)
    try {
      if (mode === 'register') {
        await authApi.register({ username: username.trim(), password, nickname: nickname.trim() })
        toast.success('注册成功，自动登录中')
      }
      const out = await authApi.login({ username: username.trim(), password })
      setAuth(out.token, out.expires_at)
      toast.success(mode === 'login' ? '欢迎回来' : '欢迎加入 FriendBook')

      // 过场动画后跳转（来源页或仪表盘）
      const from = (location.state as { from?: string })?.from ?? '/dashboard'
      if (prefersReducedMotion()) {
        navigate(from, { replace: true })
      } else {
        gsap.to(formRef.current, {
          opacity: 0,
          x: -24,
          duration: 0.3,
          ease: 'power2.in',
          onComplete: () => navigate(from, { replace: true }),
        })
      }
    } catch (err) {
      setError((err as Error).message || '操作失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex bg-ink-950">
      {/* ===== 左：品牌区（小屏隐藏） ===== */}
      <div ref={brandRef} className="relative hidden lg:flex w-[46%] flex-col justify-between p-12 overflow-hidden border-r border-ink-700/40">
        {/* 漂浮光斑 */}
        <div data-orb="1" className="absolute -top-24 -left-24 h-96 w-96 rounded-full bg-brand-500/20 blur-3xl" />
        <div data-orb="2" className="absolute bottom-0 left-1/3 h-80 w-80 rounded-full bg-sky-500/10 blur-3xl" />

        {/* Slogan */}
        <div className="relative">
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-xl bg-gradient-to-br from-brand-400 to-brand-600 flex items-center justify-center text-white font-bold shadow-glow">
              FB
            </div>
            <span className="text-lg font-bold tracking-wide">FriendBook AI</span>
          </div>
        </div>

        <div className="relative">
          <h1 className="text-4xl font-bold leading-snug">
            让每一段关系
            <br />
            都被
            <span className="bg-gradient-to-r from-brand-400 to-rose-400 bg-clip-text text-transparent"> 认真对待</span>
          </h1>
          <p className="mt-5 text-sm leading-relaxed text-fog-500 max-w-sm">
            好感度、关系温度、多平台账号、互动点滴 ——
            用 AI 助手管理你的社交圈，从「夯」到「拉」，一切关系尽在掌握。
          </p>

          {/* 温度光谱展示 */}
          <div className="mt-8 flex gap-3">
            {[1, 2, 3, 4, 5].map((t) => (
              <div
                key={t}
                data-temp-chip
                className="flex flex-col items-center gap-1.5 px-4 py-3 rounded-2xl border"
                style={{
                  borderColor: `${TEMP_META[t as 1].color}33`,
                  background: `${TEMP_META[t as 1].color}0d`,
                }}
              >
                <span className="text-xl">{TEMP_META[t as 1].emoji}</span>
                <span className="text-xs font-semibold" style={{ color: TEMP_META[t as 1].color }}>
                  {TEMP_META[t as 1].label}
                </span>
              </div>
            ))}
          </div>
        </div>

        <p className="relative text-xs text-fog-300">Gin + GORM + Eino · 前后端分离教学项目</p>
      </div>

      {/* ===== 右：表单区 ===== */}
      <div className="flex-1 flex items-center justify-center p-6">
        <div ref={formRef} className="w-full max-w-sm">
          <h2 className="text-2xl font-bold">{mode === 'login' ? '欢迎回来' : '创建账号'}</h2>
          <p className="mt-1.5 text-sm text-fog-500">
            {mode === 'login' ? '登录你的好友管理系统' : '注册后开启智能好友管理之旅'}
          </p>

          {/* 模式切换 */}
          <div className="mt-7 flex p-1 rounded-xl bg-ink-800 border border-ink-700/60 text-sm">
            {(['login', 'register'] as Mode[]).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => switchMode(m)}
                className={`flex-1 h-9 rounded-lg transition-all duration-200 ${
                  mode === m
                    ? 'bg-gradient-to-r from-brand-500 to-brand-600 text-white font-medium shadow-glow'
                    : 'text-fog-500 hover:text-fog-900'
                }`}
              >
                {m === 'login' ? '登录' : '注册'}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4" noValidate>
            <div>
              <label className="fb-label" htmlFor="username">用户名</label>
              <input
                id="username"
                className="fb-input"
                placeholder="3-64 位字符"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>

            {mode === 'register' && (
              <div>
                <label className="fb-label" htmlFor="nickname">昵称（可选）</label>
                <input
                  id="nickname"
                  className="fb-input"
                  placeholder="展示用昵称"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                />
              </div>
            )}

            <div>
              <label className="fb-label" htmlFor="password">密码</label>
              <input
                id="password"
                type="password"
                className="fb-input"
                placeholder="至少 6 位"
                autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {/* 错误提示 */}
            {error && (
              <p className="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/30 rounded-lg px-3 py-2">
                {error}
              </p>
            )}

            <button type="submit" disabled={loading} className="fb-btn-primary h-11 mt-1">
              {loading ? (
                <span className="flex items-center gap-2">
                  <span className="h-4 w-4 rounded-full border-2 border-white/30 border-t-white animate-spin" />
                  处理中…
                </span>
              ) : mode === 'login' ? (
                '登录'
              ) : (
                '注册并登录'
              )}
            </button>
          </form>

          <p className="mt-6 text-center text-xs text-fog-300">
            继续即代表同意 FriendBook 的虚构服务条款（教学项目）
          </p>
        </div>
      </div>
    </div>
  )
}
