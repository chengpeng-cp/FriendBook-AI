/**
 * 数据统计页：全维度社交圈分析
 * 复用图表组件：温度分布 / 互动趋势 / 平台分布
 */
import { useEffect, useState } from 'react'
import { statsApi } from '@/api/data'
import type { PlatformDistItem, TemperatureDist, TrendPoint } from '@/types'
import { usePageEnter } from '@/hooks/useGsap'
import { PageLoading } from '@/components/ui/Basics'
import TempBars from '@/components/charts/TempBars'
import TrendChart from '@/components/charts/TrendChart'
import PlatformDonut from '@/components/charts/PlatformDonut'
import { toast } from '@/stores/toast'

export default function Stats() {
  const pageRef = usePageEnter<HTMLDivElement>()
  const [tempDist, setTempDist] = useState<TemperatureDist | null>(null)
  const [trend, setTrend] = useState<TrendPoint[] | null>(null)
  const [platforms, setPlatforms] = useState<PlatformDistItem[] | null>(null)

  useEffect(() => {
    Promise.all([statsApi.tempDist(), statsApi.trend(), statsApi.platformDist()])
      .then(([td, tr, pf]) => {
        setTempDist(td)
        // list 兜底为空数组：后端历史缓存可能返回 null，null 会让图表区永久转圈
        setTrend(tr.list ?? [])
        setPlatforms(pf.list ?? [])
      })
      .catch((e) => toast.error((e as Error).message))
  }, [])

  /** 温度健康度：热档（夯+热）占比 */
  const tempTotal =
    tempDist ? tempDist.hot + tempDist.warm + tempDist.mild + tempDist.cool + tempDist.cold : 0
  const healthyPct = tempTotal > 0 ? Math.round(((tempDist!.hot + tempDist!.warm) / tempTotal) * 100) : 0
  const trendTotal = trend?.reduce((s, d) => s + d.count, 0) ?? 0

  return (
    <div ref={pageRef} className="page-enter max-w-6xl mx-auto">
      <h1 className="text-xl font-bold mb-1">数据统计</h1>
      <p className="text-xs text-fog-500 mb-7">用数据读懂你的社交圈（后端 Redis 缓存 5 分钟，指标有轻微延迟属正常）</p>

      {/* 摘要行 */}
      <div className="grid grid-cols-3 gap-4 mb-5">
        <div data-stagger className="fb-card p-5">
          <p className="text-xs text-fog-500">关系健康度</p>
          <p className="mt-1.5 text-3xl font-bold text-brand-400 tabular-nums">
            {tempDist ? `${healthyPct}%` : '—'}
          </p>
          <p className="mt-1 text-[11px] text-fog-500">「夯 + 热」档好友占比</p>
        </div>
        <div data-stagger className="fb-card p-5">
          <p className="text-xs text-fog-500">30 天互动总量</p>
          <p className="mt-1.5 text-3xl font-bold text-sky-400 tabular-nums">{trend ? trendTotal : '—'}</p>
          <p className="mt-1 text-[11px] text-fog-500">平均 {(trend ? (trendTotal / 30).toFixed(1) : '—')} 次/天</p>
        </div>
        <div data-stagger className="fb-card p-5">
          <p className="text-xs text-fog-500">在册平台账号</p>
          <p className="mt-1.5 text-3xl font-bold text-violet-400 tabular-nums">
            {platforms ? platforms.reduce((s, p) => s + p.count, 0) : '—'}
          </p>
          <p className="mt-1 text-[11px] text-fog-500">{platforms ? `${platforms.length} 个平台` : ''}</p>
        </div>
      </div>

      {/* 图表区 */}
      <div className="grid lg:grid-cols-5 gap-4 mb-4">
        <div data-stagger className="fb-card p-5 lg:col-span-3">
          <h2 className="text-sm font-semibold mb-4">近 30 天互动趋势</h2>
          {trend ? <TrendChart data={trend} /> : <PageLoading />}
        </div>
        <div data-stagger className="fb-card p-5 lg:col-span-2">
          <h2 className="text-sm font-semibold mb-4">关系温度分布</h2>
          {tempDist ? <TempBars data={tempDist} /> : <PageLoading />}
        </div>
      </div>

      <div data-stagger className="fb-card p-5">
        <h2 className="text-sm font-semibold mb-4">好友平台分布</h2>
        {platforms ? <PlatformDonut data={platforms} /> : <PageLoading />}
      </div>
    </div>
  )
}
