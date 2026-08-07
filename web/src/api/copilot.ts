import { post } from '@/utils/request'
import { streamSSE } from '@/utils/sse'
import type { CopilotChatRequest, CopilotProposal, CopilotResponse, SSECallbacks } from '@/types/models'

export const chatCopilot = (
  input: CopilotChatRequest,
  callbacks: SSECallbacks,
  signal?: AbortSignal,
) => streamSSE('/api/v1/copilot/chat', callbacks, { body: input, signal })

export const applyCopilotProposal = (input: {
  resume_id: number
  base_version_id: number
  content: string
  change_note?: string
  project_index?: number
  replacement_description?: string
  replacement_tech_stack?: string[]
}) => {
  return post('/api/v1/copilot/apply', input)
}

export type { CopilotProposal, CopilotResponse }
