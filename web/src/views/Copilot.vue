<template>
  <div class="copilot-page">
    <a-drawer
      v-model:open="showHistory"
      title="对话记录"
      placement="right"
      :width="340"
      class="copilot-history-drawer"
    >
      <div class="history-list">
        <button
          v-for="session in copilotStore.sessions"
          :key="session.id"
          type="button"
          class="history-item"
          :class="{ active: session.id === copilotStore.activeSessionId }"
          @click="selectExistingSession(session.id); showHistory = false"
        >
          <span>{{ taskLabel(session.task) }}</span>
          <small>{{ formatTime(session.updated_at) }}</small>
        </button>
        <a-empty v-if="!copilotStore.sessions.length" :image="false" description="还没有对话记录" />
      </div>
      <template #footer>
        <a-button block @click="clearLocalSessions">清除全部记录</a-button>
      </template>
    </a-drawer>

    <div class="copilot-layout">
      <aside class="copilot-sidebar">
        <section class="context-panel">
          <div class="panel-title">工作上下文</div>
          <a-select v-model:value="selectedResumeId" allow-clear placeholder="选择简历，或清空后使用粘贴内容" :loading="loadingResumes" @change="handleResumeChange">
            <a-select-option v-for="resume in resumes" :key="resume.id" :value="resume.id">{{ resume.name }}</a-select-option>
          </a-select>
          <a-select v-model:value="selectedVersionId" placeholder="选择版本" :loading="loadingVersions" :disabled="!selectedResumeId" @change="loadVersionContent">
            <a-select-option v-for="version in versions" :key="version.id" :value="version.id">{{ version.version_label }}</a-select-option>
          </a-select>
          <a-textarea v-model:value="jd" :rows="5" placeholder="粘贴目标岗位 JD（匹配和面试预测需要）" />
          <a-textarea v-model:value="resumePaste" :rows="3" placeholder="没有现成简历时可直接粘贴" />
          <label class="file-input"><UploadOutlined /> 上传 TXT / Markdown 简历 <input type="file" accept=".txt,.md,text/plain,text/markdown" @change="handleResumeFile" /></label>
          <div class="context-chip" v-if="selectedResume">
            <FileTextOutlined /> {{ selectedResume.name }} · {{ selectedVersionLabel }}
          </div>
        </section>

        <section class="task-panel">
          <div class="panel-title">选择任务</div>
          <button v-for="item in taskItems" :key="item.key" type="button" class="task-button" :class="{ active: task === item.key }" @click="switchTask(item.key)">
            <component :is="item.icon" />
            <span>{{ item.title }}</span>
          </button>
        </section>
      </aside>

      <main class="copilot-main">
        <div v-if="task === 'project_optimize'" class="project-picker">
          <span>优化项目</span>
          <a-select v-if="currentContent.project.length" v-model:value="projectIndex" placeholder="选择一个项目">
            <a-select-option v-for="(project, index) in currentContent.project" :key="index" :value="index">{{ project.name || `项目 ${index + 1}` }}</a-select-option>
          </a-select>
          <div v-else class="project-empty"><InfoCircleOutlined /> 当前简历没有项目经历</div>
        </div>
        <div class="chat-shell">
          <div class="chat-title">
            <strong>{{ taskLabel(task) }}</strong>
            <div class="chat-title-actions">
              <a-button type="text" @click="showHistory = true"><HistoryOutlined /> 历史</a-button>
              <a-button v-if="activeSession" type="text" @click="newConversation">新对话</a-button>
            </div>
          </div>
          <div ref="messagesRef" class="chat-messages">
            <div v-if="!activeSession || !activeSession.messages.length" class="chat-welcome">
              <RobotOutlined />
              <h2>{{ welcomeTitle }}</h2>
              <div class="prompt-row">
                <button v-for="prompt in quickPrompts" :key="prompt" type="button" :disabled="!hasValidProject" @click="sendMessage(prompt)">{{ prompt }}</button>
              </div>
            </div>
            <article v-for="(msg, index) in activeSession?.messages || []" :key="`${msg.created_at}-${index}`" class="chat-message" :class="msg.role">
              <div class="message-avatar">{{ msg.role === 'assistant' ? 'AI' : '我' }}</div>
              <div class="message-body">
                <div class="message-content" v-html="renderMessageContent(msg.content)"></div>
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
            <div v-if="copilotStore.loading" class="typing"><span class="typing-dots"><i></i><i></i><i></i></span></div>
          </div>
          <div class="composer">
            <a-textarea v-model:value="input" :rows="3" :disabled="copilotStore.loading" :placeholder="composerPlaceholder" @keydown.enter.exact="handleEnter" />
            <div class="composer-foot"><a-button type="primary" :loading="copilotStore.loading" :disabled="!canSend" @click="sendMessage()"><SendOutlined />发送</a-button></div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { BulbOutlined, FileSearchOutlined, HistoryOutlined, InfoCircleOutlined, MessageOutlined, RobotOutlined, RocketOutlined, SendOutlined, FileTextOutlined, UploadOutlined } from '@ant-design/icons-vue'
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
const showHistory = ref(false)
const messagesRef = ref<HTMLElement | null>(null)

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

const renderMessageContent = (content: string): string =>
  escapeMessageHtml(content)
    .replace(/\*\*([\s\S]*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*\*/g, '')
    .replace(/\r?\n/g, '<br>')

const selectedResume = computed(() => resumes.value.find((item) => item.id === selectedResumeId.value))
const selectedVersionLabel = computed(() => versions.value.find((item) => item.id === selectedVersionId.value)?.version_label || '当前版本')
const activeSession = computed(() => copilotStore.activeSession)
const hasValidProject = computed(() => task.value !== 'project_optimize'
  || (currentContent.project.length > 0 && projectIndex.value >= 0 && projectIndex.value < currentContent.project.length))
const hasResumeContext = computed(() => Boolean(selectedResumeId.value || resumePaste.value.trim()))
const canSend = computed(() => Boolean(input.value.trim()) && hasResumeContext.value && hasValidProject.value)

const taskItems = [
  { key: 'jd_match' as const, title: '简历-JD 匹配', icon: FileSearchOutlined },
  { key: 'project_optimize' as const, title: '项目经历优化', icon: RocketOutlined },
  { key: 'interview_predict' as const, title: '岗位风险预测', icon: BulbOutlined },
  { key: 'career_chat' as const, title: '求职问答', icon: MessageOutlined },
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

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  })
}

const loadVersionContent = async () => {
  if (!selectedResumeId.value || !selectedVersionId.value) return
  const response = await getVersion(selectedResumeId.value, selectedVersionId.value)
  try {
    const parsed = JSON.parse(response.data.data?.content || '{}')
    currentContent.project = Array.isArray(parsed.project) ? parsed.project : []
    if (projectIndex.value < 0 || projectIndex.value >= currentContent.project.length) projectIndex.value = 0
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
  await copilotStore.createSession({ task: task.value, resumeId: selectedResumeId.value, versionId: selectedVersionId.value, jd: jd.value.trim(), draftContent: sessionDraftContent.value, projectIndex: task.value === 'project_optimize' && hasValidProject.value ? projectIndex.value : undefined })
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
  scrollToBottom()
}

const ensureSession = async () => {
  if (activeSession.value && activeSession.value.task === task.value && activeSession.value.resume_id === selectedResumeId.value && activeSession.value.version_id === selectedVersionId.value && activeSession.value.jd === jd.value.trim() && activeSession.value.draft_content === sessionDraftContent.value && (task.value !== 'project_optimize' || activeSession.value.project_index === projectIndex.value)) return activeSession.value
  await newConversation()
  return activeSession.value
}

const sendMessage = async (quick?: string) => {
  const content = (quick || input.value).trim()
  if (!content || (!selectedResumeId.value && !resumePaste.value.trim())) return
  if (!hasValidProject.value) {
    message.warning('当前简历没有可优化的项目经历，请先补充项目内容')
    return
  }
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
const composerPlaceholder = computed(() => task.value === 'career_chat' ? '例如：大三找实习，简历需要写几个项目？' : '告诉我你最关心的部分，支持继续追问和反驳…')
const quickPrompts = computed(() => task.value === 'jd_match' ? ['先给我一个匹配度结论', '哪些经历应该优先强化？'] : task.value === 'project_optimize' ? ['先分析这个项目的问题', '按实习校招风格改写'] : task.value === 'interview_predict' ? ['列出最高频的技术问题', '哪些简历细节最容易被追问？'] : ['Agent 项目简历怎么写？', '大三找实习要准备什么？'])

const clearLocalSessions = async () => {
  await copilotStore.clearAll()
  message.success('Copilot 对话已清除')
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

watch(() => activeSession.value?.messages.length, scrollToBottom)
watch(() => {
  const messages = activeSession.value?.messages || []
  return messages[messages.length - 1]?.content
}, scrollToBottom)
watch(() => copilotStore.loading, scrollToBottom)

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
.copilot-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px 20px;
  color: var(--foreground);
}

.copilot-layout {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(260px, 300px) minmax(0, 1fr);
  align-items: stretch;
  gap: 16px;
  overflow: hidden;
}

.copilot-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  overflow: auto;
}

.context-panel,
.task-panel,
.chat-shell,
.project-picker {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--card);
}

.context-panel,
.task-panel {
  padding: 14px;
}

.context-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.panel-title {
  margin-bottom: 2px;
  color: var(--foreground);
  font-size: 13px;
  font-weight: 700;
}

.context-panel :deep(.ant-select),
.context-panel :deep(.ant-input),
.context-panel :deep(.ant-input-affix-wrapper) {
  width: 100%;
}

.context-panel :deep(.ant-input),
.context-panel :deep(.ant-select-selector) {
  background: var(--background-50);
}

.context-panel :deep(.ant-input::placeholder) {
  color: var(--muted-foreground);
}

.file-input {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 7px;
  padding: 9px 10px;
  border: 1px dashed var(--border);
  border-radius: var(--radius-md);
  background: var(--background-100);
  color: var(--muted-foreground);
  font-size: 12px;
  cursor: pointer;
}

.file-input:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.file-input input {
  min-width: 0;
  max-width: 100%;
  font-size: 11px;
}

.context-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 10px;
  border: 1px solid var(--brand-100);
  border-radius: var(--radius-md);
  background: var(--brand-50);
  color: var(--primary);
  font-size: 12px;
}

.task-panel {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-button,
.history-item {
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--foreground);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.16s ease, background-color 0.16s ease, color 0.16s ease;
}

.task-button {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 10px;
  font-size: 13px;
  font-weight: 500;
}

.task-button:hover,
.history-item:hover {
  border-color: var(--border);
  background: var(--background-100);
}

.task-button svg {
  flex: 0 0 auto;
  color: var(--primary);
  font-size: 16px;
}

.task-button.active {
  border-color: var(--primary);
  background: var(--brand-50);
  color: var(--primary);
}

.history-item {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 9px 10px;
  font-size: 12px;
}

.history-item span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-item small {
  color: var(--muted-foreground);
  font-size: 11px;
  white-space: nowrap;
}

.history-item.active {
  border-color: var(--brand-100);
  background: var(--brand-50);
  color: var(--primary);
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.copilot-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.project-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 10px 14px;
}

.project-picker > span {
  color: var(--muted-foreground);
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.project-picker .ant-select {
  min-width: 260px;
}

.project-empty {
  display: flex;
  align-items: center;
  gap: 7px;
  color: #a14e3a;
  font-size: 12px;
  line-height: 1.5;
}

.chat-shell {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 52px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border);
  background: var(--background-100);
}

.chat-title strong {
  color: var(--foreground);
  font-size: 15px;
  font-weight: 650;
}

.chat-title-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.chat-title-actions :deep(.ant-btn) {
  height: 30px;
  padding: 0 10px;
  color: var(--muted-foreground);
  font-size: 12px;
}

.chat-title-actions :deep(.ant-btn:hover) {
  color: var(--primary);
}

.chat-messages {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 20px clamp(16px, 3vw, 32px);
  background: var(--card);
}

.chat-welcome {
  max-width: 560px;
  margin: 60px auto 0;
  padding: 28px 24px;
  text-align: center;
  color: var(--muted-foreground);
}

.chat-welcome > svg {
  color: var(--primary);
  font-size: 32px;
}

.chat-welcome h2 {
  margin: 12px 0 0;
  color: var(--foreground);
  font-size: 18px;
  font-weight: 650;
}

.prompt-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-top: 18px;
}

.prompt-row button {
  padding: 7px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--card);
  color: var(--foreground);
  font: inherit;
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.16s ease, color 0.16s ease, background-color 0.16s ease;
}

.prompt-row button:hover:not(:disabled) {
  border-color: var(--primary);
  background: var(--brand-50);
  color: var(--primary);
}

.prompt-row button:disabled {
  color: var(--muted-foreground);
  background: var(--background-100);
  cursor: not-allowed;
  opacity: 0.65;
}

.chat-message {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.chat-message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  display: grid;
  width: 30px;
  height: 30px;
  flex: 0 0 30px;
  place-items: center;
  border-radius: 6px;
  background: var(--primary);
  color: var(--primary-foreground);
  font-size: 11px;
  font-weight: 650;
}

.chat-message.user .message-avatar {
  background: var(--foreground);
}

.message-body {
  max-width: min(760px, 85%);
}

.chat-message.user .message-body {
  display: flex;
  justify-content: flex-end;
}

.message-content {
  padding: 11px 15px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--background-100);
  color: var(--foreground);
  line-height: 1.72;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.message-content :deep(strong) {
  color: var(--foreground);
  font-weight: 700;
}

.chat-message.user .message-content {
  border-color: var(--brand-100);
  background: var(--brand-50);
}

.result-card {
  margin-top: 10px;
  padding: 15px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--card);
}

.result-score {
  display: flex;
  align-items: baseline;
  gap: 6px;
  color: var(--primary);
}

.result-score strong {
  font-size: 36px;
  line-height: 1;
}

.result-score small {
  color: var(--muted-foreground);
}

.result-columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.result-card h4 {
  margin: 12px 0 7px;
  color: var(--foreground);
  font-size: 13px;
}

.result-card p {
  margin: 4px 0;
  color: var(--muted-foreground);
  line-height: 1.55;
}

.requirement-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.requirement-list span {
  padding: 4px 7px;
  border-radius: 4px;
  background: var(--background-100);
  color: var(--muted-foreground);
  font-size: 11px;
}

.requirement-list .status-matched {
  background: #e9f5ed;
  color: #22704d;
}

.requirement-list .status-missing {
  background: #faece8;
  color: #a14e3a;
}

.project-result pre {
  padding: 11px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--background-100);
  font: inherit;
  line-height: 1.65;
  white-space: pre-wrap;
}

.result-actions {
  margin-top: 10px;
}

.prediction-item {
  padding: 10px 0;
  border-top: 1px solid var(--border);
}

.prediction-item strong {
  display: block;
}

.prediction-item small {
  display: block;
  margin-top: 3px;
  color: var(--primary);
}

.prediction-item p {
  margin-top: 5px;
}

.muted {
  color: var(--muted-foreground) !important;
  font-size: 12px !important;
}

.typing {
  display: flex;
  padding: 0 0 16px;
}

.typing-dots {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--background-100);
}

.typing-dots i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--muted-foreground);
  animation: typing-bounce 1.2s infinite ease-in-out;
}

.typing-dots i:nth-child(2) { animation-delay: 0.15s; }
.typing-dots i:nth-child(3) { animation-delay: 0.3s; }

@keyframes typing-bounce {
  0%, 60%, 100% { opacity: 0.3; transform: translateY(0); }
  30% { opacity: 1; transform: translateY(-3px); }
}

.composer {
  padding: 12px 16px 14px;
  border-top: 1px solid var(--border);
  background: var(--card);
}

.composer :deep(.ant-input) {
  min-height: 72px;
  resize: vertical;
}

.composer-foot {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

@media (max-width: 1100px) {
  .copilot-layout {
    grid-template-columns: 260px minmax(0, 1fr);
  }
}

@media (max-width: 900px) {
  .copilot-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }

  .copilot-sidebar {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-items: start;
    overflow: visible;
  }
}

@media (max-width: 600px) {
  .copilot-page {
    padding: 10px 12px;
  }

  .copilot-sidebar {
    display: flex;
  }

  .project-picker {
    align-items: stretch;
    flex-direction: column;
  }

  .project-picker .ant-select {
    min-width: 0;
    width: 100%;
  }

  .result-columns {
    grid-template-columns: 1fr;
  }
}
</style>
