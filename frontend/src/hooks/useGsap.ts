/**
 * GSAP 动画 Hooks
 * - usePageEnter: 页面进入动效（淡入 + 上移），自动清理
 * - useCountUp:   数字滚动（仪表盘统计卡用）
 * - prefersReducedMotion: 尊重系统"减少动态效果"偏好
 */
import { useEffect, useLayoutEffect, useRef, type RefObject } from 'react'
import gsap from 'gsap'

/** 系统是否偏好减少动效 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * 页面进入动画：容器整体淡入上移，子元素 [data-stagger] 依次入场。
 * 用法：给页面根元素绑定 ref 并加 class="page-enter"
 */
export function usePageEnter<T extends HTMLElement>(deps: unknown[] = []): RefObject<T> {
  const ref = useRef<T>(null)

  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    if (prefersReducedMotion()) {
      gsap.set(el, { opacity: 1, y: 0 })
      return
    }
    const ctx = gsap.context(() => {
      const tl = gsap.timeline({ defaults: { ease: 'power3.out' } })
      tl.to(el, { opacity: 1, y: 0, duration: 0.45 })
      const items = el.querySelectorAll('[data-stagger]')
      if (items.length > 0) {
        tl.from(items, { opacity: 0, y: 16, duration: 0.4, stagger: 0.06 }, '-=0.2')
      }
    }, ref)
    return () => ctx.revert()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)

  return ref
}

/**
 * 数字滚动动画：从 0（或上一值）滚动到目标值。
 * 返回文本节点 ref（支持 HTML 与 SVG 文本节点），每次 target 变化重新播放。
 */
export function useCountUp<T extends HTMLElement | SVGElement = HTMLSpanElement>(
  target: number,
  decimals = 0,
): RefObject<T> {
  const ref = useRef<T>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    if (prefersReducedMotion()) {
      el.textContent = target.toFixed(decimals)
      return
    }
    const state = { v: 0 }
    const tween = gsap.to(state, {
      v: target,
      duration: 1.1,
      ease: 'power2.out',
      onUpdate: () => {
        el.textContent = state.v.toFixed(decimals)
      },
    })
    return () => {
      tween.kill()
    }
  }, [target, decimals])

  return ref
}

/**
 * 列表交错入场：对容器内每个直接子元素做 stagger 入场。
 * deps 变化（如数据刷新）时重播。
 */
export function useStaggerIn<T extends HTMLElement>(deps: unknown[] = []): RefObject<T> {
  const ref = useRef<T>(null)

  useLayoutEffect(() => {
    if (!ref.current) return
    const children = Array.from(ref.current.children)
    if (children.length === 0) return
    if (prefersReducedMotion()) {
      gsap.set(children, { opacity: 1, y: 0 })
      return
    }
    const ctx = gsap.context(() => {
      gsap.fromTo(
        children,
        { opacity: 0, y: 18, scale: 0.98 },
        { opacity: 1, y: 0, scale: 1, duration: 0.45, stagger: 0.05, ease: 'power3.out' },
      )
    })
    return () => ctx.revert()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)

  return ref
}
