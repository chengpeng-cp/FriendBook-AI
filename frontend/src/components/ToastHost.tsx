/**
 * Toast 通知宿主（挂载于 App 根部）
 * GSAP 右侧滑入 + 自动淡出（由 store 控制移除时机）
 */
import { useEffect, useRef } from 'react'
import gsap from 'gsap'
import { useToastStore, type ToastItem } from '@/stores/toast'
import { prefersReducedMotion } from '@/hooks/useGsap'

const ICONS = {
  success: '✓',
  error: '✕',
  info: 'i',
} as const

const COLORS = {
  success: 'bg-emerald-500/15 border-emerald-500/40 text-emerald-400',
  error: 'bg-rose-500/15 border-rose-500/40 text-rose-400',
  info: 'bg-sky-500/15 border-sky-500/40 text-sky-400',
} as const

function ToastRow({ item }: { item: ToastItem }) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!ref.current || prefersReducedMotion()) return
    gsap.fromTo(
      ref.current,
      { x: 60, opacity: 0 },
      { x: 0, opacity: 1, duration: 0.35, ease: 'back.out(1.6)' },
    )
  }, [])

  return (
    <div
      ref={ref}
      className={`flex items-center gap-2.5 px-4 py-3 rounded-xl border backdrop-blur-md shadow-card text-sm text-fog-900 ${COLORS[item.type]}`}
      role="status"
    >
      <span className="flex h-5 w-5 items-center justify-center rounded-full border border-current text-[10px] font-bold">
        {ICONS[item.type]}
      </span>
      <span className="max-w-64">{item.message}</span>
    </div>
  )
}

export default function ToastHost() {
  const toasts = useToastStore((s) => s.toasts)
  return (
    <div className="fixed top-5 right-5 z-[100] flex flex-col gap-2">
      {toasts.map((t) => (
        <ToastRow key={t.id} item={t} />
      ))}
    </div>
  )
}
