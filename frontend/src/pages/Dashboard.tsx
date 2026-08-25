/**
 * 仪表盘：社交圈总览
 * - 4 张统计卡（GSAP 数字滚动）
 * - 平均好感度环形仪表
 * - 温度分布条形图 / 互动趋势面积图 / 平台分布环形图
 * - 今日提醒 + 变凉好友快捷列表
 */
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { statsApi, reminderApi } from '@/api/data'
import { friendApi } from '@/api/friends'
import type { Friend, Reminder, StatsOverview } from '@/types'
import { usePageEnter, useCountUp } from '@/hooks/useGsap'
import { Avatar, Empty, PageLoading, SectionTitle, TempBadge } from '@/components/ui/Basics'
import TempBars from '@/components/charts/TempBars'
import TrendChart from '@/components/charts/TrendChart'
import PlatformDonut from '@/components/charts/PlatformDonut'
import { daysUntil, fmtDate, fmtRelative } from '@/utils/format'
import { REMINDER_META } from '@/constants'
import type { TemperatureDist } from '@/types'

/** 统计卡（数字滚动） */
function StatCard({
  label,
  value,
  decimals = 0,
  suffix,
  icon,
  accent,
}: {
  label: string
  value: number
  decimals?: number
  suffix?: string
  icon: string
  accent: string
}) {
  const numRef = useCountUp(value, decimals)
  return (
    <div data-stagger className="fb-card p-5 relative overflow-hidden group hover:border-fog-300/20 transition-colors">
      {/* 悬浮光晕 */}
      <div
        className="absolute -right-6 -top-6 h-20 w-20 rounded-full blur-2xl opacity-20 group-hover:opacity-40 transition-opacity"
        style={{ background: accent }}
      />
      <div className="flex items-center justify-between">
        <span className="text-xs text-fog-500">{label}</span>
        <span className="text-sm">{icon}</span>
      </div>
      <div className="mt-2 text-3xl font-bold tabular-nums" style={{ color: accent }}>
        <span ref={numRef}>0</span>
        {suffix && <span className="text-sm font-medium text-fog-500 ml-1">{suffix}</span>}
      </div>
    </div>
  )
}

/** 平均好感度环形仪表（SVG stroke 动画） */
function FavorGauge({ avg }: { avg: number }) {
  const R = 52
  const C = 2 * Math.PI * R
  const pct = Math.min(100, Math.max(0, avg)) / 100
  const numRef = useCountUp<SVGTSpanElement>(avg, 1)

  return (
    <div data-stagger className="fb-card p-5 flex items-center gap-6">
      <svg width="132" height="132" viewBox="0 0 132 132" className="shrink-0">
        <defs>
          <linearGradient id="gaugeGrad" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor="#FFB25C" />
            <stop offset="100%" stopColor="#F0653A" />
          </linearGradient>
        </defs>
        <g transform="translate(66,66) rotate(-90)">
          <circle r={R} fill="none" stroke="#1F2532" strokeWidth="10" />
          <circle
            r={R}
            fill="none"
            stroke="url(#gaugeGrad)"
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={C}
            strokeDashoffset={C * (1 - pct)}
            style={{ transition: 'stroke-dashoffset 1.1s cubic-bezier(.22,1,.36,1)' }}
          />
        </g>
        <text x="66" y="62" textAnchor="middle" fontSize="26" fontWeight="700" fill="#E8ECF3">
          <tspan ref={numRef}>0</tspan>
        </text>
        <text x="66" y="82" textAnchor="middle" fontSize="10" fill="#5A6478">
          平均好感度
        </text>
      </svg>
      <div className="flex-1">
        <p className="text-sm font-medium text-fog-900">社交圈温度计</p>
        <p className="mt-2 text-xs leading-relaxed text-fog-500">
          {avg >= 75
            ? '你的关系维护得非常到位，继续保持高频互动！'
            : avg >= 60
              ? '整体关系健康，注意别让「温」档的好友滑向「凉」。'
              : avg >= 40
                ? '不少关系正在降温，试着主动发起一次互动。'
                : '社交圈偏冷了，让 AI 助手帮你找出该挽回的好友吧。'}
        </p>
        <Link
          to="/ai"
          className="mt-3 inline-flex items-center gap-1.5 text-xs text-brand-400 hover:text-brand-500 transition-colors"
        >
          问问 AI 助手 →
        </Link>
      </div>
    </div>
  )
}

export default function Dashboard() {
  const pageRef = usePageEnter<HTMLDivElement>()
  const [overview, setOverview] = useState<StatsOverview | null>(null)
  const [tempDist, setTempDist] = useState<TemperatureDist | null>(null)
  const [platforms, setPlatforms] = useState<{ platform: string; count: number }[] | null>(null)
  const [trend, setTrend] = useState<{ date: string; count: number }[] | null>(null)
  const [reminders, setReminders] = useState<Reminder[] | null>(null)
  const [coldFriends, setColdFriends] = useState<Friend[] | null>(null)
  const [error, setError] = useState('')

  // 并行拉取全部看板数据
  useEffect(() => {
    Promise.all([
      statsApi.overview(),
      statsApi.tempDist(),
      statsApi.platformDist(),
      statsApi.trend(),
      reminderApi.today(),
      friendApi.cold(1, 5),
    ])
      .then(([ov, td, pf, tr, rm, cold]) => {
        setOverview(ov)
        setTempDist(td)
        // list 兜底为空数组：后端历史缓存可能返回 null，null 会让列表区永久转圈
        setPlatforms(pf.list ?? [])
        setTrend(tr.list ?? [])
        setReminders(rm.list ?? [])
        setColdFriends(cold.list ?? [])
      })
      .catch((e) => setError((e as Error).message))
  }, [])

  if (error) {
    return (
      <div className="fb-card p-10 text-center text-sm text-rose-400">
        加载失败：{error}
        <button className="block mx-auto mt-4 fb-btn-ghost" onClick={() => location.reload()}>
          重试
        </button>
      </div>
    )
  }
  if (!overview) return <PageLoading text="正在汇交你的社交圈数据…" />

  return (
    <div ref={pageRef} className="page-enter max-w-6xl mx-auto">
      <h1 className="text-xl font-bold mb-1">仪表盘</h1>
      <p className="text-xs text-fog-500 mb-7">你的社交圈健康度一览（统计数据每 5 分钟刷新一次缓存）</p>

      {/* 统计卡行 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="好友总数" value={overview.total} icon="👥" accent="#FF8A4C" />
        <StatCard label="本月新增" value={overview.new_this_month} icon="🌱" accent="#4FD1C5" />
        <StatCard label="今日提醒" value={overview.today_reminders} icon="⏰" accent="#F6C453" />
        <StatCard label="关系变凉" value={overview.cold_count} icon="❄️" accent="#63B3ED" />
      </div>

      {/* 第二行：仪表 + 温度分布 */}
      <div className="mt-5 grid lg:grid-cols-2 gap-4">
        <FavorGauge avg={overview.avg_favor} />
        <div data-stagger className="fb-card p-5">
          <SectionTitle>关系温度分布</SectionTitle>
          <TempBars data={tempDist} />
        </div>
      </div>

      {/* 第三行：趋势 + 平台 */}
      <div className="mt-5 grid lg:grid-cols-5 gap-4">
        <div data-stagger className="fb-card p-5 lg:col-span-3">
          <SectionTitle>近 30 天互动趋势</SectionTitle>
          <TrendChart data={trend} />
        </div>
        <div data-stagger className="fb-card p-5 lg:col-span-2">
          <SectionTitle>好友平台分布</SectionTitle>
          <PlatformDonut data={platforms} />
        </div>
      </div>

      {/* 第四行：提醒 + 变凉好友 */}
      <div className="mt-5 grid lg:grid-cols-2 gap-4">
        <div data-stagger className="fb-card p-5">
          <SectionTitle
            action={
              <Link to="/reminders" className="text-xs text-fog-500 hover:text-fog-900 transition-colors">
                查看全部 →
              </Link>
            }
          >
            今日提醒（含未来 3 天）
          </SectionTitle>
          {reminders === null ? (
            <PageLoading />
          ) : reminders.length === 0 ? (
            <Empty icon="🎂" title="近期没有提醒" hint="为好友填写生日后会自动生成" />
          ) : (
            <ul className="flex flex-col gap-2.5">
              {reminders.slice(0, 5).map((r) => {
                const days = daysUntil(r.date)
                const meta = REMINDER_META[r.type]
                return (
                  <li
                    key={r.id}
                    className="flex items-center gap-3 p-3 rounded-xl bg-ink-900/60 border border-ink-700/50"
                  >
                    <span className="text-lg">{meta.emoji}</span>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate">{r.title}</p>
                      <p className="text-[11px] text-fog-500">
                        {meta.label} · {fmtDate(r.date)}
                      </p>
                    </div>
                    <span
                      className={`text-xs font-semibold px-2 py-1 rounded-lg ${
                        days <= 0
                          ? 'bg-brand-500/15 text-brand-400'
                          : days <= 3
                            ? 'bg-amber-500/15 text-amber-400'
                            : 'bg-ink-700/50 text-fog-500'
                      }`}
                    >
                      {days <= 0 ? '就在今天' : `还有 ${days} 天`}
                    </span>
                  </li>
                )
              })}
            </ul>
          )}
        </div>

        <div data-stagger className="fb-card p-5">
          <SectionTitle
            action={
              <Link to="/friends?tab=cold" className="text-xs text-fog-500 hover:text-fog-900 transition-colors">
                查看全部 →
              </Link>
            }
          >
            关系变凉预警
          </SectionTitle>
          {coldFriends === null ? (
            <PageLoading />
          ) : coldFriends.length === 0 ? (
            <Empty icon="☀️" title="没有变凉的好友" hint="你的关系维护非常棒" />
          ) : (
            <ul className="flex flex-col gap-2.5">
              {coldFriends.map((f) => (
                <li key={f.id}>
                  <Link
                    to={`/friends/${f.id}`}
                    className="flex items-center gap-3 p-3 rounded-xl bg-ink-900/60 border border-ink-700/50 hover:border-sky-500/30 transition-colors group"
                  >
                    <Avatar id={f.id} name={f.name} size={36} />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate group-hover:text-white transition-colors">
                        {f.name}
                        {f.remark && <span className="text-fog-500 ml-1.5 text-xs">{f.remark}</span>}
                      </p>
                      <p className="text-[11px] text-fog-500">最近互动：{fmtRelative(f.last_interact)}</p>
                    </div>
                    <TempBadge temp={f.temperature} size="sm" />
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}
