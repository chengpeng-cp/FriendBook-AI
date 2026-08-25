/**
 * 模态框组件
 * - 打开：背景淡入 + 面板缩放弹入（GSAP）
 * - 关闭：反向动画后卸载（由组件内部管理 closing 状态）
 * - Esc 关闭 / 点击遮罩关闭 / 锁定 body 滚动
 */
import { useEffect, useRef, useState, type ReactNode } from 'react'
import gsap from 'gsap'
import { prefersReducedMotion } from '@/hooks/useGsap'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  /** 顶部说明文字 */
  subtitle?: string
  children: ReactNode
  /** 底部操作区（按钮等） */
  footer?: ReactNode
  width?: string
}

export default function Modal({ open, onClose, title, subtitle, children, footer, width = 'max-w-lg' }: ModalProps) {
  const [mounted, setMounted] = useState(open) // 动画期间保持挂载
  const backdropRef = useRef<HTMLDivElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)

  // open → mounted 状态同步 + 入场动画
  useEffect(() => {
    if (open) {
      setMounted(true)
      return
    }
    // 关闭：先播出场动画再卸载
    if (!mounted) return
    if (prefersReducedMotion()) {
      setMounted(false)
      return
    }
    const tl = gsap.timeline({
      defaults: { ease: 'power2.in' },
      onComplete: () => setMounted(false),
    })
    tl.to(panelRef.current, { opacity: 0, scale: 0.96, y: 8, duration: 0.18 })
    tl.to(backdropRef.current, { opacity: 0, duration: 0.18 }, '<')
    return () => {
      tl.kill()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  // 入场动画（mounted 且 open 时触发）
  useEffect(() => {
    if (!mounted || !open) return
    if (prefersReducedMotion()) return
    const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })
    tl.fromTo(backdropRef.current, { opacity: 0 }, { opacity: 1, duration: 0.2 })
    tl.fromTo(
      panelRef.current,
      { opacity: 0, scale: 0.95, y: 14 },
      { opacity: 1, scale: 1, y: 0, duration: 0.28 },
      '-=0.1',
    )
  }, [mounted, open])

  // Esc 关闭 + body 滚动锁定
  useEffect(() => {
    if (!mounted) return
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    document.addEventListener('keydown', onKey)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  }, [mounted, onClose])

  if (!mounted) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* 遮罩 */}
      <div
        ref={backdropRef}
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      {/* 面板 */}
      <div ref={panelRef} className={`relative w-full ${width} fb-card bg-ink-800 overflow-hidden`}>
        {/* 标题栏 */}
        <div className="flex items-start justify-between px-6 pt-5 pb-4 border-b border-ink-700/60">
          <div>
            <h3 className="text-base font-semibold text-fog-900">{title}</h3>
            {subtitle && <p className="mt-0.5 text-xs text-fog-500">{subtitle}</p>}
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-fog-500 hover:text-fog-900 hover:bg-ink-700/60 transition-colors"
            aria-label="关闭"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
            </svg>
          </button>
        </div>
        {/* 内容区 */}
        <div className="px-6 py-5 max-h-[65vh] overflow-y-auto">{children}</div>
        {footer && (
          <div className="flex justify-end gap-3 px-6 py-4 border-t border-ink-700/60 bg-ink-900/40">
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
