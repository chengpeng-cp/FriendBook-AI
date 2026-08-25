/**
 * Markdown 渲染组件：用于 AI 助手回复气泡
 * - remark-gfm：支持表格 / 删除线 / 任务列表等扩展语法
 * - components：手动映射元素样式，适配项目深色主题（ink/fog/brand 色板）
 *   （项目未引入 @tailwindcss/typography，故按元素显式定制，视觉更可控）
 */
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'

/** 元素 → 样式映射（与聊天气泡内边距协调，标题收敛避免撑破气泡） */
const components: Components = {
  h1: ({ children }) => <h1 className="text-base font-bold mt-3 mb-1.5 first:mt-0">{children}</h1>,
  h2: ({ children }) => <h2 className="text-[15px] font-bold mt-3 mb-1.5 first:mt-0">{children}</h2>,
  h3: ({ children }) => <h3 className="text-sm font-bold mt-2.5 mb-1 first:mt-0">{children}</h3>,
  h4: ({ children }) => <h4 className="text-sm font-bold mt-2 mb-1 first:mt-0">{children}</h4>,
  p: ({ children }) => <p className="my-1.5 leading-relaxed first:mt-0 last:mb-0">{children}</p>,
  ul: ({ children }) => <ul className="my-1.5 pl-4 list-disc space-y-0.5">{children}</ul>,
  ol: ({ children }) => <ol className="my-1.5 pl-4 list-decimal space-y-0.5">{children}</ol>,
  li: ({ children }) => <li className="leading-relaxed marker:text-fog-500">{children}</li>,
  strong: ({ children }) => <strong className="font-bold text-brand-300">{children}</strong>,
  em: ({ children }) => <em className="italic text-fog-400">{children}</em>,
  del: ({ children }) => <del className="text-fog-500">{children}</del>,
  a: ({ children, href }) => (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="text-brand-400 underline underline-offset-2 hover:text-brand-300"
    >
      {children}
    </a>
  ),
  hr: () => <hr className="my-3 border-ink-700/60" />,
  blockquote: ({ children }) => (
    <blockquote className="my-2 pl-3 border-l-2 border-brand-500/50 text-fog-400">{children}</blockquote>
  ),
  /* 行内代码：小胶囊；代码块：等宽 + 可横向滚动（不引入高亮库，保持轻量） */
  code: ({ className, children }) => {
    const isBlock = /language-/.test(className ?? '')
    if (isBlock) {
      return (
        <code className={`${className ?? ''} block font-mono text-xs leading-relaxed overflow-x-auto`}>
          {children}
        </code>
      )
    }
    return <code className="px-1.5 py-0.5 rounded-md bg-ink-950/70 text-[13px] font-mono text-brand-300">{children}</code>
  },
  pre: ({ children }) => (
    <pre className="my-2 p-3 rounded-xl bg-ink-950/70 border border-ink-700/60 overflow-x-auto">{children}</pre>
  ),
  /* 表格：AI 统计回复的高频输出，边框 + 紧凑行高 + 横向滚动兜底 */
  table: ({ children }) => (
    <div className="my-2 overflow-x-auto">
      <table className="w-full text-[13px] border-collapse border border-ink-700/60">{children}</table>
    </div>
  ),
  thead: ({ children }) => <thead className="bg-ink-950/60">{children}</thead>,
  th: ({ children }) => (
    <th className="px-2.5 py-1.5 text-left font-bold border border-ink-700/60 whitespace-nowrap">{children}</th>
  ),
  td: ({ children }) => (
    <td className="px-2.5 py-1.5 border border-ink-700/60 whitespace-nowrap">{children}</td>
  ),
  img: ({ src, alt }) => (
    <img src={src} alt={alt ?? ''} className="my-2 max-w-full rounded-xl border border-ink-700/60" />
  ),
}

/**
 * 预处理：修复「** 紧贴中文标点」导致的强调解析失败。
 * CommonMark 的 flanking 规则会把紧贴 Unicode 标点（“”「」等）的 ** 判为
 * 无效分隔符，字面输出星号（如 从**“场景”**入手）。在边界插入零宽空格
 * （U+200B，非标点非空白）即可恢复判定，且不影响显示文本。
 */
const CJK_PUNCT = /^[“”「」『』（）：，。！？；、]/
const ZWSP = '\u200b'

function fixEmphasisFlanking(s: string): string {
  return s.replace(/\*\*(\S[^*]*?\S|\S)\*\*/g, (m, inner: string) => {
    const head = CJK_PUNCT.test(inner) ? ZWSP : ''
    const tail = CJK_PUNCT.test(inner[inner.length - 1]) ? ZWSP : ''
    // 边界正常时原样返回，避免无谓改写
    return head === '' && tail === '' ? m : `**${head}${inner}${tail}**`
  })
}

/** Markdown 渲染器（AI 回复专用） */
export default function Markdown({ content }: { content: string }) {
  return (
    <div className="md-body text-sm break-words">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {fixEmphasisFlanking(content)}
      </ReactMarkdown>
    </div>
  )
}
