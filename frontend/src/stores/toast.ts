/**
 * 轻量 Toast 通知（全局 store，无需 Provider）
 * 类型：success | error | info
 */
import { create } from 'zustand'

export type ToastType = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

interface ToastState {
  toasts: ToastItem[]
  push: (type: ToastType, message: string) => void
  remove: (id: number) => void
}

let nextId = 1

export const useToastStore = create<ToastState>((set) => ({
  toasts: [],
  push: (type, message) => {
    const id = nextId++
    set((s) => ({ toasts: [...s.toasts, { id, type, message }] }))
    // 3 秒自动消失
    setTimeout(() => {
      set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
    }, 3000)
  },
  remove: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}))

/** 函数式便捷调用（组件外也可用） */
export const toast = {
  success: (msg: string) => useToastStore.getState().push('success', msg),
  error: (msg: string) => useToastStore.getState().push('error', msg),
  info: (msg: string) => useToastStore.getState().push('info', msg),
}
