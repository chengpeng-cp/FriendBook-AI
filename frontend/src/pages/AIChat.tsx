/**
 * AI 助手对话页
 * - SSE 流式输出：打字机式追加文本 + 光标闪烁
 * - 工具调用指示：显示"正在调用 xxx 工具"
 * - 多轮上下文：最近 10 条历史随请求携带
 * - AI 调用的工具会真实修改数据，操作后可去对应页面查看
 */
import { useEffect, useRef, useState, type FormEvent } from 'react'
import Markdown from '@/components/Markdown'
import { streamChat, aiHistoryApi } from '@/api/ai'
import type { ChatMessage } from '@/types'
import { usePageEnter } from '@/hooks/useGsap'
import { toast } from '@/stores/toast'

/** 展示用消息（含工具调用过程态） */
interface DisplayMsg {
  role: 'user' | 'assistant'
  content: string
  /** 本条回复中出现过的工具调用（用于回显） */
  tools: string[]
  streaming?: boolean
}

/** 建议指令 */
const SUGGESTIONS = [
  { icon: '🔍', text: '帮我看看哪些好友关系变凉了' },
  { icon: '➕', text: '帮我加一个好友：张三，微信号 zhangsan_wx，好感度 80' },
  { icon: '📊', text: '分析一下我的社交圈健康状况' },
  { icon: '🎂', text: '最近有谁要过生日吗' },
  { icon: '🌡', text: '列出好感度最高的 5 位好友' },
]

/** AI 助手身份头衔 */
const PERSONA_TAG = 'Eino ReAct Agent · 已接入 12 个工具'

/** 无历史时的欢迎语 */
const WELCOME_MSG: DisplayMsg = {
  role: 'assistant',
  content:
    '你好呀，我是 FriendBook AI 助手 👋\n\n我可以直接帮你管理好友：增删改查、调好感度、记录互动、分析社交圈……\n试试对我说「加一个好友张三」或点击下方建议指令。',
  tools: [],
}

export default function AIChat() {
  const pageRef = usePageEnter<HTMLDivElement>()
  const [messages, setMessages] = useState<DisplayMsg[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [activeTool, setActiveTool] = useState('') // 当前正在调用的工具
  const abortRef = useRef<AbortController | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)

  // 初始化：优先加载服务端持久化的对话历史，无历史时展示欢迎语
  useEffect(() => {
    aiHistoryApi
      .list()
      .then((res) => {
        const history = res.list ?? []
        if (history.length > 0) {
          setMessages(history.map((m) => ({ role: m.role, content: m.content, tools: [] })))
        } else {
          setMessages([WELCOME_MSG])
        }
      })
      .catch(() => setMessages([WELCOME_MSG])) // 历史加载失败降级为欢迎语
  }, [])

  // 新消息自动滚底
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [messages, activeTool])

  /** 发送消息（SSE 流式接收） */
  const send = async (text?: string) => {
    const message = (text ?? input).trim()
    if (!message || sending) return

    setInput('')
    setSending(true)

    // 追加用户消息 + 立即创建"思考中"的助手占位（避免 AI 响应慢时界面无反馈）
    const history: ChatMessage[] = messages
      .filter((m) => m.content && !m.streaming)
      .slice(-10) // 最近 10 轮随行
      .map((m) => ({ role: m.role, content: m.content }))
    let assistantIdx = -1
    setMessages((prev) => {
      const next = [...prev, { role: 'user', content: message, tools: [] }]
      next.push({ role: 'assistant', content: '', tools: [], streaming: true })
      assistantIdx = next.length - 1
      return next
    })

    const ctrl = new AbortController()
    abortRef.current = ctrl

    await streamChat(
      message,
      history,
      {
        onEvent: (ev) => {
          if (ev.done) return
          // 工具调用帧：更新指示器
          if (ev.tool_call) {
            setActiveTool(ev.tool_call)
            setMessages((prev) => {
              const next = [...prev]
              // 若助手回复尚未创建，先创建占位
              if (assistantIdx < 0) {
                next.push({ role: 'assistant', content: '', tools: [], streaming: true })
                assistantIdx = next.length - 1
              }
              const msg = next[assistantIdx]
              if (!msg.tools.includes(ev.tool_call!)) {
                next[assistantIdx] = { ...msg, tools: [...msg.tools, ev.tool_call!] }
              }
              return next
            })
          }
          // 内容帧：追加文本
          if (ev.content) {
            setMessages((prev) => {
              const next = [...prev]
              if (assistantIdx < 0) {
                next.push({ role: 'assistant', content: '', tools: [], streaming: true })
                assistantIdx = next.length - 1
              }
              next[assistantIdx] = { ...next[assistantIdx], content: next[assistantIdx].content + ev.content }
              return next
            })
          }
        },
        onDone: () => {
          // 结束：清空流式态
          setActiveTool('')
          setMessages((prev) => prev.map((m) => (m.streaming ? { ...m, streaming: false } : m)))
        },
        onError: (msg) => {
          toast.error(msg)
          setMessages((prev) => [
            ...prev.filter((m) => !(m.streaming && !m.content)), // 移除空占位
            { role: 'assistant', content: `抱歉，出错了：${msg}`, tools: [] },
          ])
        },
      },
      ctrl.signal,
    )

    setSending(false)
    abortRef.current = null
  }

  /** 停止生成 */
  const stop = () => {
    abortRef.current?.abort()
    setSending(false)
    setActiveTool('')
  }

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    send()
  }

  return (
    <div ref={pageRef} className="page-enter max-w-3xl mx-auto h-[calc(100vh-8rem)] flex flex-col">
      {/* 头部 */}
      <div className="flex items-center gap-3.5 pb-4 border-b border-ink-700/50">
        <div className="relative h-11 w-11 rounded-2xl bg-gradient-to-br from-violet-400 to-brand-600 flex items-center justify-center text-xl shadow-glow">
          ✨
          {/* 在线呼吸点 */}
          <span className="absolute -right-0.5 -bottom-0.5 h-3 w-3 rounded-full bg-emerald-400 border-2 border-ink-950 animate-pulse-soft" />
        </div>
        <div>
          <h1 className="text-base font-bold">AI 助手</h1>
          <p className="text-[11px] text-fog-500">{PERSONA_TAG}</p>
        </div>
      </div>

      {/* 消息区 */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto py-5 flex flex-col gap-5 pr-1">
        {messages.map((m, i) => (
          <div key={i} className={`flex gap-3 ${m.role === 'user' ? 'flex-row-reverse' : ''}`}>
            {/* 头像 */}
            <div
              className={`h-8 w-8 shrink-0 rounded-xl flex items-center justify-center text-sm ${
                m.role === 'user'
                  ? 'bg-ink-700 text-fog-700'
                  : 'bg-gradient-to-br from-violet-400 to-brand-600'
              }`}
            >
              {m.role === 'user' ? '我' : '✨'}
            </div>

            <div className={`max-w-[78%] flex flex-col gap-1.5 ${m.role === 'user' ? 'items-end' : ''}`}>
              {/* 工具调用回显 */}
              {m.tools.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {m.tools.map((t) => (
                    <span
                      key={t}
                      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-lg text-[10px] font-medium bg-sky-500/10 border border-sky-500/30 text-sky-300"
                    >
                      🔧 {t}
                    </span>
                  ))}
                </div>
              )}

              {/* 气泡：AI 回复为 Markdown 富文本渲染，用户消息为纯文本 */}
              <div
                className={`px-4 py-3 rounded-2xl text-sm leading-relaxed break-words ${
                  m.role === 'user'
                    ? 'bg-gradient-to-r from-brand-500 to-brand-600 text-white rounded-tr-md whitespace-pre-wrap'
                    : 'bg-ink-800 border border-ink-700/60 rounded-tl-md'
                } ${m.streaming && !m.content ? 'ai-caret' : ''}`}
              >
                {m.role === 'assistant' ? (
                  <>
                    {/* 空内容流式占位（配合 ai-caret 光标） */}
                    {m.content ? <Markdown content={m.content} /> : m.streaming ? '' : '…'}
                    {m.streaming && m.content && <span className="ai-caret" />}
                  </>
                ) : (
                  m.content
                )}
              </div>
            </div>
          </div>
        ))}

        {/* 工具调用进行中指示 */}
        {activeTool && (
          <div className="flex gap-3">
            <div className="h-8 w-8 shrink-0 rounded-xl bg-gradient-to-br from-violet-400 to-brand-600 flex items-center justify-center text-sm">
              ✨
            </div>
            <div className="px-4 py-2.5 rounded-2xl rounded-tl-md bg-ink-800 border border-sky-500/30 text-xs text-sky-300 flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-sky-400 animate-pulse-soft" />
              正在调用工具 <code className="text-sky-200 bg-sky-500/10 px-1.5 py-0.5 rounded">{activeTool}</code>
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* 建议指令（无对话时展示） */}
      {messages.length <= 1 && !sending && (
        <div className="flex flex-wrap gap-2 pb-3">
          {SUGGESTIONS.map((s) => (
            <button
              key={s.text}
              className="px-3 py-1.5 rounded-xl text-xs bg-ink-800/80 border border-ink-700/60 text-fog-700 hover:border-brand-500/40 hover:text-brand-400 transition-all"
              onClick={() => send(s.text)}
            >
              {s.icon} {s.text}
            </button>
          ))}
        </div>
      )}

      {/* 输入区 */}
      <form onSubmit={onSubmit} className="flex gap-2.5 pt-3 border-t border-ink-700/50">
        <input
          className="fb-input flex-1 !h-12"
          placeholder={sending ? 'AI 正在回复…' : '对我说：帮我加个好友 / 谁快过生日了 / 分析我的社交圈…'}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          disabled={sending}
        />
        {sending ? (
          <button type="button" className="fb-btn-ghost !h-12 w-24" onClick={stop}>
            ■ 停止
          </button>
        ) : (
          <button type="submit" className="fb-btn-primary !h-12 w-24" disabled={!input.trim()}>
            发送
          </button>
        )}
      </form>
      <p className="pt-2 text-center text-[10px] text-fog-300">
        AI 可直接操作你的好友数据 · 由 Eino ReAct Agent 驱动 · 请勿输入敏感信息
      </p>
    </div>
  )
}
