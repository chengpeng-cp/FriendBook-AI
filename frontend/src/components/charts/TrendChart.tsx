/**
 * 互动趋势面积图（纯 SVG + GSAP 路径描画动画）
 * - 折线 stroke-dashoffset 生长
 * - 面积淡入
 * - 数据点缩放弹出
 */
import { useLayoutEffect, useMemo, useRef } from 'react'
import gsap from 'gsap'
import { prefersReducedMotion } from '@/hooks/useGsap'
import type { TrendPoint } from '@/types'

const W = 720
const H = 200
const PAD = { top: 16, right: 12, bottom: 26, left: 30 }

export default function TrendChart({ data }: { data: TrendPoint[] | null }) {
  const svgRef = useRef<SVGSVGElement>(null)

  /** 计算折线/面积路径（Catmull-Rom → 贝塞尔平滑可选，这里用折线更直观） */
  const { linePath, areaPath, points, ticks, max } = useMemo(() => {
    const list = data ?? []
    const innerW = W - PAD.left - PAD.right
    const innerH = H - PAD.top - PAD.bottom
    const maxValue = Math.max(1, ...list.map((d) => d.count))

    const step = list.length > 1 ? innerW / (list.length - 1) : 0
    const pts = list.map((d, i) => ({
      x: PAD.left + i * step,
      y: PAD.top + innerH - (d.count / maxValue) * innerH,
      ...d,
    }))

    const line = pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(' ')
    const area =
      pts.length > 0
        ? `${line} L${pts[pts.length - 1].x.toFixed(1)},${PAD.top + innerH} L${PAD.left},${PAD.top + innerH} Z`
        : ''

    // Y 轴刻度（0 / 中值 / 峰值）
    const ticks = [0, Math.round(maxValue / 2), maxValue]

    // X 轴日期标签（最多 6 个均匀采样）
    const labelIdx: number[] = []
    if (list.length > 0) {
      const n = Math.min(6, list.length)
      for (let i = 0; i < n; i++) {
        labelIdx.push(Math.round((i * (list.length - 1)) / Math.max(1, n - 1)))
      }
    }

    return { linePath: line, areaPath: area, points: pts, ticks, max: maxValue, labelIdx }
  }, [data])

  // 入场动画
  useLayoutEffect(() => {
    if (!svgRef.current || prefersReducedMotion()) return
    const ctx = gsap.context(() => {
      const svg = svgRef.current!
      const line = svg.querySelector('[data-line]')
      const area = svg.querySelector('[data-area]')
      const dots = svg.querySelectorAll('[data-dot]')
      if (line) {
        const len = (line as SVGPathElement).getTotalLength()
        gsap.fromTo(
          line,
          { strokeDasharray: len, strokeDashoffset: len },
          { strokeDashoffset: 0, duration: 1.2, ease: 'power2.inOut' },
        )
      }
      if (area) gsap.fromTo(area, { opacity: 0 }, { opacity: 1, duration: 0.8, delay: 0.5 })
      if (dots.length) {
        gsap.fromTo(
          dots,
          { scale: 0, transformOrigin: 'center' },
          { scale: 1, duration: 0.35, stagger: 0.02, delay: 0.5, ease: 'back.out(2)' },
        )
      }
    })
    return () => ctx.revert()
  }, [data])

  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-[200px] text-xs text-fog-500">
        近 30 天暂无互动记录
      </div>
    )
  }

  return (
    <svg ref={svgRef} viewBox={`0 0 ${W} ${H}`} className="w-full" role="img" aria-label="互动趋势图">
      <defs>
        <linearGradient id="trendFill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#FF8A4C" stopOpacity="0.35" />
          <stop offset="100%" stopColor="#FF8A4C" stopOpacity="0" />
        </linearGradient>
      </defs>

      {/* 网格线 + Y 轴刻度 */}
      {ticks.map((t) => {
        const y = PAD.top + (H - PAD.top - PAD.bottom) - (t / max) * (H - PAD.top - PAD.bottom)
        return (
          <g key={t}>
            <line x1={PAD.left} x2={W - PAD.right} y1={y} y2={y} stroke="#1F2532" strokeDasharray="3 4" />
            <text x={PAD.left - 6} y={y + 3.5} textAnchor="end" fontSize="10" fill="#5A6478">
              {t}
            </text>
          </g>
        )
      })}

      {/* 面积 */}
      <path data-area d={areaPath} fill="url(#trendFill)" />
      {/* 折线 */}
      <path data-line d={linePath} fill="none" stroke="#FF8A4C" strokeWidth="2" strokeLinejoin="round" strokeLinecap="round" />
      {/* 数据点 */}
      {points.map((p, i) => (
        <circle key={i} data-dot cx={p.x} cy={p.y} r="3" fill="#0B0E14" stroke="#FF8A4C" strokeWidth="2">
          <title>{`${p.date}：${p.count} 次`}</title>
        </circle>
      ))}
      {/* X 轴日期标签 */}
      {points.map((p, i) =>
        points.length <= 6 || i % Math.ceil(points.length / 6) === 0 ? (
          <text key={i} x={p.x} y={H - 8} textAnchor="middle" fontSize="9.5" fill="#5A6478">
            {p.date.slice(5)}
          </text>
        ) : null,
      )}
    </svg>
  )
}
