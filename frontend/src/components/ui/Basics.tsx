/**
 * 基础展示组件：加载态 / 空状态 / 骨架屏 / 温度徽章
 */
import type { ReactNode } from 'react'
import { TEMP_META } from '@/constants'
import type { Temperature } from '@/types'

/** 加载指示器 */
export function Spinner({ size = 20, className = '' }: { size?: number; className?: string }) {
  return (
    <svg
      className={`animate-spin text-brand-400 ${className}`}
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
    >
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" className="opacity-20" />
      <path d="M22 12a10 10 0 0 0-10-10" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  )
}

/** 页面级加载（居中大 spinner + 文案） */
export function PageLoading({ text = '加载中…' }: { text?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-24 gap-3 text-fog-500">
      <Spinner size={28} />
      <span className="text-sm">{text}</span>
    </div>
  )
}

/** 空状态 */
export function Empty({ icon = '🌿', title, hint }: { icon?: string; title: string; hint?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-2 text-center">
      <span className="text-4xl mb-1 opacity-70">{icon}</span>
      <p className="text-sm font-medium text-fog-700">{title}</p>
      {hint && <p className="text-xs text-fog-500">{hint}</p>}
    </div>
  )
}

/** 骨架屏块 */
export function Skeleton({ className = 'h-4 w-full' }: { className?: string }) {
  return <div className={`fb-skeleton ${className}`} />
}

/** 温度徽章（含 emoji 与颜色） */
export function TempBadge({ temp, size = 'md' }: { temp: Temperature; size?: 'sm' | 'md' }) {
  const meta = TEMP_META[temp]
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full font-semibold whitespace-nowrap ${
        size === 'sm' ? 'text-[10px] px-2 py-0.5' : 'text-xs px-2.5 py-1'
      }`}
      style={{
        color: meta.color,
        backgroundColor: `${meta.color}1f`, // 12% 透明度底色
        border: `1px solid ${meta.color}44`,
      }}
      title={meta.desc}
    >
      <span>{meta.emoji}</span>
      <span>{meta.label}</span>
    </span>
  )
}

/** 好感度进度条（颜色随温度档位变化） */
export function FavorBar({ favor, animate = true }: { favor: number; animate?: boolean }) {
  const meta = TEMP_META[favor >= 90 ? 1 : favor >= 75 ? 2 : favor >= 60 ? 3 : favor >= 40 ? 4 : 5]
  return (
    <div className="h-1.5 w-full rounded-full bg-ink-700/60 overflow-hidden">
      <div
        className="h-full rounded-full transition-all duration-700 ease-out"
        style={{
          width: animate ? `${favor}%` : 0,
          background: `linear-gradient(90deg, ${meta.color}88, ${meta.color})`,
        }}
      />
    </div>
  )
}

/** 头像（无图时按 ID 生成渐变首字符） */
export function Avatar({
  id,
  name,
  size = 40,
}: {
  id: number
  name: string
  size?: number
}) {
  const gradients = [
    'from-orange-400 to-rose-500',
    'from-amber-400 to-orange-600',
    'from-sky-400 to-indigo-500',
    'from-emerald-400 to-teal-600',
    'from-fuchsia-400 to-purple-600',
    'from-cyan-400 to-blue-500',
  ]
  const initial = (name.trim()[0] ?? '友').toUpperCase()
  return (
    <div
      className={`flex items-center justify-center rounded-full bg-gradient-to-br ${
        gradients[id % gradients.length]
      } font-semibold text-white select-none shrink-0`}
      style={{ width: size, height: size, fontSize: size * 0.42 }}
    >
      {initial}
    </div>
  )
}

/** 区块标题 */
export function SectionTitle({ children, action }: { children: ReactNode; action?: ReactNode }) {
  return (
    <div className="flex items-center justify-between mb-4">
      <h2 className="text-sm font-semibold text-fog-900 tracking-wide">{children}</h2>
      {action}
    </div>
  )
}
