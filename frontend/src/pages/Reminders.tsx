/**
 * 提醒中心：今日提醒（含未来 3 天）+ 近期提醒（未来 7 天）
 */
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { reminderApi } from '@/api/data'
import type { Reminder } from '@/types'
import { usePageEnter } from '@/hooks/useGsap'
import { Empty, PageLoading, SectionTitle } from '@/components/ui/Basics'
import { REMINDER_META } from '@/constants'
import { daysUntil, fmtDate } from '@/utils/format'

/** 提醒卡片 */
function ReminderCard({ r }: { r: Reminder }) {
  const days = daysUntil(r.date)
  const meta = REMINDER_META[r.type] ?? { label: '自定义', emoji: '🔔' }

  // 紧迫度配色
  const urgent = days <= 0
  const soon = days > 0 && days <= 3

  return (
    <Link
      to={`/friends/${r.friend_id}`}
      className={`block p-4 rounded-2xl border transition-all duration-200 hover:-translate-y-0.5 ${
        urgent
          ? 'bg-brand-500/10 border-brand-500/40 hover:shadow-glow'
          : soon
            ? 'bg-amber-500/8 border-amber-500/30'
            : 'bg-ink-800/80 border-ink-700/60'
      }`}
    >
      <div className="flex items-center gap-3">
        <span
          className="h-11 w-11 shrink-0 rounded-2xl flex items-center justify-center text-xl"
          style={{
            background: urgent ? 'rgba(255,138,76,.15)' : soon ? 'rgba(246,196,83,.12)' : 'rgba(31,37,50,.8)',
          }}
        >
          {meta.emoji}
        </span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{r.title}</p>
          <p className="text-[11px] text-fog-500 mt-0.5">
            {meta.label} · {fmtDate(r.date)} · 提前 {r.days_in_advance} 天
          </p>
        </div>
        <span
          className={`shrink-0 text-xs font-semibold px-2.5 py-1.5 rounded-xl ${
            urgent ? 'bg-brand-500/20 text-brand-400' : soon ? 'bg-amber-500/15 text-amber-400' : 'bg-ink-700/50 text-fog-500'
          }`}
        >
          {urgent ? '就在今天' : `${days} 天后`}
        </span>
      </div>
    </Link>
  )
}

export default function Reminders() {
  const pageRef = usePageEnter<HTMLDivElement>()
  const [today, setToday] = useState<Reminder[] | null>(null)
  const [upcoming, setUpcoming] = useState<Reminder[] | null>(null)

  useEffect(() => {
    reminderApi
      .today()
      .then((r) => setToday(r.list))
      .catch(() => setToday([]))
    reminderApi
      .upcoming()
      .then((r) => setUpcoming(r.list))
      .catch(() => setUpcoming([]))
  }, [])

  return (
    <div ref={pageRef} className="page-enter max-w-4xl mx-auto">
      <h1 className="text-xl font-bold mb-1">提醒中心</h1>
      <p className="text-xs text-fog-500 mb-7">
        生日与纪念日提醒由系统每日定时扫描生成 · 为好友填写生日即可自动提醒
      </p>

      {/* 今日提醒 */}
      <div className="fb-card p-5 mb-5" data-stagger>
        <SectionTitle>🔔 近在眼前（今日 + 未来 3 天）</SectionTitle>
        {today === null ? (
          <PageLoading />
        ) : today.length === 0 ? (
          <Empty icon="🗓" title="近期没有提醒" hint="给好友补充生日信息，让重要日子不再错过" />
        ) : (
          <div className="grid sm:grid-cols-2 gap-3">
            {today.map((r) => (
              <ReminderCard key={r.id} r={r} />
            ))}
          </div>
        )}
      </div>

      {/* 近期提醒 */}
      <div className="fb-card p-5" data-stagger>
        <SectionTitle>📅 未来 7 天</SectionTitle>
        {upcoming === null ? (
          <PageLoading />
        ) : upcoming.length === 0 ? (
          <Empty icon="🌤" title="未来 7 天很安静" />
        ) : (
          <div className="grid sm:grid-cols-2 gap-3">
            {upcoming.map((r) => (
              <ReminderCard key={r.id} r={r} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
