import { del, get, put } from '@/utils/request'
import type { ApiResponse } from '@/types/models'

const COPILOT_STATE_KEY = 'copilot_sessions'

interface BrowserState<T> {
  key: string
  value: T
}

export const getCopilotSessionsState = <T>() =>
  get<ApiResponse<BrowserState<T>>>(`/api/v1/browser-state/${COPILOT_STATE_KEY}`)

export const putCopilotSessionsState = (value: unknown) =>
  put<ApiResponse<{ key: string }>>(`/api/v1/browser-state/${COPILOT_STATE_KEY}`, { value })

export const deleteCopilotSessionsState = () =>
  del<ApiResponse<{ key: string }>>(`/api/v1/browser-state/${COPILOT_STATE_KEY}`)
