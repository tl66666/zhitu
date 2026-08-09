<template>
  <div class="interview-list-page">
    <!-- 顶部工具栏：左侧标题 + 描述，右侧胶囊主按钮 -->
    <div class="page-header">
      <div class="header-info">
        <h1 class="page-title">面试训练场</h1>
        <p class="page-desc">进入真实场景反复演练，让考官、流程和规则共同塑造训练</p>
      </div>
      <button class="new-btn" @click="goNew">
        <PlusOutlined :style="{ fontSize: '16px' }" />
        <span>选择训练场景</span>
      </button>
    </div>

    <a-spin :spinning="interviewStore.loading" tip="加载中...">
      <div class="spin-container" :class="{ 'is-loading': interviewStore.loading }">
        <!-- 空状态：居中图标 + 标题 + 描述 -->
        <div
          v-if="!interviewStore.loading && interviewStore.interviews.length === 0"
          class="empty-state"
        >
          <InboxOutlined class="empty-icon" />
          <h3 class="empty-title">暂无面试记录</h3>
          <p class="empty-desc">点击右上角开始新面试</p>
        </div>

        <!-- 面试会话表格卡片：白底 + 1px 边框 + 16px 圆角 + shadow-sm -->
        <div v-else-if="!interviewStore.loading" class="table-card">
          <div class="table-scroll">
            <table class="interview-table">
              <thead>
                <tr>
                  <th class="col-scene">场景</th>
                  <th class="col-target">目标</th>
                  <th class="col-difficulty">难度</th>
                  <th class="col-mode">模式</th>
                  <th class="col-progress">进度</th>
                  <th class="col-status">状态</th>
                  <th class="col-created">创建时间</th>
                  <th class="col-action">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="record in interviewStore.interviews"
                  :key="record.id"
                  class="table-row"
                >
                  <!-- 场景列：胶囊标签，不同场景不同配色 -->
                  <td>
                    <span class="capsule-tag" :style="sceneStyle(record.scene)">
                      {{ sceneLabel(record.scene) }}
                    </span>
                  </td>
                  <!-- 目标列：公司 + 职位 -->
                  <td>
                    <div class="target-company">{{ record.target_company || '未指定' }}</div>
                    <div class="target-position">{{ record.target_position || '-' }}</div>
                  </td>
                  <!-- 难度列：灰色胶囊 -->
                  <td>
                    <span class="capsule-tag difficulty-tag">
                      {{ difficultyLabel(record.difficulty) }}
                    </span>
                  </td>
                  <!-- 模式列：胶囊标签，不同模式不同配色 -->
                  <td>
                    <span class="capsule-tag" :style="modeStyle(record.mode)">
                      {{ modeLabel(record.mode) }}
                    </span>
                  </td>
                  <!-- 进度列：居中等宽字体 -->
                  <td class="cell-progress">
                    {{ record.current_question_no }} / {{ record.total_questions }}
                  </td>
                  <!-- 状态列：圆点 + 文字，进行中有脉动动画 -->
                  <td>
                    <div class="status-cell">
                      <span class="status-dot" :class="`status-${record.status}`"></span>
                      <span class="status-text">{{ statusLabel(record.status) }}</span>
                    </div>
                  </td>
                  <!-- 创建时间列 -->
                  <td class="cell-created">{{ formatDate(record.created_at) }}</td>
                  <!-- 操作列：文字按钮 + 删除 -->
                  <td>
                    <div class="action-cell">
                      <button class="action-btn" @click="enterRoom(record.id)">
                        {{ record.status === 'completed' ? '查看复盘' : record.status === 'cancelled' ? '查看详情' : record.status === 'reviewing' || record.status === 'report_failed' ? '查看进度' : record.status === 'preparing' ? '进入候场' : '继续面试' }}
                      </button>
                      <button
                        class="delete-btn"
                        title="删除该面试记录"
                        @click="handleDelete(record)"
                      >
                        <DeleteOutlined :style="{ fontSize: '14px' }" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { PlusOutlined, InboxOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { useInterviewStore } from '@/stores/interview'
import type {
  Interview,
  InterviewScene,
  InterviewMode,
  InterviewDifficulty,
  InterviewStatus,
} from '@/types/models'

const router = useRouter()
const interviewStore = useInterviewStore()

// 场景标签文字映射
const sceneLabel = (s: InterviewScene | string): string => {
  const map: Record<string, string> = {
    teaching: '模拟教室',
    tech: '技术面',
    behavior: '行为面',
    pressure: '压力面',
    hr: 'HR 面',
    group: '群面',
    corporate: '企业面',
    defense: '答辩面',
    client: '客户面',
    public: '公共面',
    medical: '医疗面',
    media: '媒体面',
    remote: '远程面',
    system: '系统面',
    aviation: '航空面',
  }
  return map[s] || s
}

// 场景胶囊标签样式：不同场景使用不同背景色与文字色（与设计稿一致）
const sceneStyle = (s: InterviewScene | string): { background: string; color: string } => {
  const map: Record<string, { bg: string; color: string }> = {
    tech: { bg: 'var(--brand-50)', color: 'var(--primary)' },
    hr: { bg: '#f3eefe', color: '#5856d6' },
    behavior: { bg: '#e6fbff', color: '#00b8d9' },
    pressure: { bg: 'var(--state-error-surface)', color: 'var(--state-error)' },
    group: { bg: '#fff1e6', color: '#ff9500' },
    teaching: { bg: 'var(--state-success-surface)', color: 'var(--state-success)' },
  }
  const style = map[s] || { bg: 'var(--background-200)', color: 'var(--text-600)' }
  return { background: style.bg, color: style.color }
}

// 难度标签文字映射
const difficultyLabel = (d: InterviewDifficulty | string): string => {
  const map: Record<string, string> = {
    junior: '初级',
    mid: '中级',
    senior: '高级',
    mixed: '混合',
  }
  return map[d] || d
}

// 模式标签文字映射
const modeLabel = (m: InterviewMode | string): string => {
  const map: Record<string, string> = {
    text: '文字',
    voice: '语音',
    hybrid: '混合',
  }
  return map[m] || m
}

// 模式胶囊标签样式：语音绿色、文字灰色、混合蓝色
const modeStyle = (m: InterviewMode | string): { background: string; color: string } => {
  const map: Record<string, { bg: string; color: string }> = {
    voice: { bg: 'var(--state-success-surface)', color: 'var(--state-success)' },
    text: { bg: 'var(--background-200)', color: 'var(--text-500)' },
    hybrid: { bg: 'var(--brand-50)', color: 'var(--primary)' },
  }
  const style = map[m] || { bg: 'var(--background-200)', color: 'var(--text-500)' }
  return { background: style.bg, color: style.color }
}

// 状态标签文字映射
const statusLabel = (s: InterviewStatus | string): string => {
  const map: Record<string, string> = {
    preparing: '准备中',
    starting: '正在开场',
    ongoing: '进行中',
    reviewing: '生成复盘中',
    completed: '已完成',
    report_failed: '复盘失败',
    cancelled: '已取消',
  }
  return map[s] || s
}

// 日期格式化为 YYYY-MM-DD HH:mm（与设计稿一致）
const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 跳转到新建面试页
const goNew = () => {
  router.push('/app/interviews/new')
}

// 进入面试房间
const enterRoom = (id: number) => {
  router.push(`/app/interviews/${id}`)
}

// 删除面试记录（二次确认）
const handleDelete = (record: Interview) => {
  Modal.confirm({
    title: '删除面试记录',
    content: `确定删除「${record.target_position || '未指定职位'}」的面试记录吗？相关的问答、评分和复盘报告将一并清除，且无法恢复。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      await interviewStore.removeInterview(record.id)
    },
  })
}

onMounted(() => {
  interviewStore.fetchList()
})
</script>

<style scoped>
.interview-list-page {
  width: 100%;
}

/* ===== 顶部工具栏 ===== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--foreground);
  margin: 0 0 6px 0;
  line-height: 1.2;
}

.page-desc {
  font-size: 15px;
  color: var(--muted-foreground);
  margin: 0;
  line-height: 1.4;
}

/* 胶囊主按钮 */
.new-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  background: var(--primary);
  color: var(--primary-foreground);
  border-radius: 980px;
  padding: 10px 20px;
  font-weight: 500;
  font-size: 14px;
  border: none;
  font-family: inherit;
  transition: opacity 0.15s ease, transform 0.15s ease, box-shadow 0.15s ease;
  white-space: nowrap;
}

.new-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.new-btn:active {
  transform: translateY(0);
  opacity: 0.85;
}

/* ===== 加载容器 ===== */
.spin-container {
  width: 100%;
}

.spin-container.is-loading {
  min-height: 240px;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 80px 0;
}

.empty-icon {
  font-size: 56px;
  color: var(--muted-foreground);
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--foreground);
  margin: 16px 0 6px 0;
}

.empty-desc {
  font-size: 14px;
  color: var(--muted-foreground);
  margin: 0;
}

/* ===== 表格卡片：白底 + 1px 边框 + 16px 圆角 + shadow-sm ===== */
.table-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

/* 横向滚动容器（窄屏适配） */
.table-scroll {
  overflow-x: auto;
}

.interview-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 980px;
}

/* 表头：background-100 底色，12px 大写字母，muted-foreground 颜色 */
.interview-table thead th {
  background: var(--background-100);
  font-size: 12px;
  font-weight: 600;
  color: var(--muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 12px 16px;
  text-align: left;
  white-space: nowrap;
}

.interview-table thead th.col-progress {
  text-align: center;
}

/* 列宽设定 */
.col-scene { width: 100px; }
.col-target { width: 200px; }
.col-difficulty { width: 90px; }
.col-mode { width: 90px; }
.col-progress { width: 100px; }
.col-status { width: 110px; }
.col-created { width: 160px; }
.col-action { width: 150px; }

/* 表体行：hover 时 background-100 底色 */
.table-row {
  border-bottom: 1px solid var(--background-200);
  transition: background 0.12s ease;
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: var(--background-100);
}

.interview-table tbody td {
  padding: 14px 16px;
  vertical-align: middle;
}


/* 目标列：公司 + 职位 */
.target-company {
  font-size: 14px;
  font-weight: 500;
  color: var(--foreground);
  line-height: 1.4;
}

.target-position {
  font-size: 12px;
  color: var(--muted-foreground);
  line-height: 1.4;
}

/* 进度列：居中等宽字体 */
.cell-progress {
  text-align: center;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--foreground);
  white-space: nowrap;
}

/* 创建时间列 */
.cell-created {
  font-size: 13px;
  color: var(--muted-foreground);
  white-space: nowrap;
}

/* 胶囊标签（场景、难度、模式） */
.capsule-tag {
  display: inline-block;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 980px;
  font-weight: 500;
  white-space: nowrap;
}

/* 难度标签统一灰色 */
.difficulty-tag {
  background: var(--background-200);
  color: var(--text-600);
}

/* 状态列：圆点 + 文字 */
.status-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* 进行中：蓝色 + 脉动动画 */
.status-dot.status-ongoing {
  background: var(--primary);
  box-shadow: 0 0 0 0 rgba(0, 122, 255, 0.5);
  animation: pulse-blue 1.8s infinite;
}

/* 已完成：绿色 */
.status-dot.status-completed {
  background: var(--state-success);
}

.status-dot.status-preparing { background: var(--background-500); }
.status-dot.status-starting,
.status-dot.status-reviewing { background: var(--primary); animation: pulse-blue 1.8s infinite; }
.status-dot.status-report_failed { background: var(--state-error); }

/* 已取消：灰色 */
.status-dot.status-cancelled {
  background: var(--background-500);
}

.status-text {
  font-size: 13px;
  color: var(--foreground);
}

/* 操作列：胶囊主按钮（全局统一） */
.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 32px;
  padding: 0 14px;
  border: none;
  border-radius: 9999px;
  background: var(--primary);
  color: var(--primary-foreground);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
  transition: background-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
}

.action-btn:hover {
  background: var(--brand-600);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.action-btn:active {
  transform: translateY(0);
  opacity: 0.9;
}

/* 操作列容器：主按钮 + 删除按钮 */
.action-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 删除按钮：幽灵图标按钮，hover 时变红 */
.delete-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: 1px solid var(--border);
  border-radius: 9999px;
  background: var(--card);
  color: var(--muted-foreground);
  cursor: pointer;
  font-family: inherit;
  transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  flex-shrink: 0;
}

.delete-btn:hover {
  color: var(--state-error);
  border-color: var(--state-error);
  background: var(--state-error-surface);
}

.delete-btn:active {
  opacity: 0.85;
}

/* 进行中状态脉动动画 */
@keyframes pulse-blue {
  0% {
    box-shadow: 0 0 0 0 rgba(0, 122, 255, 0.5);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(0, 122, 255, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(0, 122, 255, 0);
  }
}

/* 响应式：窄屏标题与按钮换行 */
@media (max-width: 640px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .page-title {
    font-size: 24px;
  }
}

/* 尊重 reduced-motion 偏好：关闭脉动动画 */
@media (prefers-reduced-motion: reduce) {
  .status-dot.status-ongoing {
    animation: none;
  }
}
</style>
