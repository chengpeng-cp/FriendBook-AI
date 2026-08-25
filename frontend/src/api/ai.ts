/**
 * AI 助手 SSE 流式对话
 *
 * 后端为 POST + text/event-stream（EventSource 不支持 POST/自定义头），
 * 故用 fetch + ReadableStream 手动解析 SSE 帧：
 *   event: message \n data: {"content":"...","tool_call":"...","done":false} \n\n
 */
import { API_BASE, del, get } from './client'
import type { AiStreamEvent, ChatHistoryItem, ChatMessage, ListWrap } from '@/types'

export interface StreamHandlers {
  /** 每收到一帧触发 */
  onEvent: (ev: AiStreamEvent) => void
  /** 流正常/异常结束 */
  onDone: () => void
  /** 请求层错误（网络 / 鉴权 / 业务码） */
  onError: (msg: string) => void
}

export async function streamChat(
  message: string,
  history: ChatMessage[],
  handlers: StreamHandlers,
  signal?: AbortSignal,
): Promise<void> {
  const token = localStorage.getItem('fb_token')

  let res: Response
  try {
    res = await fetch(`${API_BASE}/api/v1/ai/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ message, history }),
      signal,
    })
  } catch (e) {
    // 用户主动中断不算错误
    if ((e as Error).name === 'AbortError') return handlers.onDone()
    return handlers.onError('网络异常，无法连接 AI 助手')
  }

  // 流开始前的业务错误：后端返回 JSON（非 SSE）
  const ctype = res.headers.get('content-type') ?? ''
  if (!ctype.includes('text/event-stream')) {
    try {
      const body = (await res.json()) as { message?: string }
      return handlers.onError(body.message || 'AI 对话失败')
    } catch {
      return handlers.onError('AI 对话失败')
    }
  }

  const reader = res.body!.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  /** 解析一行 data: 前缀的 JSON 帧 */
  const dispatch = (line: string) => {
    if (!line.startsWith('data:')) return
    const payload = line.slice(5).trim()
    if (!payload) return
    try {
      const ev = JSON.parse(payload) as AiStreamEvent
      // 流中断帧：后端透传错误信息，立即转为 onError（否则表现为"发出去没回复"的静默失败）
      if (ev.error) {
        handlers.onError(ev.error)
        return true
      }
      handlers.onEvent(ev)
      if (ev.done) return true
    } catch {
      /* 忽略无法解析的帧（如注释心跳行） */
    }
    return false
  }

  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      // SSE 以 \n 分行；逐行派发，done 帧出现即收尾
      let newlineIdx: number
      let finished = false
      while ((newlineIdx = buffer.indexOf('\n')) >= 0) {
        const line = buffer.slice(0, newlineIdx).trim()
        buffer = buffer.slice(newlineIdx + 1)
        if (dispatch(line)) {
          finished = true
          break
        }
      }
      if (finished) break
    }
  } catch {
    /* 流中断：按结束处理 */
  } finally {
    reader.releaseLock()
    handlers.onDone()
  }
}

/** AI 对话历史 API（后端持久化，刷新页面可恢复） */
export const aiHistoryApi = {
  /** 最近 50 条对话（时间正序） */
  list: () => get<ListWrap<ChatHistoryItem>>('/api/v1/ai/history'),
  /** 清空对话记录 */
  clear: () => del<null>('/api/v1/ai/history'),
}
