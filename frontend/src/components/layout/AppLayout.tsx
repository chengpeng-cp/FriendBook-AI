/**
 * 应用主布局：左侧导航 + 顶栏 + 内容区
 * - 侧边栏当前项高亮（渐变指示条）
 * - 顶栏展示当前用户，支持快捷退出
 * - Outlet 渲染子路由页面
 */
import { useEffect } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { Avatar } from '@/components/ui/Basics'

/** 导航配置（集中管理，便于扩展） */
const NAV_ITEMS = [
  { to: '/dashboard', icon: '◎', label: '仪表盘', hint: '社交圈总览' },
  { to: '/friends', icon: '👥', label: '好友', hint: '关系管理' },
  { to: '/interactions', icon: '✦', label: '互动', hint: '点滴记录' },
  { to: '/reminders', icon: '⏰', label: '提醒', hint: '生日纪念日' },
  { to: '/stats', icon: '▤', label: '统计', hint: '数据分析' },
  { to: '/ai', icon: '✨', label: 'AI 助手', hint: '对话管理' },
]

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const logout = useAuthStore((s) => s.logout)

  // 首次进入拉取用户信息（本地缓存的 user 可能已过期）
  useEffect(() => {
    authApi
      .profile()
      .then(setUser)
      .catch(() => {
        /* 拦截器已处理 401 跳转 */
      })
  }, [setUser])

  // 退出登录
  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="min-h-screen flex bg-ink-950">
      {/* ===== 侧边栏 ===== */}
      <aside className="w-60 shrink-0 flex flex-col border-r border-ink-700/50 bg-ink-900/60 backdrop-blur-sm sticky top-0 h-screen">
        {/* Logo */}
        <div className="flex items-center gap-2.5 px-5 h-16 border-b border-ink-700/50">
          <div className="h-8 w-8 rounded-xl bg-gradient-to-br from-brand-400 to-brand-600 flex items-center justify-center text-white text-sm font-bold shadow-glow">
            FB
          </div>
          <div className="leading-tight">
            <div className="text-sm font-bold tracking-wide">FriendBook</div>
            <div className="text-[10px] text-fog-500 tracking-widest">AI · 智能好友管理</div>
          </div>
        </div>

        {/* 导航 */}
        <nav className="flex-1 px-3 py-4 flex flex-col gap-1 overflow-y-auto">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `group relative flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 ${
                  isActive
                    ? 'bg-gradient-to-r from-brand-500/15 to-transparent text-fog-900 font-medium'
                    : 'text-fog-500 hover:text-fog-900 hover:bg-ink-700/40'
                }`
              }
            >
              {({ isActive }) => (
                <>
                  {/* 当前项指示条 */}
                  <span
                    className={`absolute left-0 top-1/2 -translate-y-1/2 h-5 w-1 rounded-r-full bg-gradient-to-b from-brand-400 to-brand-600 transition-all duration-200 ${
                      isActive ? 'opacity-100 scale-y-100' : 'opacity-0 scale-y-0'
                    }`}
                  />
                  <span className="text-base w-5 text-center">{item.icon}</span>
                  <span className="flex-1">{item.label}</span>
                  <span className="text-[10px] text-fog-300 opacity-0 group-hover:opacity-100 transition-opacity">
                    {item.hint}
                  </span>
                </>
              )}
            </NavLink>
          ))}
        </nav>

        {/* 底部用户卡片 */}
        <div className="p-3 border-t border-ink-700/50">
          <NavLink
            to="/profile"
            className={`flex items-center gap-3 p-2.5 rounded-xl transition-colors ${
              location.pathname === '/profile' ? 'bg-ink-700/50' : 'hover:bg-ink-700/40'
            }`}
          >
            <Avatar id={user?.id ?? 0} name={user?.nickname || user?.username || '友'} size={34} />
            <div className="flex-1 min-w-0">
              <div className="text-xs font-medium truncate">{user?.nickname || user?.username}</div>
              <div className="text-[10px] text-fog-500 truncate">@{user?.username}</div>
            </div>
          </NavLink>
          <button
            onClick={handleLogout}
            className="mt-1 w-full h-8 rounded-lg text-xs text-fog-500 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
          >
            退出登录
          </button>
        </div>
      </aside>

      {/* ===== 内容区 ===== */}
      <main className="flex-1 min-w-0 flex flex-col">
        {/* 顶栏：当前页面标识（依据路由自动推导） */}
        <header className="h-16 shrink-0 flex items-center justify-between px-8 border-b border-ink-700/50 bg-ink-950/80 backdrop-blur-sm sticky top-0 z-10">
          <div className="text-sm text-fog-500">
            {NAV_ITEMS.find((n) => location.pathname.startsWith(n.to))?.hint ?? '个人中心'} · FriendBook AI
          </div>
          <div className="text-xs text-fog-300">
            让每一段关系，都被认真对待
          </div>
        </header>

        {/* 页面容器 */}
        <div className="flex-1 p-8 overflow-x-hidden">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
