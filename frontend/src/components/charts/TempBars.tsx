/**
 * 温度分布横向条形图（纯 SVG + GSAP 生长动画）
 * 数据：夯/热/温/凉/拉 五档好友数量
 */
import { useLayoutEffect, useRef } from 'react'
import gsap from 'gsap'
import { TEMP_META } from '@/constants'
import { prefersReducedMotion } from '@/hooks/useGsap'
import type { TemperatureDist } from '@/types'

const ORDER: { key: keyof TemperatureDist; temp: 1 | 2 | 3 | 4 | 5 }[] = [
  { key: 'hot', temp: 1 },
  { key: 'warm', temp: 2 },
  { key: 'mild', temp: 3 },
  { key: 'cool', temp: 4 },
  { key: 'cold', temp: 5 },
]

export default function TempBars({ data }: { data: TemperatureDist | null }) {
  const wrapRef = useRef<HTMLDivElement>(null)

  const rows = ORDER.map(({ key, temp }) => ({
    temp,
    meta: TEMP_META[temp],
    count: data?.[key] ?? 0,
  }))
  const max = Math.max(1, ...rows.map((r) => r.count))

  // 条形生长动画：宽度从 0 → 目标比例
  useLayoutEffect(() => {
    if (!wrapRef.current || prefersReducedMotion()) return
    const ctx = gsap.context(() => {
      gsap.fromTo(
        wrapRef.current!.querySelectorAll('[data-bar]'),
        { scaleX: 0 },
        { scaleX: 1, duration: 0.8, stagger: 0.08, ease: 'power3.out' },
      )
    })
    return () => ctx.revert()
  }, [data])

  return (
    <div ref={wrapRef} className="flex flex-col gap-3">
      {rows.map(({ temp, meta, count }) => (
        <div key={temp} className="flex items-center gap-3">
          {/* 档位标签 */}
          <span className="w-14 shrink-0 text-xs font-medium" style={{ color: meta.color }}>
            {meta.emoji} {meta.label}
          </span>
          {/* 条形轨道 */}
          <div className="flex-1 h-5 rounded-md bg-ink-900/80 overflow-hidden">
            <div
              data-bar
              className="h-full rounded-md origin-left"
              style={{
                width: `${Math.max(2, (count / max) * 100)}%`,
                background: `linear-gradient(90deg, ${meta.color}55, ${meta.color})`,
              }}
            />
          </div>
          {/* 数量 */}
          <span className="w-8 shrink-0 text-right text-xs text-fog-700 font-medium tabular-nums">
            {count}
          </span>
        </div>
      ))}
    </div>
  )
}
