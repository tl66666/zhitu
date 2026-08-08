<template src="./templates/DeliveryKanban.html"></template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { listResumes } from '@/api/resume'
import { useDeliveryStore } from '@/stores/delivery'
import { message, Modal } from 'ant-design-vue'
import type { FormInstance } from 'ant-design-vue'
import type {
  Delivery,
  DeliveryStatus,
  DeliveryRound,
  CreateDeliveryRequest,
  CreateRoundRequest,
  CreateFeedbackRequest,
  Resume,
} from '@/types/models'
import dayjs, { Dayjs } from 'dayjs'
// 使用 @ant-design/icons-vue，不再使用 lucide-vue-next
import {
  SendOutlined,
  TeamOutlined,
  TrophyOutlined,
  CloseCircleOutlined,
  TableOutlined,
  PieChartOutlined,
  SearchOutlined,
  PlusOutlined,
  RightOutlined,
  RightCircleOutlined,
  EllipsisOutlined,
  CalendarOutlined,
  FileTextOutlined,
  UserOutlined,
  MessageOutlined,
  ArrowRightOutlined,
  CloseOutlined,
  ExportOutlined,
  RiseOutlined,
  ThunderboltOutlined,
  BankOutlined,
  SolutionOutlined,
  DeleteOutlined,
  InboxOutlined,
  AimOutlined,
  StarOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const deliveryStore = useDeliveryStore()

// ==================== 视图与筛选 ====================
const viewMode = ref<'personal' | 'platform'>('personal')
const searchKeyword = ref('')
const hoveredId = ref<number | null>(null)

const filters = reactive({
  status: '' as string,
  channel: '' as string,
  priority: '' as string,
})

// 状态枚举（6 列，对齐后端 + 原型配色）
const statusList = [
  { label: '待响应', value: 'pending' as DeliveryStatus, dotColor: '#aeaeb2' },
  { label: '笔试中', value: 'written_test' as DeliveryStatus, dotColor: '#ff9500' },
  { label: '面试中', value: 'interview' as DeliveryStatus, dotColor: '#007aff' },
  { label: '待Offer', value: 'waiting_offer' as DeliveryStatus, dotColor: '#af52de' },
  { label: '已Offer', value: 'offer' as DeliveryStatus, dotColor: '#34c759' },
  { label: '已拒绝', value: 'rejected' as DeliveryStatus, dotColor: '#ff3b30' },
]

const channelOptions = [
  { value: 'boss', label: 'BOSS直聘' },
  { value: 'official', label: '官网' },
  { value: 'referral', label: '内推' },
  { value: 'campus', label: '校园招聘' },
  { value: 'headhunt', label: '猎头' },
  { value: 'other', label: '其他' },
]

// 状态合法流转
const transitionMap: Record<DeliveryStatus, DeliveryStatus[]> = {
  pending: ['written_test', 'interview', 'rejected'],
  written_test: ['interview', 'waiting_offer', 'rejected'],
  interview: ['waiting_offer', 'offer', 'rejected'],
  waiting_offer: ['offer', 'rejected'],
  offer: [],
  rejected: ['interview'],
}

const getAvailableTransitions = (status: DeliveryStatus) => {
  return transitionMap[status]?.map((v) => ({ value: v, label: getStatusText(v) })) || []
}

// 本地筛选 + 搜索
const filteredDeliveries = computed(() => {
  return deliveryStore.deliveries.filter((d) => {
    if (filters.status && d.status !== filters.status) return false
    if (filters.channel && d.channel !== filters.channel) return false
    if (filters.priority && d.priority !== filters.priority) return false
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      if (!d.company.toLowerCase().includes(kw) && !d.position.toLowerCase().includes(kw)) return false
    }
    return true
  })
})

const hasActiveFilter = computed(() =>
  !!(filters.status || filters.channel || filters.priority || searchKeyword.value)
)

const clearFilters = () => {
  filters.status = ''
  filters.channel = ''
  filters.priority = ''
  searchKeyword.value = ''
}

// ==================== 统计数据 + Sparkline ====================
const stats = computed(() => {
  const total = deliveryStore.deliveries.length
  const inProgress = deliveryStore.deliveries.filter(
    (d) => d.status === 'interview' || d.status === 'written_test'
  ).length
  const offerCount = deliveryStore.deliveries.filter((d) => d.status === 'offer').length
  const rejected = deliveryStore.deliveries.filter((d) => d.status === 'rejected').length
  return { total, inProgress, offerCount, rejected }
})

const rejectRate = computed(() => {
  if (stats.value.total === 0) return 0
  return Math.round((stats.value.rejected / stats.value.total) * 100)
})

const weekCount = computed(() => {
  const weekAgo = dayjs().subtract(7, 'day')
  return deliveryStore.deliveries.filter((d) => dayjs(d.apply_date).isAfter(weekAgo)).length
})

const offerWeekCount = computed(() => {
  const weekAgo = dayjs().subtract(7, 'day')
  return deliveryStore.deliveries.filter(
    (d) => d.status === 'offer' && dayjs(d.updated_at).isAfter(weekAgo)
  ).length
})

// 生成 7 天 sparkline 数据（基于 apply_date）
const generateSparkline = (predicate: (d: Delivery) => boolean): string => {
  const points: number[] = []
  for (let i = 6; i >= 0; i--) {
    const day = dayjs().subtract(i, 'day')
    const next = day.add(1, 'day')
    const count = deliveryStore.deliveries.filter(
      (d) => predicate(d) && dayjs(d.apply_date).isAfter(day.subtract(1, 'ms')) && dayjs(d.apply_date).isBefore(next)
    ).length
    points.push(count)
  }
  const max = Math.max(...points, 1)
  return points.map((p, i) => `${(i / 6) * 80},${24 - (p / max) * 22 - 1}`).join(' ')
}

// 统计卡片：1:1 对齐设计稿配色（蓝/紫/绿/红）
const statCards = computed(() => [
  {
    key: 'total',
    label: '投递总数',
    value: stats.value.total,
    change: `↑ +${weekCount.value} 本周`,
    changeClass: 'up',
    icon: SendOutlined,
    iconBg: 'var(--brand-50)',
    iconColor: 'var(--primary)',
    accentClass: 'accent-blue',
    sparkline: generateSparkline(() => true),
    sparkColor: '#007aff',
  },
  {
    key: 'progress',
    label: '面试中',
    value: stats.value.inProgress,
    change: `↑ 进行中`,
    changeClass: 'up',
    icon: TeamOutlined,
    iconBg: 'rgba(175,82,222,0.14)',
    iconColor: '#8e3db5',
    accentClass: 'accent-purple',
    sparkline: generateSparkline((d) => d.status === 'interview' || d.status === 'written_test'),
    sparkColor: '#af52de',
  },
  {
    key: 'offer',
    label: 'Offer',
    value: stats.value.offerCount,
    change: `↑ +${offerWeekCount.value} 本月`,
    changeClass: 'up',
    icon: TrophyOutlined,
    iconBg: 'var(--state-success-surface)',
    iconColor: 'var(--success)',
    accentClass: 'accent-green',
    sparkline: generateSparkline((d) => d.status === 'offer'),
    sparkColor: '#34c759',
  },
  {
    key: 'rejected',
    label: '拒绝',
    value: stats.value.rejected,
    change: `↓ 转化率 ${rejectRate.value}%`,
    changeClass: 'down',
    icon: CloseCircleOutlined,
    iconBg: 'var(--state-error-surface)',
    iconColor: 'var(--state-error)',
    accentClass: 'accent-red',
    sparkline: generateSparkline((d) => d.status === 'rejected'),
    sparkColor: '#ff3b30',
  },
])

// ==================== 选中与详情 ====================
const selectedId = ref<number | null>(null)

const selectedDelivery = computed(() => {
  if (!selectedId.value) return null
  return deliveryStore.deliveries.find((d) => d.id === selectedId.value) || deliveryStore.currentDelivery
})

const rounds = computed(() => deliveryStore.rounds)
const feedbacks = computed(() => deliveryStore.feedbacks)

const sortedRounds = computed(() => {
  return [...deliveryStore.rounds]
    .filter((r) => r.result !== 'pending' || r.interview_time)
    .sort((a, b) => {
      const ta = a.interview_time ? new Date(a.interview_time).getTime() : 0
      const tb = b.interview_time ? new Date(b.interview_time).getTime() : 0
      return ta - tb
    })
})

const selectDelivery = async (delivery: Delivery) => {
  selectedId.value = delivery.id
  await deliveryStore.fetchDelivery(delivery.id)
}

// ==================== 进度点（5 个：笔试/一面/二面/HR面/终面）====================
const progressSlots = ['written_test', 'first_tech', 'second_tech', 'hr', 'final']

const getProgressDots = (delivery: Delivery): string[] => {
  const dots: string[] = []
  for (const slot of progressSlots) {
    const round = deliveryStore.rounds.find((r) => r.round_type === slot)
    if (!round) {
      dots.push('')
    } else if (round.result === 'pass') {
      dots.push('done')
    } else if (round.result === 'rejected') {
      dots.push('fail')
    } else {
      dots.push('current')
    }
  }
  // 无轮次数据时，根据状态推断
  if (deliveryStore.rounds.length === 0) {
    const statusProgress: Record<string, number> = {
      pending: 0,
      written_test: 1,
      interview: 2,
      waiting_offer: 4,
      offer: 5,
      rejected: 0,
    }
    const currentStep = statusProgress[delivery.status] ?? 0
    return progressSlots.map((_, i) => {
      if (i < currentStep) return 'done'
      if (i === currentStep && delivery.status !== 'offer' && delivery.status !== 'pending') return 'current'
      if (delivery.status === 'offer') return 'done'
      return ''
    })
  }
  return dots
}

const getProgressTooltip = (delivery: Delivery): string => {
  const labels = ['笔试', '一面', '二面', 'HR面', '终面']
  const dots = getProgressDots(delivery)
  return labels.map((l, i) => {
    const d = dots[i]
    if (d === 'done') return `${l}✓`
    if (d === 'current') return `${l}○`
    if (d === 'fail') return `${l}✗`
    return `${l}—`
  }).join(' ')
}

// ==================== 下次面试 / HR 最新反馈 ====================
const getNextInterview = (delivery: Delivery): DeliveryRound | null => {
  if (delivery.id !== selectedId.value && deliveryStore.rounds.length === 0) return null
  if (delivery.id !== selectedId.value) return null
  const now = Date.now()
  return sortedRounds.value.find(
    (r) => r.result === 'pending' && r.interview_time && new Date(r.interview_time).getTime() > now
  ) || null
}

const getLatestFeedback = (deliveryId: number) => {
  if (deliveryId !== selectedId.value) return null
  return feedbacks.value[0] || null
}

// ==================== 漏斗与渠道统计 ====================
const funnelData = computed(() => {
  const total = stats.value.total || 1
  const written = deliveryStore.deliveries.filter(
    (d) => ['written_test', 'interview', 'waiting_offer', 'offer'].includes(d.status)
  ).length
  const interview = deliveryStore.deliveries.filter(
    (d) => ['interview', 'waiting_offer', 'offer'].includes(d.status)
  ).length
  const waiting = deliveryStore.deliveries.filter(
    (d) => ['waiting_offer', 'offer'].includes(d.status)
  ).length
  const offer = stats.value.offerCount
  return [
    { label: '投递', value: stats.value.total, pct: 100 },
    { label: '笔试', value: written, pct: Math.round((written / total) * 100) },
    { label: '面试', value: interview, pct: Math.round((interview / total) * 100) },
    { label: '待Offer', value: waiting, pct: Math.round((waiting / total) * 100) },
    { label: 'Offer', value: offer, pct: Math.round((offer / total) * 100) },
  ]
})

// SVG 漏斗图分段
const funnelSegments = computed(() => {
  const data = funnelData.value
  const maxW = 280
  const segH = 18
  const gap = 2
  const colors = ['#007aff', '#5e5ce6', '#af52de', '#ff9500', '#34c759']
  const maxVal = data[0].value || 1
  return data.map((seg, i) => {
    const w = Math.max(20, (seg.value / maxVal) * maxW)
    const x1 = (maxW - w) / 2
    const x2 = x1 + w
    const y1 = i * (segH + gap)
    const y2 = y1 + segH
    const prevW = i === 0 ? maxW : Math.max(20, (data[i - 1].value / maxVal) * maxW)
    const prevX1 = (maxW - prevW) / 2
    const prevX2 = prevX1 + prevW
    return {
      label: seg.label,
      value: seg.value,
      pct: seg.pct,
      color: colors[i],
      points: `${prevX1},${y1} ${prevX2},${y1} ${x2},${y2} ${x1},${y2}`,
    }
  })
})

const channelStats = computed(() => {
  const counts: Record<string, number> = {}
  deliveryStore.deliveries.forEach((d) => {
    counts[d.channel] = (counts[d.channel] || 0) + 1
  })
  const max = Math.max(...Object.values(counts), 1)
  const clsMap: Record<string, string> = { boss: 'boss', official: 'web', referral: 'ref', campus: 'you', headhunt: 'avg', other: 'avg' }
  return Object.entries(counts).map(([ch, val]) => ({
    label: getChannelLabel(ch),
    value: val,
    pct: Math.round((val / max) * 100),
    cls: clsMap[ch] || 'avg',
  }))
})

const avgCycle = computed(() => {
  const offered = deliveryStore.deliveries.filter((d) => d.status === 'offer')
  if (offered.length === 0) return 0
  const days = offered.map((d) => {
    const start = dayjs(d.apply_date)
    const end = dayjs(d.updated_at)
    return end.diff(start, 'day')
  })
  return Math.round(days.reduce((a, b) => a + b, 0) / days.length)
})

// ==================== Quick Capture（顶部快捷输入）====================
const quickCapture = ref('')
const quickCaptureParsed = ref<{ company?: string; position?: string }>({})

const handlePaste = (e: ClipboardEvent) => {
  const text = e.clipboardData?.getData('text') || ''
  // 简单 JD 解析：尝试匹配「公司：xxx」「岗位：xxx」
  const companyMatch = text.match(/公司[：:]\s*(.+)/)
  const positionMatch = text.match(/岗位[：:]\s*(.+)/) || text.match(/职位[：:]\s*(.+)/)
  quickCaptureParsed.value = {
    company: companyMatch?.[1]?.trim(),
    position: positionMatch?.[1]?.trim(),
  }
}

const handleQuickCapture = async () => {
  const text = quickCapture.value.trim()
  if (!text) return
  const parsed = quickCaptureParsed.value
  if (parsed.company && parsed.position) {
    createFromQuickCapture(parsed.company, parsed.position, text)
    return
  }
  if (text.includes('/')) {
    const [company, position] = text.split('/').map((s) => s.trim())
    if (company && position) {
      createFromQuickCapture(company, position, '')
      return
    }
  }
  createForm.company = text
  createForm.jd_text = text.length > 20 ? text : ''
  await showCreateModal()
  quickCapture.value = ''
  quickCaptureParsed.value = {}
}

const createFromQuickCapture = async (company: string, position: string, jdText: string) => {
  const data: CreateDeliveryRequest = {
    company,
    position,
    channel: 'other',
    apply_date: dayjs().format('YYYY-MM-DD'),
    priority: 'medium',
    jd_text: jdText,
  }
  await deliveryStore.createDelivery(data)
  quickCapture.value = ''
  quickCaptureParsed.value = {}
}

// ==================== 键盘快捷键 ====================
const handleKeydown = (e: KeyboardEvent) => {
  const target = e.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return
  if (target.isContentEditable) return

  const list = filteredDeliveries.value
  const currentIdx = list.findIndex((d) => d.id === selectedId.value)

  if (e.key === 'j' || e.key === 'J') {
    e.preventDefault()
    const next = list[Math.min(currentIdx + 1, list.length - 1)] || list[0]
    if (next) selectDelivery(next)
  } else if (e.key === 'k' || e.key === 'K') {
    e.preventDefault()
    const prev = list[Math.max(currentIdx - 1, 0)]
    if (prev) selectDelivery(prev)
  } else if (e.key === 'n' || e.key === 'N') {
    e.preventDefault()
    showCreateModal()
  } else if (e.key === 'Escape') {
    selectedId.value = null
  }
}

// ==================== 新增投递 ====================
const createModalVisible = ref(false)
const createFormRef = ref<FormInstance>()
const resumes = ref<Resume[]>([])
const resumeOptions = computed(() => resumes.value.filter((resume) => resume.current_version_id > 0))
const createForm = reactive<CreateDeliveryRequest & { apply_date?: Dayjs }>({
  company: '',
  position: '',
  channel: 'boss',
  apply_date: undefined,
  jd_text: '',
  remark: '',
  resume_version_id: undefined,
})

const createRules = {
  company: [{ required: true, message: '请输入公司名称', trigger: 'blur' }],
  position: [{ required: true, message: '请输入职位名称', trigger: 'blur' }],
  channel: [{ required: true, message: '请选择投递渠道', trigger: 'change' }],
  apply_date: [{ required: true, message: '请选择投递日期', trigger: 'change' }],
  resume_version_id: [{ required: true, message: '请选择简历', trigger: 'change' }],
}

const loadResumes = async () => {
  try {
    const response = await listResumes()
    resumes.value = response.data.data || []
  } catch (error) {
    console.error('加载简历列表失败:', error)
    message.error('加载简历列表失败')
  }
}

const showCreateModal = async () => {
  await loadResumes()
  createModalVisible.value = true
}

const handleCreate = async () => {
  try {
    await createFormRef.value?.validateFields()
    const data: CreateDeliveryRequest = {
      company: createForm.company,
      position: createForm.position,
      channel: createForm.channel as CreateDeliveryRequest['channel'],
      apply_date: createForm.apply_date?.format('YYYY-MM-DD') || '',
      jd_text: createForm.jd_text,
      remark: createForm.remark,
      resume_version_id: createForm.resume_version_id,
    }
    await deliveryStore.createDelivery(data)
    createModalVisible.value = false
    resetCreateForm()
  } catch (error) {
    console.error('验证失败:', error)
  }
}

const resetCreateForm = () => {
  createFormRef.value?.resetFields()
  Object.assign(createForm, {
    company: '',
    position: '',
    channel: 'boss',
    apply_date: undefined,
    jd_text: '',
    remark: '',
    resume_version_id: undefined,
  })
}

// ==================== 状态变更与删除 ====================
const handleStatusChange = async (delivery: Delivery, status: DeliveryStatus) => {
  await deliveryStore.changeStatus(delivery.id, { status })
}

const handleDelete = (id: number) => {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这条投递记录吗？',
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      await deliveryStore.deleteDelivery(id)
      if (selectedId.value === id) selectedId.value = null
    },
  })
}

// ==================== 轮次操作 ====================
const roundModalVisible = ref(false)
const editingRound = ref<DeliveryRound | null>(null)
const roundForm = reactive<{
  round_type: string
  interview_time: Dayjs | null
  format: string
  interviewer_name: string
  interviewer_title: string
  question_summary: string
  feedback: string
  result: string
}>({
  round_type: 'first_tech',
  interview_time: null,
  format: 'video',
  interviewer_name: '',
  interviewer_title: '',
  question_summary: '',
  feedback: '',
  result: 'pending',
})

const showRoundModal = (round?: DeliveryRound) => {
  if (round) {
    editingRound.value = round
    Object.assign(roundForm, {
      round_type: round.round_type,
      interview_time: round.interview_time ? dayjs(round.interview_time) : null,
      format: round.format,
      interviewer_name: round.interviewer_name,
      interviewer_title: round.interviewer_title,
      question_summary: round.question_summary,
      feedback: round.feedback,
      result: round.result,
    })
  } else {
    editingRound.value = null
    Object.assign(roundForm, {
      round_type: 'first_tech',
      interview_time: null,
      format: 'video',
      interviewer_name: '',
      interviewer_title: '',
      question_summary: '',
      feedback: '',
      result: 'pending',
    })
  }
  roundModalVisible.value = true
}

const showRoundModalFor = (delivery: Delivery) => {
  selectDelivery(delivery).then(() => showRoundModal())
}

const handleSaveRound = async () => {
  if (!selectedDelivery.value) return
  const deliveryId = selectedDelivery.value.id
  const data: CreateRoundRequest = {
    round_type: roundForm.round_type,
    interview_time: roundForm.interview_time?.format('YYYY-MM-DD HH:mm:ss') || '',
    format: roundForm.format,
    interviewer_name: roundForm.interviewer_name,
    interviewer_title: roundForm.interviewer_title,
    question_summary: roundForm.question_summary,
    feedback: roundForm.feedback,
    result: roundForm.result,
  }
  if (editingRound.value) {
    await deliveryStore.updateRound(deliveryId, editingRound.value.id, data)
  } else {
    await deliveryStore.createRound(deliveryId, data)
  }
  roundModalVisible.value = false
}

const handleDeleteRound = (roundId: number) => {
  if (!selectedDelivery.value) return
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该面试轮次吗？',
    onOk: async () => {
      await deliveryStore.deleteRound(selectedDelivery.value!.id, roundId)
    },
  })
}

// ==================== 反馈操作 ====================
const feedbackModalVisible = ref(false)
const feedbackForm = reactive<{
  contact_time: Dayjs | null
  method: string
  summary: string
  next_action: string
}>({
  contact_time: null,
  method: 'wechat',
  summary: '',
  next_action: '',
})

const showFeedbackModal = () => {
  Object.assign(feedbackForm, {
    contact_time: dayjs(),
    method: 'wechat',
    summary: '',
    next_action: '',
  })
  feedbackModalVisible.value = true
}

const showFeedbackModalFor = (delivery: Delivery) => {
  selectDelivery(delivery).then(() => showFeedbackModal())
}

const handleSaveFeedback = async () => {
  if (!selectedDelivery.value) return
  if (!feedbackForm.contact_time || !feedbackForm.summary) {
    message.warning('请填写联系时间和反馈摘要')
    return
  }
  const data: CreateFeedbackRequest = {
    contact_time: feedbackForm.contact_time.format('YYYY-MM-DD HH:mm'),
    method: feedbackForm.method,
    summary: feedbackForm.summary,
    next_action: feedbackForm.next_action,
  }
  await deliveryStore.createFeedback(selectedDelivery.value.id, data)
  feedbackModalVisible.value = false
}

const handleDeleteFeedback = (feedbackId: number) => {
  if (!selectedDelivery.value) return
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除该反馈吗？',
    onOk: async () => {
      await deliveryStore.deleteFeedback(selectedDelivery.value!.id, feedbackId)
    },
  })
}

// ==================== 跳转面试训练场 ====================
const goToInterviewTraining = (delivery: Delivery) => {
  router.push({
    path: '/interview',
    query: {
      company: delivery.company,
      position: delivery.position,
      jd: delivery.jd_text || '',
    },
  })
}

// ==================== 辅助函数 ====================
const handleFilter = async () => {
  await deliveryStore.fetchDeliveries({
    status: filters.status || undefined,
    channel: filters.channel || undefined,
  })
}

const getStatusText = (status: string): string => {
  return statusList.find((s) => s.value === status)?.label || status
}

const getStatusClass = (status: string): string => {
  const map: Record<string, string> = {
    pending: 'gray',
    written_test: 'orange',
    interview: 'blue',
    waiting_offer: 'purple',
    offer: 'green',
    rejected: 'red',
  }
  return map[status] || 'gray'
}

const getStatusColor = (status: string): string => {
  const map: Record<string, string> = {
    pending: '#aeaeb2',
    written_test: '#ff9500',
    interview: '#007aff',
    waiting_offer: '#af52de',
    offer: '#34c759',
    rejected: '#ff3b30',
  }
  return map[status] || '#aeaeb2'
}

const getChannelLabel = (channel: string): string => {
  return channelOptions.find((c) => c.value === channel)?.label || channel || '-'
}

const getChannelClass = (channel: string): string => {
  const map: Record<string, string> = {
    boss: 'ch-boss',
    official: 'ch-official',
    referral: 'ch-referral',
    campus: 'ch-campus',
    headhunt: 'ch-headhunt',
    other: 'ch-other',
  }
  return map[channel] || 'ch-other'
}

const getPriorityText = (priority: string): string => {
  const map: Record<string, string> = { high: '高', medium: '中', low: '低' }
  return map[priority] || priority || '-'
}

const getPriorityClass = (priority: string): string => {
  const map: Record<string, string> = { high: 'high', medium: 'mid', low: 'low' }
  return map[priority] || 'low'
}

const getRoundResultText = (result?: string): string => {
  if (!result) return '待定'
  const map: Record<string, string> = { pass: '通过', pending: '待定', rejected: '未通过' }
  return map[result] || result
}

const getRoundResultClass = (result?: string): string => {
  if (result === 'pass') return 'result-pass'
  if (result === 'rejected') return 'result-fail'
  return 'result-pending'
}

const getRoundTypeText = (type?: string): string => {
  const map: Record<string, string> = {
    written_test: '笔试',
    first_tech: '技术一面',
    second_tech: '技术二面',
    third_tech: '技术三面',
    cross: '交叉面',
    hr: 'HR 面',
    additional: '加面',
    final: '终面',
  }
  return map[type] || type || '未知'
}

const getRoundFormatText = (format?: string): string => {
  const map: Record<string, string> = { onsite: '现场', video: '视频', phone: '电话' }
  return map[format] || format || '-'
}

const getMethodText = (method?: string): string => {
  const map: Record<string, string> = { wechat: '微信', phone: '电话', email: '邮件' }
  return map[method] || method || '-'
}

const getTimelineLineClass = (round: DeliveryRound, idx: number): string => {
  if (idx === sortedRounds.value.length - 1) return 'line-none'
  return round.result === 'pass' ? 'line-solid' : 'line-dashed'
}

const getTimelineDotClass = (round: DeliveryRound): string => {
  if (round.result === 'pass') return 'green'
  if (round.result === 'rejected') return 'red'
  return 'blue'
}

const parseHrContact = (json: string): { name?: string; wechat?: string; phone?: string; email?: string } => {
  if (!json) return {}
  try {
    return JSON.parse(json)
  } catch {
    return {}
  }
}

const parseOfferDetail = (json: string): {
  salary_base?: string
  annual_bonus?: string
  stock?: string
  benefits?: string
  deadline?: string
} => {
  if (!json) return {}
  try {
    return JSON.parse(json)
  } catch {
    return {}
  }
}

const truncate = (text: string, len: number): string => {
  if (!text) return ''
  return text.length > len ? text.slice(0, len) + '…' : text
}

const formatDate = (date: string): string => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD')
}

const formatShortDateTime = (date: string): string => {
  if (!date) return '-'
  return dayjs(date).format('MM-DD HH:mm')
}

const formatDateTime = (date: string): string => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

// ==================== 平台数据视图（按日期种子稳定） ====================
const seedRand = (seed: number) => {
  let s = seed
  return () => {
    s = (s * 9301 + 49297) % 233280
    return s / 233280
  }
}

const todaySeed = computed(() => {
  const d = new Date()
  return d.getFullYear() * 10000 + (d.getMonth() + 1) * 100 + d.getDate()
})

// 平台核心指标（基于种子稳定生成，避免每次刷新跳变）
const platformStats = computed(() => {
  const rand = seedRand(todaySeed.value)
  const totalUsers = 12800 + Math.floor(rand() * 3200)
  const totalDeliveries = 86000 + Math.floor(rand() * 8000)
  const successRate = 38 + Math.floor(rand() * 12)
  const avgCycleDays = 28 + Math.floor(rand() * 12)
  return { totalUsers, totalDeliveries, successRate, avgCycleDays }
})

// 投递结果分布饼图分段
const pieSegments = computed(() => {
  const rand = seedRand(todaySeed.value + 1)
  const offer = 4200 + Math.floor(rand() * 600)
  const interview = 8600 + Math.floor(rand() * 1200)
  const rejected = 12000 + Math.floor(rand() * 2000)
  const pending = platformStats.value.totalDeliveries - offer - interview - rejected
  const total = platformStats.value.totalDeliveries
  const segs = [
    { label: '已 Offer', value: offer, color: '#34c759' },
    { label: '面试中', value: interview, color: '#007aff' },
    { label: '已拒绝', value: rejected, color: '#ff3b30' },
    { label: '待响应', value: pending, color: '#aeaeb2' },
  ]
  const circumference = 2 * Math.PI * 70
  let accOffset = 0
  return segs.map((seg) => {
    const pct = total > 0 ? Math.round((seg.value / total) * 100) : 0
    const dashLen = (seg.value / total) * circumference
    const segData = {
      ...seg,
      pct,
      fill: 'none',
      stroke: seg.color,
      strokeWidth: 28,
      dashArray: `${dashLen} ${circumference - dashLen}`,
      dashOffset: -accOffset,
    }
    accOffset += dashLen
    return segData
  })
})

// 服务行业分布饼图分段
const industryPieSegments = computed(() => {
  const rand = seedRand(todaySeed.value + 2)
  const total = platformStats.value.totalUsers
  const data = [
    { label: '互联网', value: 4200 + Math.floor(rand() * 800), color: '#007aff' },
    { label: '金融', value: 2800 + Math.floor(rand() * 600), color: '#34c759' },
    { label: '制造', value: 2200 + Math.floor(rand() * 500), color: '#ff9500' },
    { label: '教育', value: 1800 + Math.floor(rand() * 400), color: '#5856d6' },
    { label: '医疗', value: 1400 + Math.floor(rand() * 300), color: '#af52de' },
    { label: '其他', value: total - 12400, color: '#aeaeb2' },
  ]
  const circumference = 2 * Math.PI * 70
  let accOffset = 0
  return data.map((seg) => {
    const pct = total > 0 ? Math.round((seg.value / total) * 100) : 0
    const dashLen = (seg.value / total) * circumference
    const segData = {
      ...seg,
      pct,
      fill: 'none',
      stroke: seg.color,
      strokeWidth: 28,
      dashArray: `${dashLen} ${circumference - dashLen}`,
      dashOffset: -accOffset,
    }
    accOffset += dashLen
    return segData
  })
})

// 求职转化漏斗（平台）
const platformFunnel = computed(() => {
  const rand = seedRand(todaySeed.value + 3)
  const total = platformStats.value.totalDeliveries
  const written = Math.floor(total * (0.62 + rand() * 0.08))
  const interview = Math.floor(written * (0.55 + rand() * 0.08))
  const waiting = Math.floor(interview * (0.42 + rand() * 0.08))
  const offer = Math.floor(waiting * (0.68 + rand() * 0.1))
  const stages = [
    { label: '投递', value: total, pct: 100, color: '#007aff', convPct: 0 },
    { label: '笔试', value: written, pct: Math.round((written / total) * 100), color: '#5e5ce6', convPct: Math.round((written / total) * 100) },
    { label: '面试', value: interview, pct: Math.round((interview / total) * 100), color: '#af52de', convPct: Math.round((interview / written) * 100) },
    { label: '待Offer', value: waiting, pct: Math.round((waiting / total) * 100), color: '#ff9500', convPct: Math.round((waiting / interview) * 100) },
    { label: 'Offer', value: offer, pct: Math.round((offer / total) * 100), color: '#34c759', convPct: Math.round((offer / waiting) * 100) },
  ]
  return stages
})

// 热门求职方向
const hotJobs = computed(() => {
  const rand = seedRand(todaySeed.value + 4)
  const data = [
    { name: '前端工程师', count: 8200 + Math.floor(rand() * 800) },
    { name: '产品经理', count: 6800 + Math.floor(rand() * 600) },
    { name: '后端工程师', count: 7400 + Math.floor(rand() * 700) },
    { name: '算法工程师', count: 5200 + Math.floor(rand() * 500) },
    { name: '数据分析师', count: 4600 + Math.floor(rand() * 400) },
  ]
  const max = Math.max(...data.map((d) => d.count))
  return data.map((d) => ({ ...d, pct: Math.round((d.count / max) * 100) }))
})

// 渠道效果对比
const channelComparison = computed(() => {
  const rand = seedRand(todaySeed.value + 5)
  const data = [
    { name: 'BOSS直聘', count: 4200 + Math.floor(rand() * 400), pct: 0, color: '#007aff' },
    { name: '内推', count: 3600 + Math.floor(rand() * 400), pct: 0, color: '#af52de' },
    { name: '官网', count: 2800 + Math.floor(rand() * 300), pct: 0, color: '#5856d6' },
    { name: '校招', count: 2200 + Math.floor(rand() * 200), pct: 0, color: '#34c759' },
    { name: '猎头', count: 1400 + Math.floor(rand() * 200), pct: 0, color: '#ff9500' },
  ]
  const max = Math.max(...data.map((d) => d.count))
  return data.map((d) => ({ ...d, pct: Math.round((d.count / max) * 100) }))
})

// 平台数据最后更新时间
const platformUpdated = computed(() => {
  return dayjs().format('YYYY-MM-DD HH:mm')
})

// ==================== 生命周期 ====================
onMounted(async () => {
  await deliveryStore.fetchDeliveries()
})

onUnmounted(() => {
  deliveryStore.clearCurrentDelivery()
})
</script>

<style scoped src="./styles/delivery-kanban-shell.css"></style>
<style scoped src="./styles/delivery-kanban-list.css"></style>
<style scoped src="./styles/delivery-kanban-platform.css"></style>
<style scoped src="./styles/delivery-kanban-detail.css"></style>
<style scoped src="./styles/delivery-kanban-responsive.css"></style>
<style scoped>
.form-help {
  margin-top: 6px;
  color: var(--muted-foreground);
  font-size: 12px;
  line-height: 1.4;
}
</style>
