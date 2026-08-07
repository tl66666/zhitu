import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { message } from 'ant-design-vue'
import { chatCopilot } from '@/api/copilot'
import type {
  CopilotChatRequest,
  CopilotMessage,
  CopilotResponse,
  CopilotSession,
  CopilotTask,
} from '@/types/models'

const DB_NAME = 'zhitu-copilot'
const DB_VERSION = 1
const STORE_NAME = 'sessions'
const COOKIE_NAME = 'zhitu-copilot-session'
const FALLBACK_KEY = 'zhitu-copilot-sessions'

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

const openDB = (): Promise<IDBDatabase | null> => new Promise((resolve) => {
  if (typeof indexedDB === 'undefined') {
    resolve(null)
    return
  }
  const request = indexedDB.open(DB_NAME, DB_VERSION)
  request.onupgradeneeded = () => {
    if (!request.result.objectStoreNames.contains(STORE_NAME)) request.result.createObjectStore(STORE_NAME, { keyPath: 'id' })
  }
  request.onsuccess = () => resolve(request.result)
  request.onerror = () => resolve(null)
})

const readSessions = async (): Promise<CopilotSession[]> => {
  const db = await openDB()
  if (!db) {
    try { return JSON.parse(localStorage.getItem(FALLBACK_KEY) || '[]') as CopilotSession[] } catch { return [] }
  }
  return new Promise((resolve) => {
    const request = db.transaction(STORE_NAME, 'readonly').objectStore(STORE_NAME).getAll()
    request.onsuccess = () => resolve((request.result || []) as CopilotSession[])
    request.onerror = () => resolve([])
  })
}

const writeSession = async (session: CopilotSession) => {
  const db = await openDB()
  if (!db) {
    const sessions = await readSessions()
    const next = [session, ...sessions.filter((item) => item.id !== session.id)].slice(0, 30)
    localStorage.setItem(FALLBACK_KEY, JSON.stringify(next))
    return
  }
  await new Promise<void>((resolve) => {
    const request = db.transaction(STORE_NAME, 'readwrite').objectStore(STORE_NAME).put(session)
    request.onsuccess = () => resolve()
    request.onerror = () => resolve()
  })
}

const deleteSession = async (id: string) => {
  const db = await openDB()
  if (!db) {
    const sessions = await readSessions()
    localStorage.setItem(FALLBACK_KEY, JSON.stringify(sessions.filter((item) => item.id !== id)))
    return
  }
  await new Promise<void>((resolve) => {
    const request = db.transaction(STORE_NAME, 'readwrite').objectStore(STORE_NAME).delete(id)
    request.onsuccess = () => resolve()
    request.onerror = () => resolve()
  })
}

export const useCopilotStore = defineStore('copilot', () => {
  const sessions = ref<CopilotSession[]>([])
  const activeSessionId = ref('')
  const loading = ref(false)
  const initialized = ref(false)
  const activeSession = computed(() => sessions.value.find((item) => item.id === activeSessionId.value) || null)

  const persist = async (session: CopilotSession) => {
    const index = sessions.value.findIndex((item) => item.id === session.id)
    if (index >= 0) sessions.value[index] = { ...session }
    else sessions.value.unshift(session)
    sessions.value.sort((a, b) => b.updated_at.localeCompare(a.updated_at))
    await writeSession(session)
    setCookieValue(session.id)
  }

  const init = async () => {
    if (initialized.value) return
    sessions.value = (await readSessions()).sort((a, b) => b.updated_at.localeCompare(a.updated_at))
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
        onCopilotDone: (data) => {
          const result = data.result
          if (!result) return
          const assistant: CopilotMessage = {
            role: 'assistant',
            content: data.message?.content || result.reply,
            created_at: new Date().toISOString(),
            result,
          }
          session.messages.push(assistant)
          session.summary = data.memory_summary || result.memory_summary || session.summary
          session.updated_at = new Date().toISOString()
          void persist(session)
          success = true
          input.onResult?.(result)
        },
        onError: (error) => message.error(error || 'Copilot 暂时无法回答'),
      })
      return success
    } catch (error) {
      console.error('Copilot 请求失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  const remove = async (id: string) => {
    await deleteSession(id)
    sessions.value = sessions.value.filter((item) => item.id !== id)
    if (activeSessionId.value === id) {
      activeSessionId.value = sessions.value[0]?.id || ''
      if (activeSessionId.value) setCookieValue(activeSessionId.value)
    }
  }

  const clearAll = async () => {
    for (const session of sessions.value) await deleteSession(session.id)
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
