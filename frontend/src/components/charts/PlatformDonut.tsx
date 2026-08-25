/**
 * 平台分布环形图（SVG stroke-dasharray + GSAP 扇区展开）
 */
import { useLayoutEffect, useMemo, useRef } from 'react'
import gsap from 'gsap'
import { PLATFORM_META } from '@/constants'
import { prefersReducedMotion } from '@/hooks/useGsap'
import type { PlatformDistItem } from '@/types'

const PALETTE = ['#FF8A4C', '#63B3ED', '#9F7AEA', '#F6C453', '#F687B3', '#4FD1C5', '#8A94A3']
const SIZE = 160
const R = 62 // 环半径
const STROKE = 18 // 环厚度

export default function PlatformDonut({ data }: { data: PlatformDistItem[] | null }) {
  const svgRef = useRef<SVGSVGElement>(null)
  const list = useMemo(
    () => (data ?? []).filter((d) => d.count > 0).slice(0, 7), // 最多展示 7 个平台
    [data],
  )
  const total = list.reduce((s, d) => s + d.count, 0)

  // 每个扇区的起止角度与路径
  const segments = useMemo(() => {
    const C = 2 * Math.PI * R // 周长
    let offset = 0
    return list.map((d, i) => {
      const frac = total > 0 ? d.count / total : 0
      const seg = { ...d, frac, dash: frac * C, gap: C - frac * C, offset: -offset, color: PALETTE[i % PALETTE.length] }
      offset += frac * C
      return seg
    })
  }, [list, total])

  useLayoutEffect(() => {
    if (!svgRef.current || prefersReducedMotion() || segments.length === 0) return
    const ctx = gsap.context(() => {
      // 先全部隐藏，再按扇区依次旋转画出
      gsap.fromTo(
        svgRef.current!.querySelectorAll('[data-seg]'),
        { opacity: 0, rotate: '-30deg', transformOrigin: 'center' },
        { opacity: 1, rotate: '0deg', duration: 0.55, stagger: 0.09, ease: 'back.out(1.4)' },
      )
    })
    return () => ctx.revert()
  }, [segments])

  if (segments.length === 0) {
    return (
      <div className="flex items-center justify-center h-[200px] text-xs text-fog-500">
        暂无平台账号数据，给好友添加账号后展示
      </div>
    )
  }

  return (
    <div className="flex items-center gap-6">
      <svg ref={svgRef} width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} className="shrink-0">
        <g transform={`translate(${SIZE / 2},${SIZE / 2}) rotate(-90)`}>
          {/* 背景环 */}
          <circle r={R} fill="none" stroke="#161B26" strokeWidth={STROKE} />
          {/* 数据扇区：dasharray = [弧长, 空隙]，offset 控制起点 */}
          {segments.map((s, i) => (
            <circle
              key={i}
              data-seg
              r={R}
              fill="none"
              stroke={s.color}
              strokeWidth={STROKE}
              strokeDasharray={`${s.dash} ${s.gap}`}
              strokeDashoffset={s.offset}
              strokeLinecap="butt"
            >
              <title>{`${PLATFORM_META[s.platform]?.label ?? s.platform}：${s.count}`}</title>
            </circle>
          ))}
        </g>
        {/* 中心总数 */}
        <text x={SIZE / 2} y={SIZE / 2 - 2} textAnchor="middle" fontSize="24" fontWeight="700" fill="#E8ECF3">
          {total}
        </text>
        <text x={SIZE / 2} y={SIZE / 2 + 16} textAnchor="middle" fontSize="10" fill="#5A6478">
          账号总数
        </text>
      </svg>

      {/* 图例 */}
      <div className="flex flex-col gap-2 min-w-0">
        {segments.map((s, i) => (
          <div key={i} className="flex items-center gap-2 text-xs">
            <span className="h-2.5 w-2.5 rounded-sm shrink-0" style={{ background: s.color }} />
            <span className="text-fog-700">
              {PLATFORM_META[s.platform]?.emoji} {PLATFORM_META[s.platform]?.label ?? s.platform}
            </span>
            <span className="text-fog-500 tabular-nums ml-auto">
              {s.count} · {Math.round(s.frac * 100)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
