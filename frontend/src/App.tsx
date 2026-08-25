/**
 * 应用根组件：路由表 + 鉴权守卫
 * - /login 无需鉴权
 * - 其余路由套 RequireAuth + AppLayout
 */
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { authSnapshot } from '@/stores/auth'
import AppLayout from '@/components/layout/AppLayout'
import ToastHost from '@/components/ToastHost'
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import Friends from '@/pages/Friends'
import FriendDetail from '@/pages/FriendDetail'
import Interactions from '@/pages/Interactions'
import Reminders from '@/pages/Reminders'
import Stats from '@/pages/Stats'
import AIChat from '@/pages/AIChat'
import Profile from '@/pages/Profile'
import type { ReactNode } from 'react'

/** 鉴权守卫：未登录跳 /login（记录来源页，登录后回跳） */
function RequireAuth({ children }: { children: ReactNode }) {
  const location = useLocation()
  if (!authSnapshot().isAuthenticated()) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <>
      <Routes>
        {/* 公开路由 */}
        <Route path="/login" element={<Login />} />

        {/* 受保护路由（统一布局） */}
        <Route
          element={
            <RequireAuth>
              <AppLayout />
            </RequireAuth>
          }
        >
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/friends" element={<Friends />} />
          <Route path="/friends/:id" element={<FriendDetail />} />
          <Route path="/interactions" element={<Interactions />} />
          <Route path="/reminders" element={<Reminders />} />
          <Route path="/stats" element={<Stats />} />
          <Route path="/ai" element={<AIChat />} />
          <Route path="/profile" element={<Profile />} />
        </Route>

        {/* 兜底：跳仪表盘 */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>

      {/* 全局 Toast */}
      <ToastHost />
    </>
  )
}
