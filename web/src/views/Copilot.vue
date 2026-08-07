<template>
  <div class="copilot-page">
    <header class="copilot-header">
      <div>
        <p class="eyebrow">职途 · 求职工作台</p>
        <h1>求职 Copilot</h1>
        <p class="subtitle">围绕你的真实简历和目标岗位，分析差距、打磨项目、预测面试。</p>
      </div>
      <a-button @click="clearLocalSessions">清除本地对话</a-button>
    </header>

    <div class="copilot-layout">
      <aside class="copilot-sidebar">
        <section class="context-panel">
          <div class="panel-title"><span>工作上下文</span><small>仅用于当前对话</small></div>
          <a-select v-model:value="selectedResumeId" allow-clear placeholder="选择简历，或清空后使用粘贴内容" :loading="loadingResumes" @change="handleResumeChange">
            <a-select-option v-for="resume in resumes" :key="resume.id" :value="resume.id">{{ resume.name }}</a-select-option>
          </a-select>
          <a-select v-model:value="selectedVersionId" placeholder="选择版本" :loading="loadingVersions" :disabled="!selectedResumeId" @change="loadVersionContent">
            <a-select-option v-for="version in versions" :key="version.id" :value="version.id">{{ version.version_label }}</a-select-option>
          </a-select>
          <a-textarea v-model:value="jd" :rows="5" placeholder="粘贴目标岗位 JD（匹配和面试预测需要）" />
          <a-textarea v-model:value="resumePaste" :rows="3" placeholder="没有现成简历时可直接粘贴；已有简历可补充一段经历或数据" />
          <label class="file-input"><UploadOutlined /> 上传 TXT / Markdown 简历 <input type="file" accept=".txt,.md,text/plain,text/markdown" @change="handleResumeFile" /></label>
          <div class="context-chip" v-if="selectedResume">
            <FileTextOutlined /> {{ selectedResume.name }} · {{ selectedVersionLabel }}
          </div>
        </section>

        <section class="task-panel">
          <div class="panel-title"><span>选择任务</span><small>每个任务独立记忆</small></div>
          <button v-for="item in taskItems" :key="item.key" type="button" class="task-button" :class="{ active: task === item.key }" @click="switchTask(item.key)">
            <component :is="item.icon" />
            <span><strong>{{ item.title }}</strong><small>{{ item.description }}</small></span>
          </button>
        </section>

        <section class="history-panel">
          <div class="panel-title"><span>本地对话</span><small>{{ copilotStore.sessions.length }} 个</small></div>
          <button v-for="session in copilotStore.sessions" :key="session.id" type="button" class="history-item" :class="{ active: session.id === copilotStore.activeSessionId }" @click="selectExistingSession(session.id)">
            <span>{{ taskLabel(session.task) }}</span>
            <small>{{ formatTime(session.updated_at) }}</small>
          </button>
          <a-empty v-if="!copilotStore.sessions.length" :image="null" description="还没有本地对话" />
        </section>
      </aside>

      <main class="copilot-main">
        <div v-if="task === 'project_optimize'" class="project-picker">
          <span>优化项目</span>
          <a-select v-model:value="projectIndex" :disabled="!currentContent.project.length" placeholder="选择一个项目">
            <a-select-option v-for="(project, index) in currentContent.project" :key="index" :value="index">{{ project.name || `项目 ${index + 1}` }}</a-select-option>
          </a-select>
        </div>
        <div class="chat-shell">
          <div class="chat-title">
            <div><strong>{{ taskLabel(task) }}</strong><span v-if="activeSession"> · 本地保存</span></div>
            <a-button v-if="activeSession" type="text" @click="newConversation">新对话</a-button>
          </div>
          <div class="chat-messages">
            <div v-if="!activeSession || !activeSession.messages.length" class="chat-welcome">
              <RobotOutlined />
              <h2>{{ welcomeTitle }}</h2>
              <p>{{ welcomeDescription }}</p>
              <div class="prompt-row">
                <button v-for="prompt in quickPrompts" :key="prompt" type="button" @click="sendMessage(prompt)">{{ prompt }}</button>
              </div>
            </div>
            <article v-for="(msg, index) in activeSession?.messages || []" :key="`${msg.created_at}-${index}`" class="chat-message" :class="msg.role">
              <div class="message-avatar">{{ msg.role === 'assistant' ? 'AI' : '我' }}</div>
              <div class="message-body">
                <div class="message-content" v-text="msg.content"></div>
                <div v-if="msg.result" class="result-card">
                  <div v-if="msg.result.match" class="match-result">
                    <div class="result-score"><strong>{{ msg.result.match.match_score }}</strong><small>/100 匹配度</small></div>
                    <div class="result-columns">
                      <div><h4>优势点</h4><p v-for="item in msg.result.match.strengths" :key="item">{{ item }}</p></div>
                      <div><h4>缺失能力</h4><p v-for="item in msg.result.match.missing_capabilities" :key="item">{{ item }}</p></div>
                    </div>
                    <div class="requirement-list"><span v-for="item in msg.result.match.requirement_map" :key="item.title" :class="`status-${item.status}`">{{ item.title }} · {{ item.status }}</span></div>
                  </div>
                  <div v-if="msg.result.project" class="project-result">
                    <h4>项目改写候选</h4>
                    <pre>{{ msg.result.project.rewritten_description }}</pre>
                    <div class="result-actions" v-if="msg.result.proposals?.length">
                      <a-button type="primary" size="small" :loading="applying" @click="applyProposal(msg.result.proposals[0])">应用为新版本</a-button>
                    </div>
                    <p v-for="item in msg.result.project.missing_evidence" :key="item" class="muted">待补充：{{ item }}</p>
                  </div>
                  <div v-if="msg.result.prediction" class="prediction-result">
                    <h4>面试预测</h4>
                    <article v-for="item in msg.result.prediction.questions" :key="item.question" class="prediction-item">
                      <strong>{{ item.question }}</strong><small>{{ item.type }} · {{ item.priority }}</small><p>{{ item.answer_plan }}</p>
                    </article>
                  </div>
                </div>
              </div>
            </article>
            <div v-if="copilotStore.loading" class="typing"><LoadingOutlined /> Copilot 正在整理建议…</div>
          </div>
          <div class="composer">
            <a-textarea v-model:value="input" :rows="3" :disabled="copilotStore.loading" :placeholder="composerPlaceholder" @keydown.enter.exact="handleEnter" />
            <div class="composer-foot"><span>对话会保存在当前浏览器，不会写入服务端聊天记录</span><a-button type="primary" :loading="copilotStore.loading" :disabled="!input.trim() || (!selectedResumeId && !resumePaste.trim())" @click="sendMessage()"><SendOutlined />发送</a-button></div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { BulbOutlined, FileSearchOutlined, LoadingOutlined, MessageOutlined, RobotOutlined, RocketOutlined, SendOutlined, FileTextOutlined, UploadOutlined } from '@ant-design/icons-vue'
import { listResumes, listVersions, getVersion } from '@/api/resume'
import { applyCopilotProposal } from '@/api/copilot'
import { useCopilotStore } from '@/stores/copilot'
import type { CopilotProposal, CopilotTask, Resume, ResumeVersion } from '@/types/models'

const copilotStore = useCopilotStore()
const route = useRoute()
const resumes = ref<Resume[]>([])
const versions = ref<ResumeVersion[]>([])
const selectedResumeId = ref<number | null>(null)
const selectedVersionId = ref<number | null>(null)
const currentContent = reactive<any>({ project: [] })
const loadingResumes = ref(false)
const loadingVersions = ref(false)
const jd = ref('')
const resumePaste = ref('')
const draftOverride = ref('')
const task = ref<CopilotTask>('jd_match')
const projectIndex = ref(0)
const input = ref('')
const applying = ref(false)

const selectedResume = computed(() => resumes.value.find((item) => item.id === selectedResumeId.value))
const selectedVersionLabel = computed(() => versions.value.find((item) => item.id === selectedVersionId.value)?.version_label || '当前版本')
const activeSession = computed(() => copilotStore.activeSession)

const taskItems = [
  { key: 'jd_match' as const, title: '简历-JD 匹配', description: '看清优势和缺口', icon: FileSearchOutlined },
  { key: 'project_optimize' as const, title: '项目经历优化', description: '按 STAR 打磨文案', icon: RocketOutlined },
  { key: 'interview_predict' as const, title: '岗位风险预测', description: '提前准备高频问题', icon: BulbOutlined },
  { key: 'career_chat' as const, title: '求职问答', description: '自由聊求职问题', icon: MessageOutlined },
]

const currentVersion = computed(() => versions.value.find((item) => item.id === selectedVersionId.value))
const updateCurrentContent = (raw: string) => {
  try {
    const parsed = JSON.parse(raw || '{}')
    Object.assign(currentContent, parsed)
    currentContent.project = Array.isArray(parsed.project) ? parsed.project : []
  } catch {
    currentContent.project = []
  }
}

const currentDraftContent = computed(() => {
  if (draftOverride.value) return draftOverride.value
  if (resumePaste.value.trim() && !selectedResumeId.value) return resumePaste.value.trim()
  if (!resumePaste.value.trim()) return currentVersion.value?.content || ''
  try {
    const parsed = JSON.parse(currentVersion.value?.content || '{}')
    parsed.custom = Array.isArray(parsed.custom) ? parsed.custom : []
    parsed.custom = [...parsed.custom, { title: '本轮补充经历', content: resumePaste.value.trim() }]
    return JSON.stringify(parsed)
  } catch { return currentVersion.value?.content || '' }
})
const sessionDraftContent = computed(() => {
  if (!selectedResumeId.value) return currentDraftContent.value
  return (draftOverride.value || resumePaste.value).trim() ? currentDraftContent.value : ''
})

const taskLabel = (value: CopilotTask) => taskItems.find((item) => item.key === value)?.title || value
const formatTime = (value: string) => new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })

const loadVersionContent = async () => {
  if (!selectedResumeId.value || !selectedVersionId.value) return
  const response = await getVersion(selectedResumeId.value, selectedVersionId.value)
  try {
    Object.assign(currentContent, JSON.parse(response.data.data?.content || '{}'))
    currentContent.project = Array.isArray(currentContent.project) ? currentContent.project : []
  } catch { currentContent.project = [] }
}

const loadVersionsForResume = async (resumeId: number, preferredVersionId?: number | null) => {
  loadingVersions.value = true
  try {
    const response = await listVersions(resumeId)
    versions.value = response.data.data || []
    selectedVersionId.value = versions.value.some((item) => item.id === preferredVersionId)
      ? preferredVersionId || null
      : (versions.value[0]?.id || null)
    await loadVersionContent()
  } finally { loadingVersions.value = false }
}

const handleResumeChange = async (resumeId?: number) => {
  if (!resumeId) {
    selectedResumeId.value = null
    versions.value = []
    selectedVersionId.value = null
    currentContent.project = []
    jd.value = ''
    return
  }
  draftOverride.value = ''
  resumePaste.value = ''
  jd.value = resumes.value.find((item) => item.id === resumeId)?.target_jd || ''
  await loadVersionsForResume(resumeId)
}

const restoreSessionDraft = (draft: string) => {
  draftOverride.value = ''
  resumePaste.value = ''
  if (!draft.trim()) return
  try {
    JSON.parse(draft)
    draftOverride.value = draft
  } catch {
    resumePaste.value = draft
  }
}

const newConversation = async () => {
  if (!selectedResumeId.value && !resumePaste.value.trim()) return
  await copilotStore.createSession({ task: task.value, resumeId: selectedResumeId.value, versionId: selectedVersionId.value, jd: jd.value.trim(), draftContent: sessionDraftContent.value, projectIndex: task.value === 'project_optimize' ? projectIndex.value : undefined })
}

const switchTask = async (next: CopilotTask) => {
  task.value = next
  if (next === 'project_optimize' && projectIndex.value >= currentContent.project.length) projectIndex.value = 0
  await newConversation()
}

const selectExistingSession = (id: string) => {
  copilotStore.selectSession(id)
  const session = copilotStore.sessions.find((item) => item.id === id)
  if (!session) return
  task.value = session.task
  selectedResumeId.value = session.resume_id
  selectedVersionId.value = session.version_id
  jd.value = session.jd
  restoreSessionDraft(session.draft_content || '')
  if (!selectedResumeId.value && (draftOverride.value || resumePaste.value)) updateCurrentContent(draftOverride.value || resumePaste.value)
  projectIndex.value = session.project_index || 0
  if (selectedResumeId.value) void loadVersionsForResume(selectedResumeId.value, session.version_id)
}

const ensureSession = async () => {
  if (activeSession.value && activeSession.value.task === task.value && activeSession.value.resume_id === selectedResumeId.value && activeSession.value.version_id === selectedVersionId.value && activeSession.value.jd === jd.value.trim() && activeSession.value.draft_content === sessionDraftContent.value && (task.value !== 'project_optimize' || activeSession.value.project_index === projectIndex.value)) return activeSession.value
  await newConversation()
  return activeSession.value
}

const sendMessage = async (quick?: string) => {
  const content = (quick || input.value).trim()
  if (!content || (!selectedResumeId.value && !resumePaste.value.trim())) return
  const session = await ensureSession()
  if (!session) return
  input.value = ''
  await copilotStore.send({ session, content, draftContent: sessionDraftContent.value })
}

const handleEnter = (event: KeyboardEvent) => {
  if (event.isComposing || event.shiftKey) return
  event.preventDefault()
  void sendMessage()
}

const applyProposal = async (proposal: CopilotProposal) => {
  if (!selectedResumeId.value || !selectedVersionId.value || !currentVersion.value) return
  applying.value = true
  try {
    const response = await applyCopilotProposal({
      resume_id: selectedResumeId.value,
      base_version_id: selectedVersionId.value,
      content: currentDraftContent.value,
      project_index: proposal.project_index,
      replacement_description: proposal.replacement_description,
      replacement_tech_stack: proposal.replacement_tech_stack,
      change_note: `Copilot：${proposal.title}`,
    })
    const newVersion = (response as any).data?.data as ResumeVersion | undefined
    if (newVersion) {
      selectedVersionId.value = newVersion.id
      versions.value.unshift(newVersion)
      await loadVersionContent()
    }
    message.success('已生成新的简历版本，原版本仍保留')
  } catch (error) {
    message.error(error instanceof Error ? error.message : '应用建议失败')
  } finally { applying.value = false }
}

const welcomeTitle = computed(() => task.value === 'career_chat' ? '先问一个求职问题' : `开始${taskLabel(task.value)}`)
const welcomeDescription = computed(() => task.value === 'project_optimize' ? '选择项目后，我会先指出 STAR 结构和证据缺口，再和你一起改写。' : '我会读取当前简历上下文，给出可验证、可执行的建议。')
const composerPlaceholder = computed(() => task.value === 'career_chat' ? '例如：大三找实习，简历需要写几个项目？' : '告诉我你最关心的部分，支持继续追问和反驳…')
const quickPrompts = computed(() => task.value === 'jd_match' ? ['先给我一个匹配度结论', '哪些经历应该优先强化？'] : task.value === 'project_optimize' ? ['先分析这个项目的问题', '按实习校招风格改写'] : task.value === 'interview_predict' ? ['列出最高频的技术问题', '哪些简历细节最容易被追问？'] : ['Agent 项目简历怎么写？', '大三找实习要准备什么？'])

const clearLocalSessions = async () => {
  await copilotStore.clearAll()
  message.success('本地 Copilot 对话已清除')
}

const handleResumeFile = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  resumePaste.value = await file.text()
  if (!selectedResumeId.value) updateCurrentContent(resumePaste.value)
}

watch(resumePaste, (value) => {
  if (!selectedResumeId.value && value.trim()) updateCurrentContent(value)
})

onMounted(async () => {
  await copilotStore.init()
  loadingResumes.value = true
  try {
    const response = await listResumes()
    resumes.value = response.data.data || []
    const remembered = activeSession.value
    const requestedResumeId = Number(route.query.resume_id)
    selectedResumeId.value = requestedResumeId || (remembered ? remembered.resume_id : resumes.value[0]?.id || null)
    if (selectedResumeId.value) await handleResumeChange(selectedResumeId.value)
    const queryTask = String(route.query.task || '') as CopilotTask
    const queryHasTask = taskItems.some((item) => item.key === queryTask)
    const handoff = localStorage.getItem('zhitu-copilot-draft')
    if (handoff) {
      try {
        const parsed = JSON.parse(handoff)
        if (!parsed.resume_id || Number(parsed.resume_id) === selectedResumeId.value) {
          draftOverride.value = parsed.content || ''
          if (!selectedResumeId.value) updateCurrentContent(draftOverride.value)
        }
      } catch { /* ignore stale handoff */ }
      localStorage.removeItem('zhitu-copilot-draft')
    }
    if (queryHasTask) task.value = queryTask
    if (route.query.jd) jd.value = String(route.query.jd)
    if (route.query.project_index !== undefined) projectIndex.value = Number(route.query.project_index) || 0
    if (route.query.question) input.value = String(route.query.question)
    if (remembered && !queryHasTask && !requestedResumeId) {
      task.value = remembered.task
      selectedVersionId.value = remembered.version_id || selectedVersionId.value
      jd.value = remembered.jd
      restoreSessionDraft(remembered.draft_content || '')
      if (!selectedResumeId.value && (draftOverride.value || resumePaste.value)) updateCurrentContent(draftOverride.value || resumePaste.value)
      projectIndex.value = remembered.project_index || 0
      await loadVersionContent()
    }
  } catch { message.error('加载简历上下文失败') } finally { loadingResumes.value = false }
})
</script>

<style scoped>
.copilot-page{min-height:calc(100dvh - 64px);padding:28px clamp(18px,4vw,56px) 42px;background:#eef1ed;color:#17342c}.copilot-header{max-width:1480px;margin:0 auto 22px;display:flex;justify-content:space-between;gap:20px;align-items:flex-start}.eyebrow{margin:0 0 6px;color:#b45c36;font-size:12px;font-weight:700;letter-spacing:.08em}.copilot-header h1{margin:0;font-family:"Songti SC","STSong",serif;font-size:36px;line-height:1.1}.subtitle{margin:10px 0 0;color:#64746d}.copilot-layout{max-width:1480px;margin:0 auto;display:grid;grid-template-columns:290px minmax(0,1fr);gap:18px}.copilot-sidebar{display:flex;flex-direction:column;gap:14px}.context-panel,.task-panel,.history-panel,.chat-shell{border:1px solid #cbd4ce;background:#fbfcf9}.context-panel,.task-panel,.history-panel{padding:16px}.context-panel{display:flex;flex-direction:column;gap:10px}.panel-title{display:flex;justify-content:space-between;align-items:center;margin-bottom:3px;font-size:13px;font-weight:700}.panel-title small{color:#82918b;font-weight:400;font-size:11px}.context-chip{padding:8px 9px;background:#edf4ee;color:#426057;font-size:12px}.file-input{display:flex;align-items:center;gap:6px;color:#5b6f65;font-size:12px;cursor:pointer}.file-input input{max-width:145px;font-size:11px}.task-panel{display:flex;flex-direction:column;gap:6px}.task-button,.history-item{border:1px solid transparent;background:transparent;text-align:left;cursor:pointer}.task-button{display:flex;align-items:center;gap:10px;padding:10px}.task-button svg{font-size:18px;color:#b45c36}.task-button span{display:flex;flex-direction:column;gap:2px}.task-button small,.history-item small{color:#7a8982;font-size:11px}.task-button.active{border-color:#b45c36;background:#fbf1ec}.history-panel{max-height:280px;overflow:auto}.history-item{display:flex;width:100%;justify-content:space-between;padding:8px;color:#42564e}.history-item.active{color:#a14e2f;background:#f6ece8}.copilot-main{min-width:0}.project-picker{display:flex;align-items:center;gap:10px;margin-bottom:10px}.project-picker>span{font-size:13px;color:#64746d}.project-picker .ant-select{min-width:260px}.chat-shell{min-height:690px;display:flex;flex-direction:column}.chat-title{padding:14px 18px;border-bottom:1px solid #e1e7e2;display:flex;justify-content:space-between}.chat-title span{color:#84918b;font-size:12px;font-weight:400}.chat-messages{flex:1;min-height:400px;max-height:calc(100dvh - 290px);overflow:auto;padding:22px}.chat-welcome{text-align:center;max-width:560px;margin:90px auto 0;color:#61736b}.chat-welcome>svg{font-size:34px;color:#b45c36}.chat-welcome h2{margin:12px 0 6px;color:#23473b}.chat-welcome p{margin:0}.prompt-row{display:flex;flex-wrap:wrap;justify-content:center;gap:8px;margin-top:18px}.prompt-row button{border:1px solid #ccd7d0;background:#fff;padding:8px 10px;color:#476158;cursor:pointer}.chat-message{display:flex;gap:10px;margin-bottom:20px}.chat-message.user{flex-direction:row-reverse}.message-avatar{width:30px;height:30px;flex:0 0 30px;display:grid;place-items:center;border-radius:50%;background:#b45c36;color:#fff;font-size:11px}.chat-message.user .message-avatar{background:#3d7562}.message-body{max-width:min(820px,85%)}.chat-message.user .message-body{display:flex;justify-content:flex-end}.message-content{padding:11px 14px;background:#f0f4f0;white-space:pre-wrap;line-height:1.65}.chat-message.user .message-content{background:#e7efe9}.result-card{margin-top:10px;border:1px solid #d8e1da;background:#fff;padding:14px}.result-score{display:flex;align-items:baseline;gap:6px;color:#b45c36}.result-score strong{font-size:36px}.result-score small{color:#74847c}.result-columns{display:grid;grid-template-columns:1fr 1fr;gap:14px}.result-card h4{margin:12px 0 7px;color:#34564b}.result-card p{margin:4px 0;color:#5f7169;line-height:1.55}.requirement-list{display:flex;flex-wrap:wrap;gap:6px;margin-top:10px}.requirement-list span{padding:4px 7px;font-size:11px;background:#f3f5f2}.requirement-list .status-matched{color:#22704d;background:#e9f5ed}.requirement-list .status-missing{color:#a14e3a;background:#faece8}.project-result pre{white-space:pre-wrap;background:#f5f7f4;padding:10px;line-height:1.65;font:inherit}.result-actions{margin-top:10px}.prediction-item{padding:9px 0;border-top:1px solid #edf0ed}.prediction-item strong{display:block}.prediction-item small{display:block;color:#b45c36;margin-top:3px}.prediction-item p{margin-top:5px}.muted{font-size:12px!important;color:#87938d!important}.typing{padding:0 22px 12px;color:#83918b;font-size:12px}.composer{border-top:1px solid #e1e7e2;padding:14px 18px}.composer-foot{display:flex;justify-content:space-between;align-items:center;gap:10px;margin-top:8px;color:#89968f;font-size:11px}@media(max-width:900px){.copilot-layout{grid-template-columns:1fr}.copilot-sidebar{display:grid;grid-template-columns:1fr 1fr}.history-panel{grid-column:1/-1}.chat-messages{max-height:none}}@media(max-width:600px){.copilot-header{display:block}.copilot-header .ant-btn{margin-top:14px}.copilot-sidebar{display:flex}.result-columns{grid-template-columns:1fr}.composer-foot{align-items:flex-end;flex-direction:column}.chat-shell{min-height:620px}}
</style>
