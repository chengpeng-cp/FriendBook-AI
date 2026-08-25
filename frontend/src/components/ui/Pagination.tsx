/**
 * 分页器（配合后端 Paged 结构）
 */

interface PaginationProps {
  page: number
  pageSize: number
  total: number
  onChange: (page: number) => void
}

export default function Pagination({ page, pageSize, total, onChange }: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  if (total === 0) return null

  // 页码窗口：当前页前后各 2 页
  const pages: number[] = []
  const start = Math.max(1, page - 2)
  const end = Math.min(totalPages, start + 4)
  for (let i = start; i <= end; i++) pages.push(i)

  const btnBase =
    'h-8 min-w-8 px-2 rounded-lg text-xs transition-colors disabled:opacity-40 disabled:pointer-events-none'

  return (
    <div className="flex items-center justify-between mt-6 text-xs text-fog-500">
      <span>
        共 <span className="text-fog-900 font-medium">{total}</span> 条 · 第 {page}/{totalPages} 页
      </span>
      <div className="flex items-center gap-1.5">
        <button
          className={`${btnBase} bg-ink-700/50 hover:bg-ink-700`}
          disabled={page <= 1}
          onClick={() => onChange(page - 1)}
        >
          上一页
        </button>
        {start > 1 && <span className="px-1">…</span>}
        {pages.map((p) => (
          <button
            key={p}
            className={`${btnBase} ${
              p === page
                ? 'bg-gradient-to-r from-brand-500 to-brand-600 text-white font-semibold'
                : 'bg-ink-700/50 hover:bg-ink-700'
            }`}
            onClick={() => onChange(p)}
          >
            {p}
          </button>
        ))}
        {end < totalPages && <span className="px-1">…</span>}
        <button
          className={`${btnBase} bg-ink-700/50 hover:bg-ink-700`}
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
        >
          下一页
        </button>
      </div>
    </div>
  )
}
