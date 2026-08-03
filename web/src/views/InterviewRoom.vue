<template src="./templates/InterviewRoom.html"></template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Modal, message, Empty, type TableColumnsType } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  PoweroffOutlined,
  MessageOutlined,
  SoundOutlined,
  AudioOutlined,
  SendOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  BulbOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons-vue'
import { useInterviewStore } from '@/stores/interview'
import { transcribeVoice } from '@/api/interview'
import { listResumes, listVersions } from '@/api/resume'
import type {
  InterviewScene,
  InterviewMode,
  InterviewDifficulty,
  InterviewStatus,
  InterviewDimension,
  InterviewMessage,
  QuestionFeedbackItem,
  Resume,
  ResumeVersion,
} from '@/types/models'

const route = useRoute()
const router = useRouter()
const interviewStore = useInterviewStore()

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const escapeMessageHtml = (content: string): string =>
  content.replace(/[&<>"']/g, (character) => {
    const entities: Record<string, string> = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;',
    }
    return entities[character]
  })

const plainMessageContent = (content: string): string =>
  content.replace(/\*\*([\s\S]*?)\*\*/g, '$1').replace(/\*\*/g, '')

const renderMessageContent = (content: string): string =>
  escapeMessageHtml(content)
    .replace(/\*\*([\s\S]*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*\*/g, '')
    .replace(/\r?\n/g, '<br>')

const interviewId = computed(() => Number(route.params.id))

// 输入与界面状态
const inputText = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const sideTab = ref<'info' | 'report' | 'scores'>('info')

// 麦克风录音状态
const isRecording = ref(false)
const voicePreparing = ref(false)
const recordingFinalizing = ref(false)
const recordingSeconds = ref(0)
const transcribingDraft = ref(false)
type AudioDraft = {
  file: File
  objectUrl: string
  durationSeconds: number
}
const audioDraft = ref<AudioDraft | null>(null)
const recordingTimeLabel = computed(() => {
  const minutes = Math.floor(recordingSeconds.value / 60)
  const seconds = recordingSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})
let microphoneStream: MediaStream | null = null
let recordingTimer: number | null = null
let audioContext: AudioContext | null = null
let microphoneSource: MediaStreamAudioSourceNode | null = null
let recordingProcessor: ScriptProcessorNode | null = null
let silentGain: GainNode | null = null
let recordedPCMChunks: Float32Array[] = []
let recordingSampleRate = 44100
let discardRecordedVoice = false

// 加载状态
const reportLoading = ref(false)
const scoresLoading = ref(false)
// TTS 播放状态（调后端 xiaomi mimo TTS → blob → <audio> 播放）
const ttsLoadingId = ref<number | null>(null)
const playingMessageId = ref<number | null>(null)
const audioPlayer = ref<HTMLAudioElement | null>(null)
let ttsObjectUrl: string | null = null

// 发送简历相关状态
const resumeModalVisible = ref(false)
const loadingResumes = ref(false)
const loadingVersions = ref(false)
const attaching = ref(false)
const resumeList = ref<Resume[]>([])
const versionList = ref<ResumeVersion[]>([])
const selectedResumeId = ref<number | null>(null)
const selectedVersionId = ref<number>(0)
const resumeAttached = computed(
  () => !!interviewStore.currentInterview?.resume_snapshot
)

// 计算属性
const isOngoing = computed(
  () => interviewStore.currentInterview?.status === 'ongoing'
)
const isCompleted = computed(
  () => interviewStore.currentInterview?.status === 'completed'
)
const canSendVoice = computed(() => {
  const mode = interviewStore.currentInterview?.mode
  return mode === 'voice' || mode === 'hybrid'
})
const canPlayTts = computed(() => {
  const mode = interviewStore.currentInterview?.mode
  return mode === 'voice' || mode === 'hybrid'
})
const isTeachingScene = computed(() => interviewStore.currentInterview?.scene === 'teaching')
const latestQuestion = computed(() => {
  const assistantMessages = interviewStore.messages.filter((item) => item.role === 'assistant')
  const content = assistantMessages[assistantMessages.length - 1]?.content
  return content ? plainMessageContent(content) : '考官正在准备第一道结构化问题…'
})
const currentTeachingPhase = computed(() => {
  const no = interviewStore.currentInterview?.current_question_no || 1
  if (no <= 2) return '结构化问答'
  if (no <= 4) return '模拟试讲'
  return '考官答辩'
})
const teachingTimeHint = computed(() => currentTeachingPhase.value === '模拟试讲' ? '8 分钟' : '2 分钟')

// 显示的消息列表（隐藏流式时的临时占位）
const displayMessages = computed(() => interviewStore.messages)

// 自动滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(
  () => interviewStore.messages.length,
  () => scrollToBottom()
)
watch(
  () => interviewStore.streamingText,
  () => scrollToBottom()
)

// 场景
const sceneLabel = (s: InterviewScene | string): string => {
  const map: Record<string, string> = {
    teaching: '模拟教室',
    corporate: '企业会议室',
    group: '群面讨论室',
    defense: '项目答辩室',
    client: '客户会议室',
    pressure: '压力面试室',
    public: '结构化面试厅',
    medical: '医疗面试室',
    media: '媒体演播室',
    remote: '远程面试间',
    system: '系统设计室',
    aviation: '航空面试厅',
    tech: '技术面',
    behavior: '行为面',
    hr: 'HR 面',
  }
  return map[s] || s
}
const sceneColor = (s: InterviewScene | string): string => {
  const map: Record<string, string> = {
    teaching: 'green',
    corporate: 'blue',
    group: 'orange',
    defense: 'geekblue',
    client: 'cyan',
    pressure: 'red',
    public: 'gold',
    medical: 'green',
    media: 'magenta',
    remote: 'purple',
    system: 'volcano',
    aviation: 'lime',
    tech: 'blue',
    behavior: 'cyan',
    hr: 'purple',
  }
  return map[s] || 'default'
}

// 难度
const difficultyLabel = (d: InterviewDifficulty | string): string => {
  const map: Record<string, string> = {
    junior: '初级',
    mid: '中级',
    senior: '高级',
    mixed: '混合',
  }
  return map[d] || d
}

// 模式
const modeLabel = (m: InterviewMode | string): string => {
  const map: Record<string, string> = {
    text: '文字',
    voice: '语音',
    hybrid: '混合',
  }
  return map[m] || m
}
const modeColor = (m: InterviewMode | string): string => {
  const map: Record<string, string> = {
    text: 'default',
    voice: 'green',
    hybrid: 'geekblue',
  }
  return map[m] || 'default'
}

// 状态
const statusLabel = (s: InterviewStatus | string): string => {
  const map: Record<string, string> = {
    ongoing: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  }
  return map[s] || s
}
const statusBadge = (s: InterviewStatus | string): 'success' | 'processing' | 'default' | 'error' => {
  const map: Record<string, 'success' | 'processing' | 'default' | 'error'> = {
    ongoing: 'processing',
    completed: 'success',
    cancelled: 'default',
  }
  return map[s] || 'default'
}

// 消息角色
const roleLabel = (r: string): string => (r === 'assistant' ? 'AI 面试官' : '我')
const avatarText = (r: string): string => (r === 'assistant' ? 'AI' : '我')
const avatarStyle = (r: string): Record<string, string> => {
  if (r === 'assistant') return { background: '#1890ff', color: '#fff' }
  return { background: '#52c41a', color: '#fff' }
}

// 维度
const dimensionLabel = (d: InterviewDimension | string): string => {
  const map: Record<string, string> = {
    professional: '专业能力',
    expression: '表达能力',
    logic: '逻辑思维',
    adaptability: '应变能力',
    pace: '节奏掌控',
  }
  return map[d] || d
}

// 评分相关样式
const scoreStyle = (score: number): Record<string, string> => {
  if (score >= 80) return { color: '#52c41a' }
  if (score >= 60) return { color: '#faad14' }
  return { color: '#f5222d' }
}
const scoreTagColor = (score: number): string => {
  if (score >= 80) return 'green'
  if (score >= 60) return 'orange'
  return 'red'
}

// 评分表列
const scoreColumns: TableColumnsType = [
  { title: '维度', key: 'dimension', width: 120 },
  { title: '分数', key: 'score', width: 80, align: 'center' },
  { title: '评语', dataIndex: 'comment', key: 'comment' },
]

// JSON 数组解析（容错）
const parseJsonArray = (str: string | null | undefined): string[] => {
  if (!str) return []
  try {
    const parsed = JSON.parse(str)
    if (Array.isArray(parsed)) return parsed.map((x) => String(x))
    if (typeof parsed === 'string') return [parsed]
    return [JSON.stringify(parsed)]
  } catch {
    // 不是 JSON，按行分割
    return str.split('\n').filter((s) => s.trim())
  }
}

const parseQuestionFeedback = (
  value: string | null | undefined
): QuestionFeedbackItem[] => {
  if (!value) return []
  try {
    const parsed = JSON.parse(value)
    if (!Array.isArray(parsed)) return []
    return parsed
      .map((item, index) => {
        const rawScore = Number(item?.score)
        return {
          question_no: Number(item?.question_no) || index + 1,
          question: String(item?.question || '').trim(),
          answer: String(item?.answer || '').trim(),
          score: Number.isFinite(rawScore)
            ? Math.max(0, Math.min(100, Math.round(rawScore)))
            : 0,
          comment: String(item?.comment || '').trim(),
          suggestion: String(item?.suggestion || '').trim(),
        }
      })
      .filter((item) =>
        Boolean(item.question || item.answer || item.comment || item.suggestion)
      )
      .sort((a, b) => a.question_no - b.question_no)
  } catch {
    return []
  }
}

const questionFeedbackItems = computed(() =>
  parseQuestionFeedback(interviewStore.report?.question_feedback)
)

// 时间格式化
const formatTime = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return d.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// 返回列表
const backToList = () => {
  interviewStore.clearCurrent()
  router.push('/app/interviews')
}

// ============ 发送简历相关 ============

// 前往简历实验室（无简历时）
const goToResumeLab = () => {
  resumeModalVisible.value = false
  router.push('/app/resumes')
}

// 版本时间格式化（用于下拉展示）
const formatVersionTime = (dateStr: string): string => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return d.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 拉取用户简历列表
const loadResumes = async () => {
  loadingResumes.value = true
  try {
    const resp = await listResumes()
    resumeList.value = resp.data.data || []
    // 默认选中第一份简历
    if (resumeList.value.length > 0 && selectedResumeId.value === null) {
      selectedResumeId.value = resumeList.value[0].id
      await loadVersions(resumeList.value[0].id)
    }
  } catch (error) {
    console.error('加载简历列表失败:', error)
    message.error('加载简历列表失败')
  } finally {
    loadingResumes.value = false
  }
}

// 选择简历时加载版本列表
const onResumeChange = async (resumeId: number) => {
  selectedVersionId.value = 0
  versionList.value = []
  if (!resumeId) return
  await loadVersions(resumeId)
}

const loadVersions = async (resumeId: number) => {
  loadingVersions.value = true
  try {
    const resp = await listVersions(resumeId)
    versionList.value = resp.data.data || []
  } catch (error) {
    console.error('加载简历版本失败:', error)
  } finally {
    loadingVersions.value = false
  }
}

// 打开发送简历弹窗
const openResumeModal = async () => {
  resumeModalVisible.value = true
  if (resumeList.value.length === 0) {
    await loadResumes()
  }
}

// 确认发送简历
const confirmAttachResume = async () => {
  if (!selectedResumeId.value) {
    message.warning('请先选择一份简历')
    return
  }
  if (!isOngoing.value) {
    message.warning('面试已结束，无法发送简历')
    return
  }
  attaching.value = true
  try {
    const attached = await interviewStore.attachResume(interviewId.value, {
      resume_id: selectedResumeId.value,
      version_id: selectedVersionId.value || undefined,
    })
    if (!attached) return
    resumeModalVisible.value = false
    // 切到面试信息 Tab，让用户看到已发送简历的展示
    sideTab.value = 'info'
  } finally {
    attaching.value = false
  }
}

// 发送文字回答或音频草稿
const handleSendAnswer = async () => {
  if (!isOngoing.value) {
    message.warning('面试已结束，无法继续作答')
    return
  }

  if (audioDraft.value) {
    const draft = audioDraft.value
    const sent = await interviewStore.sendVoice(interviewId.value, draft.file)
    if (sent) clearAudioDraft()
    return
  }

  const content = inputText.value.trim()
  if (!content) return
  inputText.value = ''
  // 等待 DOM 更新完成，避免 a-textarea 在 disabled 状态变化时把旧值同步回 inputText
  await nextTick()
  try {
    await interviewStore.sendMessage(interviewId.value, content)
  } finally {
    inputText.value = ''
  }
}

// 中文、日文等输入法确认候选词时也会触发 Enter。
// 组合输入期间必须交给 IME 处理，只有普通 Enter 才发送回答。
const handleAnswerKeydown = (event: KeyboardEvent) => {
  if (event.isComposing || event.keyCode === 229) return
  event.preventDefault()
  void handleSendAnswer()
}

// ============ 麦克风录音回答 ============

const stopRecordingTimer = () => {
  if (recordingTimer !== null) {
    window.clearInterval(recordingTimer)
    recordingTimer = null
  }
}

const releaseMicrophone = () => {
  microphoneStream?.getTracks().forEach((track) => track.stop())
  microphoneStream = null
}

const clearAudioDraft = () => {
  if (audioDraft.value) {
    URL.revokeObjectURL(audioDraft.value.objectUrl)
    audioDraft.value = null
  }
}

const teardownAudioGraph = async () => {
  if (recordingProcessor) {
    recordingProcessor.onaudioprocess = null
    recordingProcessor.disconnect()
    recordingProcessor = null
  }
  microphoneSource?.disconnect()
  microphoneSource = null
  silentGain?.disconnect()
  silentGain = null
  const context = audioContext
  audioContext = null
  if (context && context.state !== 'closed') {
    await context.close()
  }
  releaseMicrophone()
}

const mergePCMChunks = (chunks: Float32Array[]) => {
  const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0)
  const merged = new Float32Array(totalLength)
  let offset = 0
  chunks.forEach((chunk) => {
    merged.set(chunk, offset)
    offset += chunk.length
  })
  return merged
}

const downsamplePCM = (
  samples: Float32Array,
  sourceRate: number,
  targetRate: number
) => {
  if (sourceRate <= targetRate) return samples
  const ratio = sourceRate / targetRate
  const outputLength = Math.round(samples.length / ratio)
  const output = new Float32Array(outputLength)
  for (let i = 0; i < outputLength; i += 1) {
    const start = Math.floor(i * ratio)
    const end = Math.min(Math.floor((i + 1) * ratio), samples.length)
    let sum = 0
    for (let j = start; j < end; j += 1) sum += samples[j]
    output[i] = sum / Math.max(1, end - start)
  }
  return output
}

const encodePCMAsWav = (samples: Float32Array, sampleRate: number) => {
  const buffer = new ArrayBuffer(44 + samples.length * 2)
  const view = new DataView(buffer)
  const writeText = (offset: number, value: string) => {
    for (let i = 0; i < value.length; i += 1) {
      view.setUint8(offset + i, value.charCodeAt(i))
    }
  }

  writeText(0, 'RIFF')
  view.setUint32(4, 36 + samples.length * 2, true)
  writeText(8, 'WAVE')
  writeText(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, 1, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, sampleRate * 2, true)
  view.setUint16(32, 2, true)
  view.setUint16(34, 16, true)
  writeText(36, 'data')
  view.setUint32(40, samples.length * 2, true)

  let offset = 44
  samples.forEach((sample) => {
    const clamped = Math.max(-1, Math.min(1, sample))
    view.setInt16(offset, clamped < 0 ? clamped * 0x8000 : clamped * 0x7fff, true)
    offset += 2
  })
  return new Blob([buffer], { type: 'audio/wav' })
}

const finalizeVoiceRecording = async () => {
  const shouldDiscard = discardRecordedVoice
  const chunks = recordedPCMChunks
  const sourceRate = recordingSampleRate
  recordedPCMChunks = []
  stopRecordingTimer()
  isRecording.value = false
  recordingFinalizing.value = !shouldDiscard
  await teardownAudioGraph()

  if (shouldDiscard) {
    discardRecordedVoice = false
    recordingFinalizing.value = false
    recordingSeconds.value = 0
    return
  }

  const pcm = mergePCMChunks(chunks)
  const targetSampleRate = Math.min(sourceRate, 16000)
  const normalizedPCM = downsamplePCM(pcm, sourceRate, targetSampleRate)
  if (normalizedPCM.length === 0) {
    message.error('没有录到声音，请重新尝试')
    recordingFinalizing.value = false
    recordingSeconds.value = 0
    return
  }
  const audioBlob = encodePCMAsWav(normalizedPCM, targetSampleRate)
  if (audioBlob.size > 7 * 1024 * 1024) {
    message.error('录音过长，请控制在约 3 分钟以内')
    recordingFinalizing.value = false
    recordingSeconds.value = 0
    return
  }

  const file = new File(
    [audioBlob],
    `voice-answer-${Date.now()}.wav`,
    { type: 'audio/wav' }
  )
  clearAudioDraft()
  audioDraft.value = {
    file,
    objectUrl: URL.createObjectURL(file),
    durationSeconds: normalizedPCM.length / targetSampleRate,
  }
  recordingSeconds.value = 0
  recordingFinalizing.value = false
}

const startVoiceRecording = async () => {
  if (!isOngoing.value) {
    message.warning('面试已结束，无法继续作答')
    return
  }
  const windowWithWebkitAudio = window as typeof window & {
    webkitAudioContext?: typeof AudioContext
  }
  const AudioContextClass = window.AudioContext || windowWithWebkitAudio.webkitAudioContext
  if (!navigator.mediaDevices?.getUserMedia || !AudioContextClass) {
    message.error('当前浏览器不支持麦克风录音，请使用最新版 Chrome、Edge 或 Safari')
    return
  }

  voicePreparing.value = true
  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
      },
    })
    microphoneStream = stream
    const context = new AudioContextClass()
    const source = context.createMediaStreamSource(stream)
    const processor = context.createScriptProcessor(4096, 1, 1)
    const gain = context.createGain()
    gain.gain.value = 0

    audioContext = context
    microphoneSource = source
    recordingProcessor = processor
    silentGain = gain
    recordingSampleRate = context.sampleRate
    recordedPCMChunks = []
    discardRecordedVoice = false

    processor.onaudioprocess = (event) => {
      if (!isRecording.value) return
      recordedPCMChunks.push(new Float32Array(event.inputBuffer.getChannelData(0)))
    }
    source.connect(processor)
    processor.connect(gain)
    gain.connect(context.destination)
    await context.resume()

    isRecording.value = true
    recordingSeconds.value = 0
    recordingTimer = window.setInterval(() => {
      recordingSeconds.value += 1
    }, 1000)
  } catch (error) {
    await teardownAudioGraph()
    if (error instanceof DOMException && error.name === 'NotAllowedError') {
      message.error('未获得麦克风权限，请在浏览器地址栏允许访问麦克风')
    } else if (error instanceof DOMException && error.name === 'NotFoundError') {
      message.error('未检测到可用的麦克风')
    } else {
      message.error('启动录音失败，请检查麦克风后重试')
    }
  } finally {
    voicePreparing.value = false
  }
}

const stopVoiceRecording = () => {
  if (!isRecording.value) return
  void finalizeVoiceRecording()
}

const toggleVoiceRecording = async () => {
  if (isRecording.value) {
    stopVoiceRecording()
    return
  }
  await startVoiceRecording()
}

const discardVoiceRecording = () => {
  discardRecordedVoice = true
  stopRecordingTimer()
  if (isRecording.value) {
    void finalizeVoiceRecording()
  } else {
    void teardownAudioGraph()
  }
  isRecording.value = false
}

const formatDuration = (seconds: number) => {
  const rounded = Math.max(0, Math.round(seconds))
  const minutes = Math.floor(rounded / 60)
  return `${String(minutes).padStart(2, '0')}:${String(rounded % 60).padStart(2, '0')}`
}

const formatFileSize = (size: number) => {
  if (size < 1024 * 1024) return `${Math.max(1, Math.round(size / 1024))} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

const handleTranscribeDraft = async () => {
  const draft = audioDraft.value
  if (!draft || transcribingDraft.value) return
  transcribingDraft.value = true
  try {
    const response = await transcribeVoice(interviewId.value, draft.file)
    const text = response.data.data?.text?.trim()
    if (!text) {
      message.error('未识别到有效文字，请重新录制')
      return
    }
    clearAudioDraft()
    inputText.value = text
    message.success('已转化为文本，可编辑后发送')
  } catch (error) {
    console.error('语音转文字失败:', error)
  } finally {
    transcribingDraft.value = false
  }
}

const rerecordVoice = async () => {
  clearAudioDraft()
  await startVoiceRecording()
}

// TTS 播放（调后端 xiaomi mimo TTS）
const stopTtsPlayback = () => {
  const player = audioPlayer.value
  if (player) {
    player.pause()
    player.onended = null
    player.onerror = null
    player.removeAttribute('src')
    player.load()
  }
  if (ttsObjectUrl) {
    URL.revokeObjectURL(ttsObjectUrl)
    ttsObjectUrl = null
  }
  playingMessageId.value = null
}

const handlePlayTts = async (msg: InterviewMessage) => {
  // 同一题 → 停止
  if (playingMessageId.value === msg.id) {
    stopTtsPlayback()
    return
  }
  // 正在合成别的 → 拒绝并发
  if (ttsLoadingId.value !== null) {
    message.warning('正在合成上一题语音，请稍候')
    return
  }
  const player = audioPlayer.value
  if (!player) return

  stopTtsPlayback()
  ttsLoadingId.value = msg.id
  try {
    const url = await interviewStore.playTts(interviewId.value, msg.id)
    if (!url) return
    // 合成期间用户切到别的题 → 释放这次 URL
    if (ttsLoadingId.value !== msg.id) {
      URL.revokeObjectURL(url)
      return
    }
    ttsObjectUrl = url
    player.src = url
    player.onended = () => {
      stopTtsPlayback()
    }
    player.onerror = () => {
      message.error('音频播放失败')
      stopTtsPlayback()
    }
    await player.play()
    playingMessageId.value = msg.id
  } catch (error) {
    console.error('TTS 合成失败:', error)
    stopTtsPlayback()
    message.error('音频合成失败，请稍后重试')
  } finally {
    if (ttsLoadingId.value === msg.id) ttsLoadingId.value = null
  }
}

onBeforeUnmount(() => {
  discardVoiceRecording()
  clearAudioDraft()
  stopTtsPlayback()
})

// 结束面试
const handleEndInterview = () => {
  Modal.confirm({
    title: '确认结束面试',
    content: '结束后将生成复盘报告，无法继续作答。是否继续？',
    okText: '结束面试',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      const report = await interviewStore.endInterview(interviewId.value)
      if (report) {
        sideTab.value = 'report'
        // 同时加载评分明细
        scoresLoading.value = true
        await interviewStore.fetchScores(interviewId.value)
        scoresLoading.value = false
      }
    },
  })
}

// 初始化加载
onMounted(async () => {
  if (!interviewId.value || isNaN(interviewId.value)) {
    message.error('无效的面试 ID')
    router.push('/app/interviews')
    return
  }
  await interviewStore.fetchOne(interviewId.value)
  // 已完成的面试：自动加载报告 + 评分
  if (interviewStore.currentInterview?.status === 'completed') {
    sideTab.value = 'report'
    reportLoading.value = true
    await interviewStore.fetchReport(interviewId.value)
    reportLoading.value = false
    scoresLoading.value = true
    await interviewStore.fetchScores(interviewId.value)
    scoresLoading.value = false
  }
  scrollToBottom()
})
</script>

<style scoped src="./styles/interview-room.css"></style>
