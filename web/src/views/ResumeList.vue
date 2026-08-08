<template>
  <div class="resume-list-page">
    <!-- 页面头部：左侧标题与描述，右侧创建按钮 -->
    <header class="page-header">
      <div class="page-title-group">
        <h1 class="page-title">简历管理</h1>
        <p class="page-desc">支持多份简历、版本管理、AI 生成与评分</p>
      </div>
      <button class="btn-create" type="button" @click="router.push('/app/resumes/new')">
        <PlusOutlined />
        <span>创建简历</span>
      </button>
    </header>

    <!-- 简历卡片网格 -->
    <a-spin :spinning="resumeStore.loading" tip="加载中...">
      <!-- 空状态 -->
      <div v-if="!resumeStore.loading && resumeStore.resumes.length === 0" class="empty-state">
        <FileTextOutlined class="empty-icon" />
        <p class="empty-title">暂无简历</p>
        <p class="empty-desc">点击右上角「创建简历」开始你的第一份简历</p>
      </div>

      <!-- 卡片网格 -->
      <section v-else class="resume-grid" aria-label="简历列表">
        <article
          v-for="resume in resumeStore.resumes"
          :key="resume.id"
          class="resume-card"
          @click="enterEditor(resume.id)"
        >
          <!-- 卡片头部：图标 + 标题 + 更多操作 -->
          <div class="card-head">
            <div class="card-title-group">
              <span class="card-icon">
                <FileTextOutlined />
              </span>
              <span class="card-title" :title="resume.name">{{ resume.name }}</span>
            </div>
            <a-dropdown :trigger="['click']" @click.stop>
              <button
                class="icon-btn"
                type="button"
                aria-label="更多操作"
                @click.stop
              >
                <MoreOutlined />
              </button>
              <template #overlay>
                <a-menu @click="(e) => handleMenuClick(e.key, resume)">
                  <a-menu-item key="rename">重命名</a-menu-item>
                  <a-menu-item key="delete" danger>删除简历</a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>

          <!-- 分隔线 -->
          <div class="card-divider"></div>

          <!-- 卡片正文：目标公司 / 目标职位 / 生成场景 -->
          <div class="card-body">
            <div class="card-row">
              <span class="card-label">目标公司</span>
              <span class="card-value">{{ resume.target_company || '未指定' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">目标职位</span>
              <span class="card-value">{{ resume.target_position || '未指定' }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">生成场景</span>
              <span class="tag" :class="sceneTagClass(resume.scene)">
                {{ sceneLabel(resume.scene) }}
              </span>
            </div>
          </div>

          <!-- 卡片底部：更新时间 -->
          <div class="card-foot">
            <span class="card-time">
              <ClockCircleOutlined />
              <span>{{ formatDate(resume.updated_at) }}</span>
            </span>
          </div>
        </article>
      </section>
    </a-spin>

    <a-modal
      v-model:open="renameModalVisible"
      title="重命名简历"
      :confirm-loading="renameLoading"
      ok-text="保存"
      cancel-text="取消"
      @ok="handleRename"
      @cancel="resetRename"
    >
      <a-input
        v-model:value="renameValue"
        placeholder="请输入简历名称"
        :maxlength="100"
        show-count
        @press-enter="handleRename"
      />
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  FileTextOutlined,
  MoreOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons-vue'
import { useResumeStore } from '@/stores/resume'
import type { Resume, ResumeScene } from '@/types/models'

const router = useRouter()
const resumeStore = useResumeStore()
const renameModalVisible = ref(false)
const renameLoading = ref(false)
const renamingResume = ref<Resume | null>(null)
const renameValue = ref('')

// 场景标签文本
const sceneLabel = (scene: ResumeScene | string): string => {
  const map: Record<string, string> = {
    manual: '手动编辑',
    jd: '基于 JD',
    scenario: '场景化',
  }
  return map[scene] || scene
}

// 场景标签样式：基于 JD 用主色，其余用中性色
const sceneTagClass = (scene: ResumeScene | string): string => {
  return scene === 'jd' ? 'tag-primary' : 'tag-neutral'
}

// 日期格式化：YYYY-MM-DD HH:mm
const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 进入编辑器
const enterEditor = (id: number) => {
  router.push(`/app/resumes/${id}`)
}

const openRenameModal = (resume: Resume) => {
  renamingResume.value = resume
  renameValue.value = resume.name
  renameModalVisible.value = true
}

const resetRename = () => {
  renameModalVisible.value = false
  renamingResume.value = null
  renameValue.value = ''
}

const handleRename = async () => {
  const resume = renamingResume.value
  const name = renameValue.value.trim()
  if (!resume || !name) {
    message.error('请输入简历名称')
    return
  }

  renameLoading.value = true
  try {
    const updated = await resumeStore.update(resume.id, { name })
    if (updated) {
      const item = resumeStore.resumes.find((candidate) => candidate.id === resume.id)
      if (item) item.name = name
      resetRename()
    }
  } finally {
    renameLoading.value = false
  }
}

// 菜单点击：重命名 / 删除
const handleMenuClick = (key: string, resume: Resume) => {
  if (key === 'rename') {
    openRenameModal(resume)
  } else if (key === 'delete') {
    Modal.confirm({
      title: '确认删除',
      content: `确定删除简历「${resume.name}」吗？所有版本将一并删除，且不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        await resumeStore.remove(resume.id)
      },
    })
  }
}

onMounted(() => {
  resumeStore.fetchList()
})
</script>

<style scoped>
/* ===== 页面容器 ===== */
.resume-list-page {
  width: 100%;
}

/* ===== 页面头部：标题 + 胶囊主按钮 ===== */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 24px;
}

.page-title-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--foreground);
}

.page-desc {
  margin: 0;
  font-size: 14px;
  color: var(--muted-foreground);
}

/* 胶囊形主按钮 */
.btn-create {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: 9999px;
  font-size: 14px;
  font-weight: 500;
  font-family: inherit;
  background: var(--primary);
  color: var(--primary-foreground);
  border: 1px solid transparent;
  box-shadow: var(--shadow-sm);
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 0.18s ease, box-shadow 0.18s ease,
    transform 0.18s ease;
}

.btn-create:hover {
  background: var(--brand-600);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.btn-create:active {
  transform: translateY(0);
}

.btn-create:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 16px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  color: var(--muted-foreground);
  margin-bottom: 16px;
}

.empty-title {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--foreground);
}

.empty-desc {
  margin: 0;
  font-size: 13px;
  color: var(--muted-foreground);
}

/* ===== 简历卡片网格 ===== */
.resume-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

/* 单张卡片：白底 + 1px 边框 + 16px 圆角 + 极淡阴影 */
.resume-card {
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: var(--shadow-sm);
  padding: 16px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.resume-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.resume-card:active {
  transform: translateY(0);
}

/* ===== 卡片头部 ===== */
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

/* 32px 圆角方块图标，brand-50 底色 */
.card-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--brand-50);
  color: var(--primary);
  flex-shrink: 0;
}

.card-icon :deep(svg) {
  width: 18px;
  height: 18px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--foreground);
  letter-spacing: -0.01em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 更多操作图标按钮 */
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: transparent;
  border: none;
  color: var(--muted-foreground);
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.icon-btn:hover {
  background: var(--background-200);
  color: var(--foreground);
}

.icon-btn :deep(svg) {
  width: 18px;
  height: 18px;
}

/* ===== 卡片分隔线 ===== */
.card-divider {
  height: 1px;
  background: var(--border);
  margin: 14px 0;
  border: none;
}

/* ===== 卡片正文 ===== */
.card-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.card-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.card-label {
  font-size: 13px;
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.card-value {
  font-size: 13px;
  color: var(--foreground);
  font-weight: 500;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ===== 场景标签 ===== */
.tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.6;
}

.tag-primary {
  background: var(--brand-50);
  color: var(--primary);
}

.tag-neutral {
  background: var(--background-200);
  color: var(--muted-foreground);
}

/* ===== 卡片底部 ===== */
.card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid var(--border);
}

.card-time {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  font-size: 12px;
  color: var(--muted-foreground);
}

.card-time :deep(svg) {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}

/* 文字按钮：编辑 */
.btn-text {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 4px;
  background: transparent;
  border: none;
  font-family: inherit;
  font-size: 14px;
  font-weight: 500;
  color: var(--primary);
  cursor: pointer;
  transition: color 0.18s ease;
}

.btn-text:hover {
  color: var(--brand-600);
}

.btn-text :deep(svg) {
  width: 16px;
  height: 16px;
}

/* ===== 响应式：≤640px 单列 ===== */
@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .btn-create {
    justify-content: center;
  }

  .resume-grid {
    grid-template-columns: 1fr;
  }
}

/* 动效降级：尊重用户偏好 */
@media (prefers-reduced-motion: reduce) {
  .resume-card,
  .btn-create,
  .icon-btn,
  .btn-text {
    transition: none;
  }

  .resume-card:hover {
    transform: none;
  }
}
</style>
