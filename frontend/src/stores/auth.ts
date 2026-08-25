/**
 * 认证状态（zustand + persist）
 * token/user 持久化到 localStorage，与 axios 拦截器读取的 key 保持一致
 */
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  expiresAt: number | null
  user: User | null
  /** 登录成功写入 */
  setAuth: (token: string, expiresAt: number) => void
  /** 缓存用户信息 */
  setUser: (user: User | null) => void
  /** 退出登录 */
  logout: () => void
  /** 是否已登录（token 存在且未过期） */
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      expiresAt: null,
      user: null,

      setAuth: (token, expiresAt) => set({ token, expiresAt }),
      setUser: (user) => set({ user }),
      logout: () => {
        set({ token: null, expiresAt: null, user: null })
        // axios 拦截器读的是 localStorage 原始 key，这里同步清理
        localStorage.removeItem('fb_token')
      },

      isAuthenticated: () => {
        const { token, expiresAt } = get()
        if (!token) return false
        if (expiresAt && expiresAt * 1000 < Date.now()) return false
        return true
      },
    }),
    {
      name: 'fb_auth',
      // 只持久化基础字段；同时冗余写一份 fb_token 供 axios/SSE 拦截器快速读取
      partialize: (s) => ({ token: s.token, expiresAt: s.expiresAt, user: s.user }),
      onRehydrateStorage: () => (state) => {
        if (state?.token) localStorage.setItem('fb_token', state.token)
      },
    },
  ),
)

/** 便捷读取（非 hook 场景，如路由守卫） */
export const authSnapshot = () => useAuthStore.getState()
