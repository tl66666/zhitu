<template src="./templates/ResumeEditor.html"></template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  AimOutlined, ArrowLeftOutlined, BankOutlined, CheckCircleOutlined, CheckOutlined, CloseOutlined,
  ColumnHeightOutlined, DeleteOutlined, DownloadOutlined, DownOutlined, FileTextOutlined,
  FontSizeOutlined, HistoryOutlined, LayoutOutlined, PlusOutlined, ProjectOutlined, CommentOutlined,
  ReadOutlined, RobotOutlined, SafetyCertificateOutlined, SaveOutlined, ToolOutlined,
  UserOutlined, SendOutlined,
  ReloadOutlined, LoadingOutlined, ThunderboltOutlined, ExclamationCircleOutlined,
  FileSearchOutlined, BulbOutlined, RocketOutlined,
} from '@ant-design/icons-vue'
import { useResumeStore } from '@/stores/resume'
import { useUserProfileStore } from '@/stores/userProfile'
import { getResumeTemplate, resumeTemplates, type ResumeTemplateId } from '@/data/resumeTemplates'
import type { ResumeVersion } from '@/types/models'

interface ResumeEducation { school: string; major: string; degree: string; start: string; end: string; courses: string; gpa: string }
interface ResumeWork { company: string; position: string; start: string; end: string; description: string; leave_reason: string }
interface ResumeProject { name: string; role: string; start: string; end: string; description: string; tech_stack: string[]; url: string }
interface ResumeSkill { category: string; name: string; proficiency: string }
interface ResumeContent {
  template_style: ResumeTemplateId
  personal: { name: string; gender: string; age: string; phone: string; email: string; github: string; avatar: string; city: string }
  intention: { position: string; city: string; salary: string; arrival: string; industry: string }
  education: ResumeEducation[]; work: ResumeWork[]; project: ResumeProject[]; skills: ResumeSkill[]
  honor: unknown[]; custom: unknown[]; module_order: string[]; module_visibility: Record<string, boolean>
}

const route = useRoute()
const router = useRouter()
const resumeStore = useResumeStore()
const userProfileStore = useUserProfileStore()
const demoMode = computed(() => route.name === 'ResumeLabPreview' || route.query.demo === '1')
const resumeId = computed(() => Number(route.params.id))
const editableName = ref('后端工程师－张明')
const templateStyle = ref<ResumeTemplateId>('classic')
const fontFamily = ref('sans')
const density = ref('comfortable')
const zoom = ref(82)
const versionDrawerOpen = ref(false)
const showSaveVersionModal = ref(false)
const savingVersion = ref(false)
const newVersionNote = ref('')
const targetJd = ref('')
const collapsed = reactive<Record<string, boolean>>({ personal: false, intention: false, education: false, work: false, project: false, skills: false })
const resumeContent = reactive<ResumeContent>(createSampleContent())

const serializedContent = computed(() => JSON.stringify(resumeContent, null, 2))
const contentStats = computed(() => ({
  chars: serializedContent.value.length,
  sections:
    (resumeContent.module_visibility.personal !== false ? 1 : 0)
    + (resumeContent.module_visibility.intention !== false ? 1 : 0)
    + (resumeContent.module_visibility.education !== false && resumeContent.education.length ? 1 : 0)
    + (resumeContent.module_visibility.work !== false && resumeContent.work.length ? 1 : 0)
    + (resumeContent.module_visibility.project !== false && resumeContent.project.length ? 1 : 0)
    + (resumeContent.module_visibility.skills !== false && resumeContent.skills.length ? 1 : 0),
}))
const intentionSummary = computed(() => [resumeContent.intention.position, resumeContent.intention.city, resumeContent.intention.salary, resumeContent.intention.arrival].filter(Boolean).join(' ｜ '))

// ==================== 智能完善：AI 大模型分析 ====================
type EditorTab = 'edit' | 'smart'
const editorTab = ref<EditorTab>('edit')

type SmartTabKey = 'analysis' | 'jd' | 'optimize'
const activeSmartTab = ref<SmartTabKey>('analysis')

const smartTabs = [
  { key: 'analysis' as const, label: 'AI 分析', icon: BulbOutlined },
  { key: 'jd' as const, label: 'JD 匹配', icon: FileSearchOutlined },
  { key: 'optimize' as const, label: 'AI 优化', icon: RocketOutlined },
]

// —— 分析相关 ——
const analysisLoading = ref(false)

// 多维度评分（基于内容实时计算 + 模拟 AI 权重）
const analysisDimensions = computed(() => {
  const personal = resumeContent.personal
  const edu = resumeContent.education
  const work = resumeContent.work
  const proj = resumeContent.project
  const skills = resumeContent.skills

  const contactComplete = [personal.name, personal.phone, personal.email, personal.city].filter(Boolean).length
  const intentionComplete = [resumeContent.intention.position, resumeContent.intention.city, resumeContent.intention.salary].filter(Boolean).length

  // 量化指标占比
  const allDesc = [...work, ...proj].map((i) => i.description).join(' ')
  const hasQuant = /[0-9]+(%|万|倍|ms|次|秒|分|倍|x|X)/.test(allDesc)
  const quantCount = (allDesc.match(/[0-9]+(%|万|倍|ms|次|秒|分|x|X)/g) || []).length

  return [
    { label: '基本信息完整度', value: Math.min(98, 40 + contactComplete * 15), color: '#087443' },
    { label: '求职意向清晰度', value: intentionComplete >= 3 ? 92 : 50 + intentionComplete * 14, color: '#007aff' },
    { label: '教育背景', value: Math.min(95, 50 + edu.length * 22), color: '#5e5ce6' },
    { label: '工作经历', value: Math.min(96, 45 + work.length * 18 + (hasQuant ? 8 : 0)), color: '#af52de' },
    { label: '项目经验', value: Math.min(94, 42 + proj.length * 22 + (quantCount >= 3 ? 12 : 0)), color: '#ff9500' },
    { label: '技能描述', value: Math.min(92, 48 + skills.length * 12), color: '#34c759' },
  ]
})

const analysisScore = computed(() => {
  const ds = analysisDimensions.value
  return Math.round(ds.reduce((a, b) => a + b.value, 0) / ds.length)
})

const analysisScoreLabel = computed(() => {
  const s = analysisScore.value
  if (s >= 85) return '优秀'
  if (s >= 70) return '良好'
  if (s >= 55) return '中等'
  return '待完善'
})

const analysisScoreDesc = computed(() => {
  const s = analysisScore.value
  if (s >= 85) return '简历内容完整且量化指标充足，仅需微调即可投递。'
  if (s >= 70) return '整体表现不错，建议补充亮点项目与量化成果。'
  if (s >= 55) return '基础信息已具备，但部分模块需进一步完善。'
  return '简历内容较少，建议从基本信息、工作经历和项目经验开始完善。'
})

// 亮点分析
const analysisHighlights = computed(() => {
  const list: { title: string; detail: string }[] = []
  const work = resumeContent.work
  const proj = resumeContent.project
  const allDesc = [...work, ...proj].map((i) => i.description).join(' ')
  const quantCount = (allDesc.match(/[0-9]+(%|万|倍|ms|次|秒|分|x|X)/g) || []).length

  if (quantCount >= 3) {
    list.push({ title: '量化成果突出', detail: `已使用 ${quantCount} 处量化指标，能有效展示业务影响力。` })
  }
  if (proj.length >= 2) {
    list.push({ title: '项目证据充分', detail: `共 ${proj.length} 个项目，能多角度展示技术能力。` })
  }
  if (work.length >= 2) {
    list.push({ title: '工作经历稳定', detail: `${work.length} 段工作经历，体现持续成长。` })
  }
  if (resumeContent.skills.length >= 4) {
    list.push({ title: '技能栈丰富', detail: `${resumeContent.skills.length} 项技能分类，覆盖面广。` })
  }
  return list
})

// 待补强项
const analysisWeakness = computed(() => {
  const list: {
    title: string
    detail: string
    actionLabel?: string
    action?: () => void
  }[] = []
  const work = resumeContent.work
  const proj = resumeContent.project
  const skills = resumeContent.skills
  const allDesc = [...work, ...proj].map((i) => i.description).join(' ')
  const hasQuant = /[0-9]+(%|万|倍|ms|次|秒|分)/.test(allDesc)

  if (!resumeContent.intention.position) {
    list.push({
      title: '求职意向缺失',
      detail: '未填写目标岗位，AI 难以做精准匹配。',
      actionLabel: '填写意向',
      action: () => { activeSmartTab.value = 'analysis'; message.info('请在左侧「求职意向」模块填写') },
    })
  }
  if (!hasQuant) {
    list.push({
      title: '缺少量化成果',
      detail: '工作与项目描述未使用数据指标，建议补充性能、规模、效率等量化结果。',
      actionLabel: 'AI 优化',
      action: () => { activeSmartTab.value = 'optimize' },
    })
  }
  if (proj.length < 2) {
    list.push({ title: '项目数量偏少', detail: `当前仅 ${proj.length} 个项目，建议补充 1-2 个能体现核心能力的项目。` })
  }
  if (skills.length < 3) {
    list.push({ title: '技能描述单薄', detail: `仅 ${skills.length} 项技能，建议按「编程语言 / 框架 / 工具」分类补全。` })
  }
  if (!resumeContent.personal.email || !resumeContent.personal.phone) {
    list.push({ title: '联系方式不全', detail: '邮箱或电话缺失，HR 无法联系到你。' })
  }
  return list
})

// 模拟 AI 分析（带 loading 动画）
const runAnalysis = async () => {
  analysisLoading.value = true
  await new Promise((r) => setTimeout(r, 800))
  analysisLoading.value = false
  message.success('AI 分析已更新')
}

// ==================== JD 智能匹配 ====================
const jdLoading = ref(false)
interface JdMatchResult {
  matchRate: number
  matched: string[]
  missing: string[]
  suggest: { title: string; detail: string }[]
}
const jdMatchResult = ref<JdMatchResult | null>(null)

// 关键词库（按技术栈分类，可扩展）
const techKeywords = [
  'Java', 'Go', 'Python', 'Node.js', 'C++', 'Rust', 'TypeScript', 'JavaScript',
  'Spring Boot', 'Spring Cloud', 'Gin', 'React', 'Vue', 'Angular', 'Next.js',
  'MySQL', 'PostgreSQL', 'Redis', 'MongoDB', 'ClickHouse', 'Elasticsearch', 'Kafka', 'RabbitMQ', 'RocketMQ',
  'Docker', 'Kubernetes', 'K8s', 'Jenkins', 'CI/CD', 'Prometheus', 'Grafana',
  '微服务', '分布式', '高并发', '高可用', '性能优化', '架构设计',
  '消息队列', '缓存', '数据库优化', '系统设计', 'DDD', '领域驱动设计',
  '机器学习', '深度学习', 'LLM', 'NLP', '推荐系统', '数据分析',
  '团队管理', '项目管理', '敏捷开发', 'Scrum', 'OKR',
  'Linux', 'Git', 'Shell', 'Python', '算法', '数据结构',
  'TOEIC', '英语', 'PMP',
]

const runJdMatch = async () => {
  if (!targetJd.value.trim()) {
    message.warning('请先粘贴 JD 内容')
    return
  }
  openCopilot('jd_match')
}

const openCopilot = (task: 'jd_match' | 'project_optimize' | 'interview_predict' | 'career_chat', projectIndex?: number, question?: string) => {
  if (!demoMode.value && resumeId.value) {
    localStorage.setItem('zhitu-copilot-draft', JSON.stringify({
      resume_id: resumeId.value,
      content: JSON.stringify(resumeContent),
    }))
  }
  router.push({
    path: '/app/copilot',
    query: {
      resume_id: resumeId.value ? String(resumeId.value) : undefined,
      task,
      project_index: projectIndex === undefined ? undefined : String(projectIndex),
      jd: targetJd.value.trim() || undefined,
      question,
    },
  })
}

const jdMatchColor = computed(() => {
  const r = jdMatchResult.value?.matchRate ?? 0
  if (r >= 80) return '#34c759'
  if (r >= 60) return '#ff9500'
  return '#ff3b30'
})
const jdMatchLabel = computed(() => {
  const r = jdMatchResult.value?.matchRate ?? 0
  if (r >= 80) return '匹配良好'
  if (r >= 60) return '部分匹配'
  return '匹配度低'
})

// ==================== AI 一键优化 ====================
const optimizeLoading = ref(false)
const optStatus = reactive<Record<string, 'idle' | 'loading' | 'done'>>({
  work: 'idle', project: 'idle', skills: 'idle', summary: 'idle',
})

const optimizeItem = (key: keyof typeof optStatus) => {
  const questions: Record<string, string> = {
    work: '请帮我分析工作经历，指出哪些内容应该量化和如何改写。',
    project: '请帮我找出项目经历中最值得强化的技术亮点。',
    skills: '请根据当前目标岗位，帮我重组技能描述并指出缺失项。',
    summary: '请根据当前简历生成一版真实、不夸大的个人简介候选文案。',
  }
  openCopilot('career_chat', undefined, questions[key])
}

const optimizeItems = computed(() => [
  {
    key: 'work',
    title: '工作经历强化',
    desc: 'AI 自动补充量化指标、业务影响力描述',
    buttonText: '优化工作经历',
    loading: optStatus.work === 'loading',
    statusText: optStatus.work === 'done' ? '已优化' : (optStatus.work === 'loading' ? '处理中' : '待优化'),
    statusClass: optStatus.work === 'done' ? 'st-done' : 'st-idle',
    run: () => optimizeItem('work'),
  },
  {
    key: 'project',
    title: '项目经历强化',
    desc: '补充技术栈细节、性能指标与业务结果',
    buttonText: '优化项目经历',
    loading: optStatus.project === 'loading',
    statusText: optStatus.project === 'done' ? '已优化' : (optStatus.project === 'loading' ? '处理中' : '待优化'),
    statusClass: optStatus.project === 'done' ? 'st-done' : 'st-idle',
    run: () => openCopilot('project_optimize', 0),
  },
  {
    key: 'skills',
    title: '技能描述重组',
    desc: '按熟练度排序、补充分类与年限',
    buttonText: '优化技能描述',
    loading: optStatus.skills === 'loading',
    statusText: optStatus.skills === 'done' ? '已优化' : (optStatus.skills === 'loading' ? '处理中' : '待优化'),
    statusClass: optStatus.skills === 'done' ? 'st-done' : 'st-idle',
    run: () => optimizeItem('skills'),
  },
  {
    key: 'summary',
    title: '生成个人简介',
    desc: '基于经历自动生成一句话自我介绍',
    buttonText: 'AI 生成简介',
    loading: optStatus.summary === 'loading',
    statusText: optStatus.summary === 'done' ? '已生成' : (optStatus.summary === 'loading' ? '处理中' : '待生成'),
    statusClass: optStatus.summary === 'done' ? 'st-done' : 'st-idle',
    run: () => optimizeItem('summary'),
  },
])

const optimizeAll = async () => {
  openCopilot('career_chat', undefined, '请从整体角度分析这份简历，给出最值得优先修改的三项建议。')
}

watch(() => resumeStore.currentVersion, (version) => {
  if (!version || demoMode.value) return
  try {
    const normalized = normalizeContent(JSON.parse(version.content))
    Object.assign(resumeContent, normalized)
    templateStyle.value = normalized.template_style
  } catch {
    message.warning('当前版本内容不是可识别的结构化简历 JSON')
  }
})

watch(templateStyle, (value) => {
  resumeContent.template_style = value
})

function createSampleContent(): ResumeContent {
  return normalizeContent({
    personal: { name: '张明', phone: '138-0000-0000', email: 'zhangming@email.com', city: '上海市浦东新区', github: 'github.com/zhangming' },
    intention: { position: '后端工程师', city: '上海', salary: '25K–35K', arrival: '1个月内', industry: '互联网 / AI' },
    education: [{ school: '上海交通大学', major: '计算机科学与技术', degree: '本科', start: '2016.09', end: '2020.06', courses: 'GPA 3.7/4.0 · 数据结构、操作系统、数据库系统', gpa: '3.7/4.0' }],
    work: [
      { company: '上海云舟科技有限公司', position: '后端工程师', start: '2021.07', end: '至今', description: '负责公司核心业务系统设计与开发，支撑日均 1000 万+ 请求量\n拆分单体服务为 12 个微服务，系统可用性提升至 99.95%\n优化数据库慢查询与缓存策略，接口平均响应时间从 120ms 降至 45ms\n推动 CI/CD 落地，发布效率提升 70%', leave_reason: '' },
      { company: '上海智汇互联有限公司', position: '后端开发工程师', start: '2020.07', end: '2021.06', description: '参与电商平台后端开发，负责订单、支付与库存核心模块\n基于 Spring Cloud Alibaba 完成服务治理与配置管理', leave_reason: '' },
    ],
    project: [
      { name: '云舟微服务平台重构项目', role: '后端负责人', start: '2022.03', end: '2022.11', description: '设计并实现基于 Spring Cloud 的微服务框架，集成 Gateway、Sentinel、SkyWalking\n项目上线后系统稳定性提升 80%，运维成本降低 30%', tech_stack: ['Spring Cloud', 'MySQL', 'Redis'], url: '' },
      { name: '实时数据分析平台', role: '核心开发', start: '2021.09', end: '2022.02', description: '使用 Kafka + Flink 实现实时流处理，数据延迟从分钟级降至毫秒级\n基于 ClickHouse 建设查询链路，查询性能提升 5 倍', tech_stack: ['Kafka', 'Flink', 'ClickHouse'], url: '' },
    ],
    skills: [
      { category: '编程语言', name: 'Java（熟练）、Go（熟练）、Python（熟悉）', proficiency: '' },
      { category: '框架', name: 'Spring Boot、Spring Cloud Alibaba、Gin、MyBatis', proficiency: '' },
      { category: '数据库', name: 'MySQL、Redis、MongoDB、ClickHouse', proficiency: '' },
      { category: '工程工具', name: 'Docker、Kubernetes、Prometheus、Grafana', proficiency: '' },
    ],
  })
}

function normalizeContent(value: any): ResumeContent {
  return {
    template_style: getResumeTemplate(value.template_style).id,
    personal: { name: '', gender: '', age: '', phone: '', email: '', github: '', avatar: '', city: '', ...(value.personal || {}) },
    intention: { position: '', city: '', salary: '', arrival: '', industry: '', ...(value.intention || {}) },
    education: Array.isArray(value.education) ? value.education : [],
    work: Array.isArray(value.work) ? value.work : [],
    project: Array.isArray(value.project) ? value.project.map((item) => ({ ...item, tech_stack: Array.isArray(item.tech_stack) ? item.tech_stack : [] })) : [],
    skills: Array.isArray(value.skills) ? value.skills : [],
    honor: Array.isArray(value.honor) ? value.honor : [], custom: Array.isArray(value.custom) ? value.custom : [],
    module_order: value.module_order || ['personal', 'intention', 'education', 'work', 'project', 'skills', 'honor'],
    module_visibility: {
      personal: true, intention: true, education: true, work: true, project: true, skills: true, honor: true,
      ...(value.module_visibility || {}),
    },
  }
}

const toggleSection = (key: string) => { collapsed[key] = !collapsed[key] }
const isVisible = (key: string) => resumeContent.module_visibility[key] !== false
const toLines = (text: string) => text.split(/\n+/).map((line) => line.replace(/^[•·\-]\s*/, '').trim()).filter(Boolean)
const addEducation = () => resumeContent.education.push({ school: '', major: '', degree: '', start: '', end: '', courses: '', gpa: '' })
const addWork = () => resumeContent.work.push({ company: '', position: '', start: '', end: '', description: '', leave_reason: '' })
const addProject = () => resumeContent.project.push({ name: '', role: '', start: '', end: '', description: '', tech_stack: [], url: '' })
const addSkill = () => resumeContent.skills.push({ category: '', name: '', proficiency: '' })
const updateTechStack = (item: ResumeProject, event: Event) => { item.tech_stack = (event.target as HTMLInputElement).value.split(/[,，]/).map((value) => value.trim()).filter(Boolean) }
const selectVersion = (version: ResumeVersion) => { resumeStore.setCurrentVersion(version); versionDrawerOpen.value = false }

const handleNameSave = async () => {
  if (demoMode.value || !resumeStore.currentResume || !editableName.value.trim()) return
  await resumeStore.update(resumeId.value, { name: editableName.value.trim() })
}
const handleSaveVersion = async () => {
  if (!newVersionNote.value.trim()) return message.warning('请输入版本备注')
  if (demoMode.value) { showSaveVersionModal.value = false; newVersionNote.value = ''; return message.success('本地预览：版本已模拟保存') }
  savingVersion.value = true
  const version = await resumeStore.createVersion(resumeId.value, { content: JSON.stringify(resumeContent), change_note: newVersionNote.value.trim() })
  savingVersion.value = false
  if (version) { showSaveVersionModal.value = false; newVersionNote.value = '' }
}
const previewNavigate = (target: 'interview' | 'delivery') => {
  message.info(target === 'interview' ? '面试训练场将在下一阶段完善' : '投递看板将在下一阶段完善')
}
const printAncestorClass = 'resume-print-ancestor'
const prepareResumePrint = () => {
  const paper = document.getElementById('resume-paper')
  if (!paper) return

  document.body.classList.add('resume-printing')
  let ancestor = paper.parentElement
  while (ancestor && ancestor !== document.body) {
    ancestor.classList.add(printAncestorClass)
    ancestor = ancestor.parentElement
  }
}
const cleanupResumePrint = () => {
  document.body.classList.remove('resume-printing')
  document.querySelectorAll(`.${printAncestorClass}`).forEach((element) => {
    element.classList.remove(printAncestorClass)
  })
}
const exportResume = () => {
  prepareResumePrint()
  window.print()
}
const backToList = () => demoMode.value ? router.push('/') : router.push('/app/resumes')
const formatVersionDate = (value: string) => value ? new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : ''

// 从个人资料 store 同步基础信息到简历 personal 字段
// 只在用户已填写个人资料时同步，且会覆盖示例默认值
const syncFromUserProfile = () => {
  if (!userProfileStore.hasFilled) return
  const b = userProfileStore.basic
  if (b.name) resumeContent.personal.name = b.name
  if (b.phone) resumeContent.personal.phone = b.phone
  if (b.email) resumeContent.personal.email = b.email
  if (b.city) resumeContent.personal.city = b.city
  if (b.github) resumeContent.personal.github = b.github
}

onMounted(async () => {
  document.documentElement.classList.add('resume-editor-scroll-lock')
  document.body.classList.add('resume-editor-scroll-lock')
  window.addEventListener('beforeprint', prepareResumePrint)
  window.addEventListener('afterprint', cleanupResumePrint)

  if (demoMode.value) return
  if (!resumeId.value || Number.isNaN(resumeId.value)) return router.push('/app/resumes')
  await resumeStore.fetchOne(resumeId.value)
  editableName.value = resumeStore.currentResume?.name || '未命名简历'
  targetJd.value = resumeStore.currentResume?.target_jd || targetJd.value
  await resumeStore.fetchVersions(resumeId.value)
  if (resumeStore.versions.length) resumeStore.setCurrentVersion(resumeStore.versions[0])
  // 加载完简历版本后，同步个人资料
  syncFromUserProfile()
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeprint', prepareResumePrint)
  window.removeEventListener('afterprint', cleanupResumePrint)
  cleanupResumePrint()
  document.documentElement.classList.remove('resume-editor-scroll-lock')
  document.body.classList.remove('resume-editor-scroll-lock')
})

// 监听个人资料变化，实时同步到简历（用户在弹窗保存后立即生效）
watch(
  () => userProfileStore.basic,
  () => syncFromUserProfile(),
  { deep: true }
)
</script>

<style scoped src="./styles/resume-editor.css"></style>
