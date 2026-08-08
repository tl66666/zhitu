// SSE 流式请求工具
// 后端 SSE 接口需 POST + JWT header，EventSource 不可用，使用 fetch + ReadableStream
import type { SSECallbacks, SSEEvent } from '@/types/models'
import { useAuthStore } from '@/stores/auth'
import { useAdminAuthStore } from '@/stores/admin'
import { getBrowserToken, isBrowserScopedUrl } from '@/utils/browserScope'

// 获取 API 基础地址
const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// isAdminUrl 判断是否为管理端请求
function isAdminUrl(url: string): boolean {
  return url.includes('/api/admin/')
}

// getToken 获取请求对应的 token
function getToken(url: string): string {
  if (isAdminUrl(url)) {
    const adminStore = useAdminAuthStore()
    return adminStore.adminToken || ''
  }
  const authStore = useAuthStore()
  return authStore.token || ''
}

// SSEOptions 选项
export interface SSEOptions {
  // 请求体（JSON）
  body?: unknown
  // 中断信号
  signal?: AbortSignal
}

// streamSSE 发起 SSE 流式请求
// url: 完整 API 路径（如 /api/v1/resumes/1/ai/generate）
// callbacks: 事件回调
// options: 请求选项
export async function streamSSE(
  url: string,
  callbacks: SSECallbacks,
  options: SSEOptions = {}
): Promise<void> {
  const fullUrl = url.startsWith('http') ? url : baseURL + url
  const token = getToken(url)

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'text/event-stream',
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  if (isBrowserScopedUrl(url)) {
    headers['X-Browser-Token'] = getBrowserToken()
  }

  let response: Response
  try {
    response = await fetch(fullUrl, {
      method: 'POST',
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : '{}',
      signal: options.signal,
    })
  } catch (err) {
    if ((err as Error).name === 'AbortError') {
      return
    }
    callbacks.onError?.('网络错误，请检查网络连接')
    return
  }

  if (!response.ok) {
    // 尝试读取错误信息
    let errMsg = `请求失败 (${response.status})`
    try {
      const data = await response.json()
      if (data?.message) errMsg = data.message
    } catch {
      // 非 JSON 错误
    }
    callbacks.onError?.(errMsg)
    return
  }

  if (!response.body) {
    callbacks.onError?.('响应无内容流')
    return
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let receivedTerminalEvent = false

  const handleEvent = (event: SSEEvent) => {
    if (['done', 'started', 'interview_ended', 'error'].includes(event.type)) {
      receivedTerminalEvent = true
    }
    dispatchEvent(event, callbacks)
  }

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      // 按双换行分割事件块
      const parts = buffer.split('\n\n')
      buffer = parts.pop() || ''

      for (const part of parts) {
        const line = part.trim()
        if (!line.startsWith('data:')) continue

        const jsonStr = line.slice(5).trim()
        if (!jsonStr) continue

        let event: SSEEvent
        try {
          event = JSON.parse(jsonStr)
        } catch {
          continue
        }

        handleEvent(event)
      }
    }

    // 处理缓冲区剩余内容
    if (buffer.trim().startsWith('data:')) {
      const jsonStr = buffer.trim().slice(5).trim()
      if (jsonStr) {
        try {
          const event = JSON.parse(jsonStr)
          handleEvent(event)
        } catch {
          // 忽略解析错误
        }
      }
    }
  } catch (err) {
    if ((err as Error).name !== 'AbortError') {
      callbacks.onError?.('读取流式响应失败')
    }
    return
  }

  if (!receivedTerminalEvent && !options.signal?.aborted) {
    callbacks.onError?.('响应提前结束，请重试')
  }
}

// dispatchEvent 分发 SSE 事件到对应回调
function dispatchEvent(event: SSEEvent, callbacks: SSECallbacks): void {
  switch (event.type) {
    case 'delta':
      if (event.content) callbacks.onDelta?.(event.content)
      break
    case 'status':
      if (event.content || typeof event.message === 'string') {
        callbacks.onStatus?.(event.content || (event.message as string))
      }
      break
    case 'done':
      if (callbacks.onCopilotDone) {
        callbacks.onCopilotDone({
          message: event.message as never,
          result: event.result as never,
          proposals: event.proposals,
          memory_summary: event.memory_summary,
        })
      } else {
        callbacks.onDone?.({
          message: event.message as never,
          version: event.version,
        })
      }
      break
    case 'started':
      callbacks.onStarted?.({
        message: event.message as never,
        interview: event.interview,
      })
      break
    case 'interview_ended':
      callbacks.onInterviewEnded?.({
        message: event.message as never,
        interview: event.interview,
      })
      break
    case 'error':
      callbacks.onError?.(
        (typeof event.message === 'string' ? event.message : '') ||
        event.content ||
        '未知错误'
      )
      break
  }
}

// streamSSEWithForm 以 multipart/form-data 发起 SSE 流式请求（用于语音上传）
export async function streamSSEWithForm(
  url: string,
  formData: FormData,
  callbacks: SSECallbacks,
  signal?: AbortSignal
): Promise<void> {
  const fullUrl = url.startsWith('http') ? url : baseURL + url
  const token = getToken(url)

  const headers: Record<string, string> = {
    Accept: 'text/event-stream',
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  if (isBrowserScopedUrl(url)) {
    headers['X-Browser-Token'] = getBrowserToken()
  }

  let response: Response
  try {
    response = await fetch(fullUrl, {
      method: 'POST',
      headers,
      body: formData,
      signal,
    })
  } catch (err) {
    if ((err as Error).name === 'AbortError') return
    callbacks.onError?.('网络错误，请检查网络连接')
    return
  }

  if (!response.ok) {
    let errMsg = `请求失败 (${response.status})`
    try {
      const data = await response.json()
      if (data?.message) errMsg = data.message
    } catch {
      // 非 JSON 错误
    }
    callbacks.onError?.(errMsg)
    return
  }

  if (!response.body) {
    callbacks.onError?.('响应无内容流')
    return
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const parts = buffer.split('\n\n')
      buffer = parts.pop() || ''

      for (const part of parts) {
        const line = part.trim()
        if (!line.startsWith('data:')) continue
        const jsonStr = line.slice(5).trim()
        if (!jsonStr) continue

        try {
          const event = JSON.parse(jsonStr) as SSEEvent
          dispatchEvent(event, callbacks)
        } catch {
          continue
        }
      }
    }
  } catch (err) {
    if ((err as Error).name !== 'AbortError') {
      callbacks.onError?.('读取流式响应失败')
    }
  }
}
