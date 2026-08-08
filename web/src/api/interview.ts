import { get, post, patch, del } from '@/utils/request'
import { streamSSE, streamSSEWithForm } from '@/utils/sse'
import type {
  ApiResponse,
  Interview,
  InterviewDetail,
  CreateInterviewRequest,
  AttachResumeRequest,
  InterviewMessage,
  InterviewReport,
  InterviewScore,
  SSECallbacks,
} from '@/types/models'

// 获取面试列表（无分页，返回数组）
export const listInterviews = () => {
  return get<ApiResponse<Interview[]>>('/api/v1/interviews')
}

// 兼容旧名称
export const getInterviews = listInterviews

// 获取面试详情（含消息列表）
export const getInterview = (id: number) => {
  return get<ApiResponse<InterviewDetail>>(`/api/v1/interviews/${id}`)
}

// 删除一条面试记录（含消息、评分、复盘）
export const deleteInterview = (id: number) => {
  return del<ApiResponse<{ id: number }>>(`/api/v1/interviews/${id}`)
}

// 创建面试会话
export const createInterview = (data: CreateInterviewRequest) => {
  return post<ApiResponse<Interview>>('/api/v1/interviews', data)
}

// 启动准备中的面试，首题会结合已绑定简历和 JD 通过 SSE 生成
export const startInterview = (
  interviewId: number,
  callbacks: SSECallbacks,
  signal?: AbortSignal
) => {
  return streamSSE(`/api/v1/interviews/${interviewId}/start`, callbacks, { signal })
}

// 面试中切换交互模式
export const setInterviewMode = (interviewId: number, mode: string) => {
  return patch<ApiResponse<Interview>>(
    `/api/v1/interviews/${interviewId}/mode`,
    { mode }
  )
}

// 在面试中发送简历（把指定简历版本的快照绑定到面试会话）
export const attachResume = (interviewId: number, data: AttachResumeRequest) => {
  return post<ApiResponse<Interview>>(
    `/api/v1/interviews/${interviewId}/resume`,
    data
  )
}

// 发送文字回答（SSE 流式）
export const sendMessage = (
  interviewId: number,
  content: string,
  callbacks: SSECallbacks,
  signal?: AbortSignal
) => {
  return streamSSE(
    `/api/v1/interviews/${interviewId}/messages`,
    callbacks,
    { body: { content }, signal }
  )
}

// 发送语音回答（SSE 流式，multipart/form-data）
export const sendVoice = (
  interviewId: number,
  audio: File,
  durationSeconds: number,
  callbacks: SSECallbacks,
  signal?: AbortSignal
) => {
  const formData = new FormData()
  formData.append('audio', audio)
  formData.append('duration_sec', String(Math.max(0, Math.round(durationSeconds))))
  return streamSSEWithForm(
    `/api/v1/interviews/${interviewId}/voice`,
    formData,
    callbacks,
    signal
  )
}

// 仅转写语音草稿，不创建回答消息、不推进面试
export const transcribeVoice = (interviewId: number, audio: File) => {
  const formData = new FormData()
  formData.append('audio', audio)
  return post<ApiResponse<{ text: string }>>(
    `/api/v1/interviews/${interviewId}/transcribe`,
    formData,
    {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000,
    }
  )
}

// 获取 AI 提问的 TTS 音频（返回二进制 Blob）
export const getTtsAudio = (interviewId: number, msgId: number) => {
  return get<Blob>(
    `/api/v1/interviews/${interviewId}/tts/${msgId}`,
    { responseType: 'blob' }
  ).then((resp) => resp.data)
}

// 结束面试并生成复盘
export const endInterview = (interviewId: number) => {
  return post<ApiResponse<InterviewReport>>(
    `/api/v1/interviews/${interviewId}/end`
  )
}

// 取消尚未开始的面试
export const cancelInterview = (interviewId: number) => {
  return post<ApiResponse<Interview>>(
    `/api/v1/interviews/${interviewId}/cancel`
  )
}

// 获取复盘报告
export const getReport = (interviewId: number) => {
  return get<ApiResponse<InterviewReport>>(
    `/api/v1/interviews/${interviewId}/report`
  )
}

// 获取评分明细
export const getScores = (interviewId: number) => {
  return get<ApiResponse<InterviewScore[]>>(
    `/api/v1/interviews/${interviewId}/scores`
  )
}
