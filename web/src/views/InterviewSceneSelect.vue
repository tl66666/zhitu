<template>
  <div class="scene-page">
    <header class="scene-header">
      <h1>选择合适的面试场景</h1>
    </header>

    <main class="scene-layout">
      <section class="scene-library" aria-label="面试场景列表">
        <button
          v-for="scene in scenes"
          :key="scene.key"
          :class="['scene-card', { selected: selectedScene === scene.key }]"
          :style="{ backgroundPosition: scene.position }"
          @click="selectScene(scene.key)"
        >
          <span class="scene-shade"></span>
          <span class="scene-number">{{ scene.no }}</span>
          <span class="scene-copy"><strong>{{ scene.title }}</strong><small>{{ scene.subtitle }}</small></span>
          <span class="scene-status">已开放</span>
        </button>
      </section>

      <section class="classroom-panel">
        <div
          class="scenario-preview"
          :style="{ backgroundPosition: selectedSceneData.position }"
        >
          <div class="preview-shade"></div>
          <div class="preview-number">{{ selectedSceneData.no }}</div>
          <div class="preview-copy">
            <span>{{ selectedSceneData.subtitle }}</span>
            <strong>{{ selectedSceneData.title }}</strong>
            <p>{{ selectedSceneData.description }}</p>
          </div>
          <div class="room-caption">
            <span class="live-dot"></span>
            {{ form.direction }} · 场景化演练
          </div>
        </div>

        <div class="config-panel">
          <div class="config-heading">
            <div>
              <p>{{ selectedSceneData.title }}</p>
              <h2>配置本次面试</h2>
            </div>
            <span>约 {{ selectedSceneData.duration }} 分钟</span>
          </div>

          <div class="source-grid">
            <label>
              <span>本次简历</span>
              <a-select
                v-model:value="form.resumeId"
                placeholder="请选择本次面试使用的简历"
                :loading="loadingResumes"
                @change="onResumeChange"
              >
                <a-select-option v-for="resume in resumes" :key="resume.id" :value="resume.id">
                  {{ resume.name }}
                </a-select-option>
              </a-select>
            </label>
            <label>
              <span>简历版本</span>
              <a-select
                v-model:value="form.versionId"
                placeholder="使用当前版本"
                :loading="loadingVersions"
                :disabled="!form.resumeId"
              >
                <a-select-option :value="0">当前版本</a-select-option>
                <a-select-option v-for="version in versions" :key="version.id" :value="version.id">
                  {{ version.version_label }}
                </a-select-option>
              </a-select>
            </label>
          </div>

          <label class="jd-field">
            <span>目标 JD <b>必填</b></span>
            <a-textarea
              v-model:value="form.targetJd"
              :rows="5"
              placeholder="粘贴目标岗位 JD，面试官将根据 JD 与简历设计问题"
            />
          </label>

          <div class="field-grid">
            <label>
              <span>训练方向</span>
              <a-select v-model:value="form.direction">
                <a-select-option
                  v-for="direction in selectedSceneData.directions"
                  :key="direction"
                  :value="direction"
                >
                  {{ direction }}
                </a-select-option>
              </a-select>
            </label>
            <label>
              <span>目标岗位</span>
              <a-input v-model:value="form.targetPosition" :placeholder="selectedSceneData.positionPlaceholder" />
            </label>
            <label>
              <span>目标公司</span>
              <a-input v-model:value="form.targetCompany" placeholder="例如：目标企业（可选）" />
            </label>
            <label>
              <span>难度</span>
              <a-select v-model:value="form.difficulty">
                <a-select-option value="junior">基础</a-select-option>
                <a-select-option value="mid">标准</a-select-option>
                <a-select-option value="senior">进阶</a-select-option>
              </a-select>
            </label>
            <label>
              <span>考官风格</span>
              <a-select v-model:value="form.examinerStyle">
                <a-select-option value="标准规范">标准规范</a-select-option>
                <a-select-option value="温和引导">温和引导</a-select-option>
                <a-select-option value="连续追问">连续追问</a-select-option>
              </a-select>
            </label>
            <label>
              <span>面试题数</span>
              <a-select v-model:value="form.totalQuestions">
                <a-select-option :value="5">5 题 · 快速演练</a-select-option>
                <a-select-option :value="8">8 题 · 标准面试</a-select-option>
                <a-select-option :value="10">10 题 · 深度模拟</a-select-option>
              </a-select>
            </label>
          </div>

          <fieldset class="mode-field">
            <legend>面试形式</legend>
            <div class="mode-options">
              <button
                v-for="option in modeOptions"
                :key="option.value"
                type="button"
                :class="['mode-option', { selected: form.mode === option.value }]"
                :aria-pressed="form.mode === option.value"
                @click="form.mode = option.value"
              >
                <component :is="option.icon" />
                <span><strong>{{ option.label }}</strong><small>{{ option.description }}</small></span>
                <CheckCircleFilled v-if="form.mode === option.value" class="mode-check" />
              </button>
            </div>
          </fieldset>

          <label class="topic-field">
            <span>训练重点（可选）</span>
            <a-input v-model:value="form.topic" :placeholder="selectedSceneData.topicPlaceholder" />
          </label>

          <div class="process-strip">
            <span v-for="(step, index) in selectedSceneData.steps" :key="step">
              <b>0{{ index + 1 }}</b> {{ step }}
            </span>
          </div>

          <button class="enter-button" :disabled="creating || !canCreate" @click="createInterview">
            <LoadingOutlined v-if="creating" />
            <PlayCircleOutlined v-else />
            {{ creating ? '正在创建面试' : '创建并进入候场' }}
          </button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  AudioOutlined,
  CheckCircleFilled,
  KeyOutlined,
  LoadingOutlined,
  PlayCircleOutlined,
  SwapOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useInterviewStore } from '@/stores/interview'
import { listResumes, listVersions } from '@/api/resume'
import type { InterviewDifficulty, InterviewMode, InterviewScene, Resume, ResumeVersion } from '@/types/models'

interface SceneConfig {
  key: InterviewScene
  no: string
  title: string
  subtitle: string
  position: string
  description: string
  duration: number
  directions: string[]
  defaultPosition: string
  positionPlaceholder: string
  topicPlaceholder: string
  steps: string[]
}

const router = useRouter()
const interviewStore = useInterviewStore()
const selectedScene = ref<InterviewScene>('teaching')
const creating = ref(false)
const form = reactive<{
  direction: string
  targetPosition: string
  targetCompany: string
  difficulty: InterviewDifficulty
  examinerStyle: string
  topic: string
  targetJd: string
  resumeId: number | null
  versionId: number
  mode: InterviewMode
  totalQuestions: number
}>({
  direction: '教师资格证面试',
  targetPosition: '小学语文教师',
  targetCompany: '',
  difficulty: 'mid',
  examinerStyle: '标准规范',
  topic: '',
  targetJd: '',
  resumeId: null,
  versionId: 0,
  mode: 'hybrid',
  totalQuestions: 8,
})

const modeOptions: Array<{
  value: InterviewMode
  label: string
  description: string
  icon: typeof KeyOutlined
}> = [
  { value: 'text', label: '文字面试', description: '全程键盘作答，适合安静复盘', icon: KeyOutlined },
  { value: 'voice', label: '语音面试', description: '模拟真实通话，训练临场表达', icon: AudioOutlined },
  { value: 'hybrid', label: '混合面试', description: '语音与文字可按题灵活选择', icon: SwapOutlined },
]

const resumes = ref<Resume[]>([])
const versions = ref<ResumeVersion[]>([])
const loadingResumes = ref(false)
const loadingVersions = ref(false)
const canCreate = computed(() => Boolean(form.resumeId && form.targetJd.trim()))

const scenes: SceneConfig[] = [
  { key: 'teaching', no: '01', title: '模拟教室', subtitle: '教资面试 · 试讲 · 答辩', position: '0% 0%', description: '面对考官完成结构化问答、试讲与答辩。', duration: 15, directions: ['教师资格证面试', '教师招聘试讲', '说课训练'], defaultPosition: '小学语文教师', positionPlaceholder: '例如：小学语文教师', topicPlaceholder: '例如：《荷花》第二课时', steps: ['结构化问答', '模拟试讲', '考官答辩'] },
  { key: 'corporate', no: '02', title: '企业会议室', subtitle: 'HR · 业务面 · 校招', position: '33.333% 0%', description: '还原企业常规面试，训练经历表达和岗位匹配。', duration: 12, directions: ['校园招聘', '社会招聘', '业务复试'], defaultPosition: '产品经理', positionPlaceholder: '例如：产品经理', topicPlaceholder: '例如：项目经历、职业规划', steps: ['自我介绍', '经历深挖', '岗位匹配'] },
  { key: 'group', no: '03', title: '群面讨论室', subtitle: '无领导小组讨论', position: '66.666% 0%', description: '模拟多人讨论，关注观点、协作和总结能力。', duration: 18, directions: ['无领导小组讨论', '角色分工讨论', '方案共创'], defaultPosition: '管培生', positionPlaceholder: '例如：管培生', topicPlaceholder: '例如：资源排序、方案决策', steps: ['个人陈述', '自由讨论', '总结汇报'] },
  { key: 'defense', no: '04', title: '项目答辩室', subtitle: '论文 · 项目 · 方案', position: '100% 0%', description: '围绕一个项目进行陈述、追问和复盘。', duration: 15, directions: ['项目答辩', '论文答辩', '方案评审'], defaultPosition: '项目负责人', positionPlaceholder: '例如：项目负责人', topicPlaceholder: '例如：项目背景与核心成果', steps: ['项目陈述', '关键追问', '价值总结'] },
  { key: 'client', no: '05', title: '客户会议室', subtitle: '销售 · 售前 · 咨询', position: '0% 50%', description: '训练需求理解、方案说服与异议处理。', duration: 12, directions: ['销售面试', '售前顾问', '咨询顾问'], defaultPosition: '解决方案顾问', positionPlaceholder: '例如：解决方案顾问', topicPlaceholder: '例如：客户异议、需求澄清', steps: ['需求理解', '方案表达', '异议处理'] },
  { key: 'pressure', no: '06', title: '压力面试室', subtitle: '突发追问 · 临场反应', position: '33.333% 50%', description: '通过连续追问训练情绪稳定和快速组织。', duration: 10, directions: ['压力面试', '连续追问', '突发问题'], defaultPosition: '运营经理', positionPlaceholder: '例如：运营经理', topicPlaceholder: '例如：失败经历、冲突处理', steps: ['直接提问', '连续追问', '压力复盘'] },
  { key: 'public', no: '07', title: '结构化面试厅', subtitle: '公务员 · 事业单位', position: '66.666% 50%', description: '按结构化流程训练分析、组织和应变。', duration: 15, directions: ['公务员面试', '事业单位面试', '选调生面试'], defaultPosition: '综合管理岗', positionPlaceholder: '例如：综合管理岗', topicPlaceholder: '例如：社会现象、组织协调', steps: ['综合分析', '组织管理', '应急应变'] },
  { key: 'medical', no: '08', title: '医疗面试室', subtitle: '医护 · 规培 · 医患沟通', position: '100% 50%', description: '训练专业判断、服务意识与医患沟通。', duration: 12, directions: ['医护面试', '规培面试', '医患沟通'], defaultPosition: '临床医师', positionPlaceholder: '例如：临床医师', topicPlaceholder: '例如：医患沟通、职业伦理', steps: ['专业认知', '情景处置', '沟通表达'] },
  { key: 'media', no: '09', title: '媒体演播室', subtitle: '主持 · 公关 · 镜头表达', position: '0% 100%', description: '面向镜头训练信息组织和自然表达。', duration: 10, directions: ['主持面试', '公关面试', '镜头表达'], defaultPosition: '新媒体主持人', positionPlaceholder: '例如：新媒体主持人', topicPlaceholder: '例如：即兴评论、危机回应', steps: ['镜头介绍', '即兴表达', '现场回应'] },
  { key: 'remote', no: '10', title: '远程面试间', subtitle: '视频面试 · 英文面试', position: '33.333% 100%', description: '模拟视频连线，训练远程沟通和英文表达。', duration: 12, directions: ['视频面试', '英文面试', '异地复试'], defaultPosition: '海外业务专员', positionPlaceholder: '例如：海外业务专员', topicPlaceholder: '例如：英文自我介绍、远程协作', steps: ['连线开场', '核心问答', '远程沟通'] },
  { key: 'system', no: '11', title: '系统设计室', subtitle: '架构白板 · 技术评审', position: '66.666% 100%', description: '训练需求拆解、架构设计和方案权衡。', duration: 20, directions: ['系统设计', '架构评审', '技术方案'], defaultPosition: '后端工程师', positionPlaceholder: '例如：后端工程师', topicPlaceholder: '例如：高并发系统、消息队列', steps: ['需求澄清', '架构设计', '方案权衡'] },
  { key: 'aviation', no: '12', title: '航空面试厅', subtitle: '空乘 · 服务 · 仪态', position: '100% 100%', description: '训练服务意识、情景处置和职业表达。', duration: 12, directions: ['空乘面试', '地勤面试', '航空服务'], defaultPosition: '乘务员', positionPlaceholder: '例如：乘务员', topicPlaceholder: '例如：旅客冲突、服务礼仪', steps: ['形象介绍', '服务情景', '应急处置'] },
]

const selectedSceneData = computed(() => scenes.find((scene) => scene.key === selectedScene.value) || scenes[0])

const selectScene = (sceneKey: InterviewScene) => {
  const scene = scenes.find((item) => item.key === sceneKey)
  if (!scene) return
  selectedScene.value = scene.key
  form.direction = scene.directions[0]
  form.targetPosition = scene.defaultPosition
  form.topic = ''
}

const loadVersions = async (resumeId: number) => {
  loadingVersions.value = true
  try {
    const response = await listVersions(resumeId)
    versions.value = response.data.data || []
  } catch (error) {
    console.error('加载简历版本失败:', error)
    versions.value = []
  } finally {
    loadingVersions.value = false
  }
}

const onResumeChange = async (resumeId: number) => {
  form.versionId = 0
  versions.value = []
  const resume = resumes.value.find((item) => item.id === resumeId)
  if (resume) {
    form.targetJd = resume.target_jd || form.targetJd
    form.targetPosition = resume.target_position || form.targetPosition
  }
  await loadVersions(resumeId)
}

const loadResumes = async () => {
  loadingResumes.value = true
  try {
    const response = await listResumes()
    resumes.value = response.data.data || []
    if (resumes.value.length > 0 && !form.resumeId) {
      form.resumeId = resumes.value[0].id
      await onResumeChange(form.resumeId)
    }
  } catch (error) {
    console.error('加载简历列表失败:', error)
    message.error('加载简历列表失败')
  } finally {
    loadingResumes.value = false
  }
}

const createInterview = async () => {
  creating.value = true
  try {
    if (!form.resumeId) {
      message.warning('请先选择本次面试使用的简历')
      return
    }
    if (!form.targetJd.trim()) {
      message.warning('请先填写目标 JD')
      return
    }
    const topic = form.topic.trim() || '按场景默认重点进行'
    const scene = selectedSceneData.value
    const interview = await interviewStore.create({
      scene: scene.key,
      target_company: form.targetCompany.trim(),
      target_position: form.targetPosition.trim() || scene.defaultPosition,
      target_jd: form.targetJd.trim(),
      resume_id: form.resumeId,
      version_id: form.versionId || undefined,
      difficulty: form.difficulty,
      total_questions: form.totalQuestions,
      mode: form.mode,
      examiner_style: form.examinerStyle,
      training_focus: `${form.direction}；${topic}；流程：${scene.steps.join('、')}`,
    })
    if (!interview) {
      message.error('考场创建失败，请稍后重试')
      return
    }
    await router.push(`/app/interviews/${interview.id}`)
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void loadResumes()
})
</script>

<style scoped>
.scene-page{box-sizing:border-box;width:100%;max-width:100%;min-height:calc(100dvh - 64px);overflow-x:hidden;padding:20px clamp(20px,4vw,64px) 48px;background:#edf0ea;color:#16342d}
.scene-header{box-sizing:border-box;width:100%;max-width:1440px;margin:0 auto 18px}
h1{margin:0;font-family:"Songti SC","STSong",serif;font-size:clamp(38px,4.2vw,64px);font-weight:600;line-height:1.12;letter-spacing:-.04em}
.scene-layout{box-sizing:border-box;width:100%;max-width:1440px;margin:auto}
.scene-library{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px;margin-bottom:20px}
.scene-card{position:relative;isolation:isolate;min-height:164px;overflow:hidden;border:1px solid rgba(255,255,255,.4);background-image:url('/scenes/interview-scenes-atlas.jpg');background-size:400% auto;display:grid;grid-template-columns:1fr auto;align-content:end;gap:4px 12px;padding:18px;text-align:left;color:#fff;cursor:pointer;box-shadow:0 8px 24px rgba(24,48,41,.09);transition:transform .22s ease,box-shadow .22s ease}
.scene-shade{position:absolute;z-index:-1;inset:0;background:linear-gradient(180deg,transparent 28%,rgba(9,25,20,.82) 100%)}
.scene-card:hover,.scene-card.selected{transform:translateY(-3px);box-shadow:0 14px 30px rgba(24,48,41,.18)}.scene-card.selected{outline:3px solid #b45c36;outline-offset:-3px}
.scene-number{position:absolute;left:14px;top:12px;padding:5px 7px;background:rgba(12,35,28,.76);font-size:11px}.scene-status{align-self:end;padding:4px 7px;background:#b45c36;font-size:10px}.scene-copy{display:flex;flex-direction:column;gap:3px}.scene-copy strong{font-size:15px}.scene-copy small{color:#dce5df}
.classroom-panel{width:100%;min-width:0;background:#f8f8f3;border:1px solid #cbd2cc;display:grid;grid-template-columns:minmax(0,1.15fr) minmax(320px,.85fr);box-shadow:0 22px 60px rgba(28,56,47,.09)}
.scenario-preview{position:relative;isolation:isolate;min-height:520px;overflow:hidden;background-image:url('/scenes/interview-scenes-atlas.jpg');background-size:400% auto;color:#fff}
.preview-shade{position:absolute;z-index:-1;inset:0;background:linear-gradient(180deg,rgba(9,27,22,.08),rgba(9,27,22,.86))}
.preview-number{position:absolute;left:28px;top:26px;padding:7px 9px;background:rgba(16,45,37,.82);font-size:12px;letter-spacing:.12em}
.preview-copy{position:absolute;left:38px;right:38px;bottom:82px}.preview-copy span{color:#e8d2c7;font-size:12px;letter-spacing:.08em}.preview-copy strong{display:block;margin:9px 0 12px;font-family:"Songti SC",serif;font-size:clamp(38px,4vw,58px);line-height:1}.preview-copy p{max-width:520px;margin:0;color:#e1e9e4;font-size:15px;line-height:1.7}
.room-caption{position:absolute;left:38px;bottom:28px;padding:9px 12px;background:rgba(19,48,40,.9);color:#fff;font-size:12px}.live-dot{display:inline-block;width:7px;height:7px;margin-right:7px;border-radius:50%;background:#e46d43}
.config-panel{padding:34px;display:flex;flex-direction:column}.config-heading{display:flex;justify-content:space-between;gap:20px;border-bottom:1px solid #d8ddd8;padding-bottom:22px}.config-heading p{margin:0 0 4px;color:#b45c36;font-size:12px;font-weight:700}.config-heading h2{margin:0;font-family:"Songti SC",serif;font-size:28px}.config-heading>span{color:#6d7c76;font-size:12px}
.field-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px 14px;margin-top:24px}.field-grid label,.topic-field{display:flex;flex-direction:column;gap:7px}.field-grid label>span,.topic-field>span{font-size:12px;color:#5c6e67}.field-grid :deep(.ant-select){width:100%}.topic-field{margin-top:18px}
.mode-field{margin:20px 0 0;padding:0;border:0}.mode-field legend{padding:0;color:#5c6e67;font-size:12px}.mode-options{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px;margin-top:8px}.mode-option{position:relative;display:flex;align-items:flex-start;gap:10px;min-height:72px;padding:12px;border:1px solid #cbd2cc;background:#fff;color:#315048;text-align:left;cursor:pointer;font:inherit;transition:border-color .15s,background-color .15s,box-shadow .15s}.mode-option:hover{border-color:#b45c36}.mode-option.selected{border-color:#b45c36;background:#fff8f3;box-shadow:0 0 0 2px rgba(180,92,54,.12)}.mode-option>.anticon{margin-top:2px;color:#b45c36;font-size:18px}.mode-option span{display:flex;min-width:0;flex-direction:column;gap:4px}.mode-option strong{font-size:13px}.mode-option small{color:#708079;font-size:11px;line-height:1.45}.mode-check{position:absolute;right:9px;top:9px;color:#b45c36;font-size:14px}
.source-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px 14px;margin-top:24px}.source-grid label,.jd-field{display:flex;flex-direction:column;gap:7px}.source-grid label>span,.jd-field>span{font-size:12px;color:#5c6e67}.source-grid :deep(.ant-select){width:100%}.jd-field{margin-top:18px}.jd-field b{color:#b45c36;font-weight:600}
.process-strip{display:grid;grid-template-columns:repeat(3,1fr);gap:8px;margin:24px 0}.process-strip span{padding:12px 8px;border-top:1px solid #b9c4be;color:#52675f;font-size:11px}.process-strip b{display:block;margin-bottom:5px;color:#b45c36}
.enter-button{margin-top:auto;display:inline-flex;align-items:center;justify-content:center;gap:8px;height:44px;padding:0 24px;border:1px solid transparent;border-radius:9999px;background:var(--primary);color:var(--primary-foreground);font-size:14px;font-weight:600;font-family:inherit;cursor:pointer;transition:background-color .15s ease,box-shadow .15s ease,transform .15s ease}.enter-button:hover{background:var(--brand-600);box-shadow:var(--shadow-md);transform:translateY(-1px)}.enter-button:disabled{opacity:.6;cursor:not-allowed;transform:none;box-shadow:none}
@media(max-width:1300px){.scene-library{grid-template-columns:repeat(3,1fr)}.classroom-panel{grid-template-columns:minmax(0,1fr) 340px}.scenario-preview{min-height:500px}}
@media(max-width:760px){.scene-page{padding:22px 14px 36px}.scene-library{grid-template-columns:1fr 1fr}.scene-card{min-height:132px;padding:14px}.scene-status{display:none}.classroom-panel{grid-template-columns:1fr}.scenario-preview{min-height:360px}.preview-copy{left:24px;right:24px;bottom:72px}.room-caption{left:24px}.config-panel{padding:24px}.field-grid,.source-grid,.mode-options{grid-template-columns:1fr}h1{font-size:36px}}
</style>
