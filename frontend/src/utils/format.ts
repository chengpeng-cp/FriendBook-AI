/** 通用格式化工具 */

/** 日期时间 → "2026-08-15 14:30" */
export function fmtDateTime(iso: string | null | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

/** 日期 → "2026-08-15" */
export function fmtDate(iso: string | null | undefined): string {
  if (!iso) return '—'
  return iso.slice(0, 10)
}

/** 相对时间："3 天前"、"刚刚"、"2 个月后" */
export function fmtRelative(iso: string | null | undefined): string {
  if (!iso) return '从未'
  const diff = Date.now() - new Date(iso).getTime()
  if (Number.isNaN(diff)) return '从未'
  const abs = Math.abs(diff)
  const minute = 60_000
  const hour = 60 * minute
  const day = 24 * hour
  const suffix = diff >= 0 ? '前' : '后'
  if (abs < minute) return '刚刚'
  if (abs < hour) return `${Math.floor(abs / minute)} 分钟${suffix}`
  if (abs < day) return `${Math.floor(abs / hour)} 小时${suffix}`
  if (abs < 30 * day) return `${Math.floor(abs / day)} 天${suffix}`
  if (abs < 365 * day) return `${Math.floor(abs / (30 * day))} 个月${suffix}`
  return `${Math.floor(abs / (365 * day))} 年${suffix}`
}

/** 距目标日期还有几天（负数表示已过去） */
export function daysUntil(dateIso: string): number {
  const target = new Date(dateIso.slice(0, 10) + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return Math.round((target.getTime() - today.getTime()) / 86_400_000)
}

/** 文件大小 */
export function fmtBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}
