import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { message } from 'ant-design-vue'
import { chatCopilot } from '@/api/copilot'
import {
  deleteCopilotSessionsState,
  getCopilotSessionsState,
  putCopilotSessionsState,
} from '@/api/browserState'
import type {
  CopilotChatRequest,
  CopilotMessage,
  CopilotResponse,
  CopilotSession,
  CopilotTask,
} from '@/types/models'

const COOKIE_NAME = 'zhitu-copilot-session'

const toPersistedSession = (session: CopilotSession): CopilotSession =>
  JSON.parse(JSON.stringify(session)) as CopilotSession

const cookieValue = () => {
  if (typeof document === 'undefined') return ''
  return document.cookie
    .split('; ')
    .find((item) => item.startsWith(`${COOKIE_NAME}=`))
    ?.slice(COOKIE_NAME.length + 1) || ''
}

const setCookieValue = (value: string) => {
  if (typeof document !== 'undefined') {
    document.cookie = `${COOKIE_NAME}=${encodeURIComponent(value)}; Max-Age=2592000; Path=/; SameSite=Lax`
  }
}

const uuid = () => {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return `copilot-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

const readSessions = async (): Promise<CopilotSession[]> => {
  const response = await getCopilotSessionsState<CopilotSession[]>()
  return Array.isArray(response.data.data.value) ? response.data.data.value : []
}

const writeSessions = async (sessions: CopilotSession[]) => {
  const snapshots = sessions.slice(0, 30).map(toPersistedSession)
  await putCopilotSessionsState(snapshots)
}

export const useCopilotStore = defineStore('copilot', () => {
  const sessions = ref<CopilotSession[]>([])
  const activeSessionId = ref('')
  const loading = ref(false)
  const initialized = ref(false)
  const activeSession = computed(() => sessions.value.find((item) => item.id === activeSessionId.value) || null)

  const persist = async (session: CopilotSession) => {
    const snapshot = toPersistedSession(session)
    const index = sessions.value.findIndex((item) => item.id === snapshot.id)
    if (index >= 0) sessions.value[index] = snapshot
    else sessions.value.unshift(snapshot)
    sessions.value.sort((a, b) => b.updated_at.localeCompare(a.updated_at))
    await writeSessions(sessions.value)
    setCookieValue(snapshot.id)
  }

  const init = async () => {
    if (initialized.value) return
    try {
      sessions.value = (await readSessions()).sort((a, b) => b.updated_at.localeCompare(a.updated_at))
    } catch (error) {
      console.error('加载 Copilot 历史失败:', error)
      message.error('加载 Copilot 历史失败，请刷新重试')
      sessions.value = []
    }
    const remembered = cookieValue()
    activeSessionId.value = sessions.value.some((item) => item.id === remembered)
      ? remembered
      : (sessions.value[0]?.id || '')
    initialized.value = true
  }

  const createSession = async (input: {
    task: CopilotTask
    resumeId?: number | null
    versionId?: number | null
    jd?: string
    draftContent?: string
    projectIndex?: number
  }) => {
    await init()
    const now = new Date().toISOString()
    const session: CopilotSession = {
      id: uuid(),
      resume_id: input.resumeId ?? null,
      version_id: input.versionId ?? null,
      task: input.task,
      jd: input.jd || '',
      draft_content: input.draftContent || '',
      project_index: input.projectIndex,
      messages: [],
      summary: '',
      created_at: now,
      updated_at: now,
    }
    activeSessionId.value = session.id
    await persist(session)
    return session
  }

  const selectSession = (id: string) => {
    if (sessions.value.some((item) => item.id === id)) {
      activeSessionId.value = id
      setCookieValue(id)
    }
  }

  const send = async (input: {
    session: CopilotSession
    content: string
    draftContent?: string
    onResult?: (result: CopilotResponse) => void
  }) => {
    const session = input.session
    if (input.draftContent !== undefined) session.draft_content = input.draftContent
    const userMessage: CopilotMessage = {
      role: 'user', content: input.content, created_at: new Date().toISOString(),
    }
    session.messages.push(userMessage)
    session.updated_at = new Date().toISOString()
    await persist(session)
    loading.value = true
    let success = false
    let pendingPersist: Promise<void> = Promise.resolve()
    let streamingMessage: CopilotMessage | null = null
    try {
      const request: CopilotChatRequest = {
        task: session.task,
        resume_id: session.resume_id || 0,
        version_id: session.version_id || undefined,
        jd: session.jd,
        project_index: session.project_index,
        draft_content: input.draftContent || session.draft_content,
        messages: session.messages.slice(-24).map(({ role, content }) => ({ role, content })),
      }
      await chatCopilot(request, {
        onStatus: () => undefined,
        onDelta: (delta) => {
          if (!streamingMessage) {
            streamingMessage = {
              role: 'assistant',
              content: '',
              created_at: new Date().toISOString(),
            }
            session.messages.push(streamingMessage)
          }
          streamingMessage.content += delta
          session.updated_at = new Date().toISOString()
        },
        onCopilotDone: (data) => {
          const result = data.result
          const reply = data.message?.content || result?.reply || ''
          if (!reply && !result) return
          if (streamingMessage) {
            streamingMessage.content = reply || streamingMessage.content || '分析已完成，请查看下方结果。'
            streamingMessage.result = result
          } else {
            session.messages.push({
              role: 'assistant',
              content: reply || '分析已完成，请查看下方结果。',
              created_at: new Date().toISOString(),
              result,
            })
          }
          session.summary = data.memory_summary || result?.memory_summary || session.summary
          session.updated_at = new Date().toISOString()
          pendingPersist = persist(session)
          success = true
          if (result) input.onResult?.(result)
        },
        onError: (error) => {
          if (streamingMessage) {
            streamingMessage.content = error || 'Copilot 暂时无法回答'
          } else {
            session.messages.push({
              role: 'assistant',
              content: error || 'Copilot 暂时无法回答',
              created_at: new Date().toISOString(),
            })
          }
          session.updated_at = new Date().toISOString()
          pendingPersist = persist(session)
          message.error(error || 'Copilot 暂时无法回答')
        },
      })
      await pendingPersist
      return success
    } catch (error) {
      console.error('Copilot 请求失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  const remove = async (id: string) => {
    sessions.value = sessions.value.filter((item) => item.id !== id)
    await writeSessions(sessions.value)
    if (activeSessionId.value === id) {
      activeSessionId.value = sessions.value[0]?.id || ''
      if (activeSessionId.value) setCookieValue(activeSessionId.value)
    }
  }

  const clearAll = async () => {
    await deleteCopilotSessionsState()
    sessions.value = []
    activeSessionId.value = ''
    setCookieValue('')
  }

  return {
    sessions,
    activeSession,
    activeSessionId,
    loading,
    initialized,
    init,
    createSession,
    selectSession,
    send,
    remove,
    clearAll,
  }
})
