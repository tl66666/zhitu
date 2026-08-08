import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
  Interview,
  InterviewMessage,
  InterviewReport,
  InterviewScore,
  CreateInterviewRequest,
  AttachResumeRequest,
  InterviewMode,
} from '@/types/models'
import * as interviewApi from '@/api/interview'
import { message } from 'ant-design-vue'

export const useInterviewStore = defineStore('interview', () => {
  // 状态
  const interviews = ref<Interview[]>([])
  const currentInterview = ref<Interview | null>(null)
  const messages = ref<InterviewMessage[]>([])
  const report = ref<InterviewReport | null>(null)
  const scores = ref<InterviewScore[]>([])
  const loading = ref(false)
  const sending = ref(false)
  const starting = ref(false)
  const ending = ref(false)
  const streamingText = ref('')

  // 获取面试列表（无分页）
  const fetchList = async () => {
    loading.value = true
    try {
      const response = await interviewApi.listInterviews()
      interviews.value = response.data.data || []
      return true
    } catch (error) {
      console.error('获取面试列表失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  // 获取面试详情（含消息列表）
  const fetchOne = async (id: number) => {
    loading.value = true
    try {
      const response = await interviewApi.getInterview(id)
      const detail = response.data.data
      if (detail) {
        currentInterview.value = detail.interview
        messages.value = detail.messages || []
      }
      return true
    } catch (error) {
      console.error('获取面试详情失败:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  // 创建面试会话
  const create = async (data: CreateInterviewRequest) => {
    try {
      const response = await interviewApi.createInterview(data)
      const i = response.data.data
      if (i) interviews.value.unshift(i)
      message.success('面试会话创建成功')
      return i || null
    } catch (error) {
      console.error('创建面试失败:', error)
      return null
    }
  }

  // 启动准备中的面试并接收基于简历/JD 的首题
  const startInterview = async (id: number, onDelta?: (delta: string) => void) => {
    starting.value = true
    streamingText.value = ''
    let started = false
    if (currentInterview.value?.id === id) {
      currentInterview.value.status = 'starting'
      currentInterview.value.status_message = ''
    }
    try {
      await interviewApi.startInterview(id, {
        onDelta: (delta) => {
          streamingText.value += delta
          onDelta?.(delta)
        },
        onStarted: (data) => {
          started = true
          streamingText.value = ''
          if (data.interview) currentInterview.value = data.interview
          if (data.message) messages.value = [data.message]
        },
        onError: (errMsg) => message.error(errMsg || '面试启动失败，请重试'),
      })
      if (!started) await fetchOne(id)
      return started
    } catch (error) {
      console.error('启动面试失败:', error)
      return false
    } finally {
      starting.value = false
      streamingText.value = ''
    }
  }

  const setMode = async (id: number, mode: InterviewMode) => {
    try {
      const response = await interviewApi.setInterviewMode(id, mode)
      const updated = response.data.data
      if (updated && currentInterview.value?.id === id) currentInterview.value = updated
      return updated || null
    } catch (error) {
      console.error('切换面试模式失败:', error)
      return null
    }
  }

  // 在面试中发送简历（绑定简历版本快照，AI 后续提问会结合简历内容）
  const attachResume = async (interviewId: number, data: AttachResumeRequest) => {
    try {
      const response = await interviewApi.attachResume(interviewId, data)
      const updated = response.data.data
      if (updated && currentInterview.value?.id === interviewId) {
        currentInterview.value = updated
      }
      message.success('简历已发送')
      return updated || null
    } catch (error) {
      console.error('发送简历失败:', error)
      return null
    }
  }

  // 发送文字回答（SSE 流式）
  // onDelta: AI 回复增量文本回调
  // 返回是否成功
  const sendMessage = async (
    id: number,
    content: string,
    onDelta?: (delta: string) => void,
    signal?: AbortSignal
  ) => {
    sending.value = true
    streamingText.value = ''
    // 先把用户消息加入列表（乐观更新）
    const userMsg: InterviewMessage = {
      id: Date.now(),
      interview_id: id,
      role: 'user',
      content,
      audio_url: '',
      input_mode: 'text',
      question_type: '',
      question_no: 0,
      duration_sec: 0,
      created_at: new Date().toISOString(),
    }
    messages.value.push(userMsg)

    let succeeded = false
    try {
      await interviewApi.sendMessage(
        id,
        content,
        {
          onDelta: (delta) => {
            streamingText.value += delta
            onDelta?.(delta)
          },
          onDone: (data) => {
            succeeded = true
            // 流结束，清空临时流文本，把 AI 消息加入列表
            streamingText.value = ''
            if (data.message) {
              messages.value.push(data.message)
              if (currentInterview.value) {
                currentInterview.value.current_question_no = data.message.question_no
              }
            }
          },
          onInterviewEnded: (data) => {
            succeeded = true
            if (data.message) {
              messages.value.push(data.message)
            }
            if (data.interview) {
              currentInterview.value = data.interview
            }
            message.info('面试已结束')
          },
          onError: (errMsg) => {
            message.error(errMsg || '发送失败')
          },
        },
        signal
      )
      if (!succeeded) {
        const optimisticIndex = messages.value.findIndex((item) => item.id === userMsg.id)
        if (optimisticIndex >= 0) messages.value.splice(optimisticIndex, 1)
        await fetchOne(id)
      }
      return succeeded
    } catch (error) {
      console.error('发送消息失败:', error)
      const optimisticIndex = messages.value.findIndex((item) => item.id === userMsg.id)
      if (optimisticIndex >= 0) messages.value.splice(optimisticIndex, 1)
      await fetchOne(id)
      return false
    } finally {
      sending.value = false
    }
  }

  // 发送语音回答（SSE 流式，multipart/form-data）
  const sendVoice = async (
    id: number,
    audio: File,
    durationSeconds = 0,
    onDelta?: (delta: string) => void,
    signal?: AbortSignal
  ) => {
    sending.value = true
    streamingText.value = ''
    let succeeded = false
    // 乐观加入用户消息占位（语音转写中）
    const userMsg: InterviewMessage = {
      id: Date.now(),
      interview_id: id,
      role: 'user',
      content: '（语音转写中...）',
      audio_url: '',
      input_mode: 'voice',
      question_type: '',
      question_no: 0,
      duration_sec: 0,
      created_at: new Date().toISOString(),
    }
    messages.value.push(userMsg)

    try {
      await interviewApi.sendVoice(
        id,
        audio,
        durationSeconds,
        {
          onDelta: (delta) => {
            streamingText.value += delta
            onDelta?.(delta)
          },
          onStatus: (msg) => {
            // 语音转写状态更新
            const last = messages.value[messages.value.length - 1]
            if (last && last.role === 'user') {
              last.content = msg
            }
          },
          onDone: (data) => {
            succeeded = true
            streamingText.value = ''
            if (data.message) {
              messages.value.push(data.message)
            }
          },
          onInterviewEnded: (data) => {
            succeeded = true
            if (data.message) {
              messages.value.push(data.message)
            }
            if (data.interview) {
              currentInterview.value = data.interview
            }
            message.info('面试已结束')
          },
          onError: (errMsg) => {
            message.error(errMsg || '语音发送失败')
          },
        },
        signal
      )
      if (succeeded) {
        // 用服务端保存的转写内容替换本地占位，并同步面试完成状态。
        await fetchOne(id)
      } else {
        const placeholderIndex = messages.value.findIndex((item) => item.id === userMsg.id)
        if (placeholderIndex >= 0) messages.value.splice(placeholderIndex, 1)
      }
      return succeeded
    } catch (error) {
      console.error('发送语音失败:', error)
      const placeholderIndex = messages.value.findIndex((item) => item.id === userMsg.id)
      if (placeholderIndex >= 0) messages.value.splice(placeholderIndex, 1)
      await fetchOne(id)
      return false
    } finally {
      sending.value = false
    }
  }

  // 获取 TTS 音频，返回 Blob URL（调用方负责 revokeObjectURL）
  const playTts = async (interviewId: number, msgId: number) => {
    try {
      const blob = await interviewApi.getTtsAudio(interviewId, msgId)
      const url = URL.createObjectURL(blob)
      return url
    } catch (error) {
      console.error('获取 TTS 音频失败:', error)
      message.error('获取音频失败')
      return null
    }
  }

  // 结束面试并生成复盘
  const endInterview = async (id: number) => {
    ending.value = true
    if (currentInterview.value?.id === id) {
      currentInterview.value.status = 'reviewing'
      currentInterview.value.status_message = ''
    }
    try {
      const response = await interviewApi.endInterview(id)
      const r = response.data.data
      if (r) report.value = r
      if (currentInterview.value) {
        currentInterview.value.status = 'completed'
        currentInterview.value.ended_at = new Date().toISOString()
      }
      message.success('面试已结束，复盘报告已生成')
      return r || null
    } catch (error) {
      console.error('结束面试失败:', error)
      await fetchOne(id)
      return null
    } finally {
      ending.value = false
    }
  }

  const cancelInterview = async (id: number) => {
    try {
      const response = await interviewApi.cancelInterview(id)
      const cancelled = response.data.data
      if (cancelled && currentInterview.value?.id === id) currentInterview.value = cancelled
      const listItem = interviews.value.find((item) => item.id === id)
      if (cancelled && listItem) Object.assign(listItem, cancelled)
      message.success('本次面试已取消')
      return cancelled || null
    } catch (error) {
      console.error('取消面试失败:', error)
      await fetchOne(id)
      return null
    }
  }

  // 删除一条面试记录（含消息、评分、复盘）
  const removeInterview = async (id: number) => {
    try {
      await interviewApi.deleteInterview(id)
      interviews.value = interviews.value.filter((item) => item.id !== id)
      if (currentInterview.value?.id === id) clearCurrent()
      message.success('面试记录已删除')
      return true
    } catch (error) {
      console.error('删除面试记录失败:', error)
      return false
    }
  }

  // 获取复盘报告
  const fetchReport = async (id: number) => {
    try {
      const response = await interviewApi.getReport(id)
      report.value = response.data.data
      return true
    } catch (error) {
      console.error('获取复盘报告失败:', error)
      return false
    }
  }

  // 获取评分明细
  const fetchScores = async (id: number) => {
    try {
      const response = await interviewApi.getScores(id)
      scores.value = response.data.data || []
      return true
    } catch (error) {
      console.error('获取评分明细失败:', error)
      return false
    }
  }

  // 清空当前面试
  const clearCurrent = () => {
    currentInterview.value = null
    messages.value = []
    report.value = null
    scores.value = []
    streamingText.value = ''
  }

  return {
    // 状态
    interviews,
    currentInterview,
    messages,
    report,
    scores,
    loading,
    sending,
    starting,
    ending,
    streamingText,
    // 操作
    fetchList,
    fetchOne,
    create,
    startInterview,
    setMode,
    attachResume,
    sendMessage,
    sendVoice,
    playTts,
    endInterview,
    cancelInterview,
    removeInterview,
    fetchReport,
    fetchScores,
    clearCurrent,
  }
})
