<template>
  <div class="editorial-home">
    <header class="site-header">
      <a class="wordmark" href="#top" @click.prevent="scrollTo('top')">
        <img src="/brand/trae-logo.png" alt="Trae" class="wordmark-logo" />
        <span>职途</span>
      </a>
      <nav class="site-nav" aria-label="首页导航">
        <a href="#method" @click.prevent="scrollTo('method')">求职方法</a>
        <a href="#tools" @click.prevent="scrollTo('tools')">产品工具</a>
        <a href="#steps" @click.prevent="scrollTo('steps')">开始使用</a>
      </nav>
      <div class="header-actions">
        <button class="text-button" type="button" @click="openAuthModal('login')">登录</button>
        <button class="solid-button compact" type="button" @click="openAuthModal('register')">
          建立求职档案
        </button>
      </div>
    </header>

    <main>
      <section id="top" class="hero-section">
        <div
          class="hero-background"
          :style="{ backgroundImage: `url(${heroImage})` }"
          role="img"
          aria-label="求职者在桌前修改纸质简历"
        ></div>
        <div class="hero-wash" aria-hidden="true"></div>
        <div class="hero-pattern" aria-hidden="true"></div>
        <div class="hero-copy">
          <p class="eyebrow">为认真求职的人，重新整理每一步</p>
          <h1>求职这条路，<br /><em>一步步走清楚</em></h1>
          <p class="hero-lead">
            从第一版简历到最后一次面试，把材料、练习和投递进度放在同一个工作台里。
          </p>
          <div class="hero-actions">
            <button class="solid-button" type="button" @click="handlePrimaryCta">
              免费开始 <ArrowRightOutlined />
            </button>
            <button class="line-button" type="button" @click="scrollTo('method')">看看怎么用</button>
          </div>
        </div>
        <div class="hero-caption">
          <span>01 / 求职工作台</span>
          <p>先把经历讲清楚，<br />再让工具帮你写得更好。</p>
        </div>
      </section>

      <section class="proof-strip" aria-label="产品能力概览">
        <p class="proof-intro">一套不打断思考的求职工作流</p>
        <div v-for="item in proofItems" :key="item.value" class="proof-item">
          <strong>{{ item.value }}</strong>
          <span>{{ item.label }}</span>
        </div>
      </section>

      <section id="method" class="editorial-section story-section">
        <div class="section-number">02</div>
        <div class="story-copy">
          <p class="eyebrow">重新梳理你的经历，让重点浮出来</p>
          <h2>一份简历，应该像你本人一样具体。</h2>
          <p class="section-lead">
            粘贴目标岗位，职途会把岗位要求与你的真实经历并排梳理。哪些值得展开，哪些可以删掉，你始终看得见。
          </p>
          <ul class="plain-list">
            <li><CheckOutlined /> 保留你的原始事实，不虚构经历</li>
            <li><CheckOutlined /> 根据不同岗位建立独立版本</li>
            <li><CheckOutlined /> A4 预览与措辞修改同步完成</li>
          </ul>
          <button class="inline-link" type="button" @click="handleFeature('/app/resumes')">
            打开简历实验室 <ArrowRightOutlined />
          </button>
        </div>
        <figure class="resume-story" aria-label="求职者根据真实经历修改纸质简历">
          <img
            class="resume-story-main"
            :src="resumePersonalImage"
            alt="一位求职者在家中书桌前逐条修改自己的纸质简历"
            loading="lazy"
          />
          <div class="resume-story-detail">
            <img
              :src="resumeDetailImage"
              alt="纸质简历上的手写圈画和经历批注"
              loading="lazy"
            />
            <span>真实材料，不是标准答案</span>
          </div>
          <figcaption>
            <span>一条经历，一处处说清楚</span>
            <strong>从你做过的事开始，而不是从漂亮话开始。</strong>
          </figcaption>
        </figure>
      </section>

      <section id="tools" class="photo-story">
        <figure>
          <img :src="interviewImage" alt="两位求职者在会议室进行模拟面试" loading="lazy" />
          <figcaption>像真实交流一样练习，让对方听见你的想法。</figcaption>
        </figure>
        <div class="photo-story-copy">
          <span class="section-number">03</span>
          <p class="eyebrow">面试训练场</p>
          <h2>练的，<br />是表达的节奏。</h2>
          <p>
            从自我介绍到连续追问，用文字或语音完成一场完整练习。每次卡顿、遗漏和有效表达，都能在结束后回看。
          </p>
          <div class="fact-row">
            <span>01</span><div><strong>从真实岗位出题</strong><small>围绕你的目标职位展开</small></div>
          </div>
          <div class="fact-row">
            <span>02</span><div><strong>追问会跟着回答走</strong><small>题目会顺着你的思路延展</small></div>
          </div>
          <div class="fact-row">
            <span>03</span><div><strong>保留整场复盘记录</strong><small>看见每一场的表达变化</small></div>
          </div>
          <button class="inline-link light" type="button" @click="handleFeature('/app/interviews')">
            开始一次练习 <ArrowRightOutlined />
          </button>
        </div>
      </section>

      <section class="editorial-section tracker-section">
        <div class="tracker-copy">
          <span class="section-number">04</span>
          <p class="eyebrow">投递看板</p>
          <h2>少一点焦虑，<br />多一点进度感。</h2>
          <p class="section-lead">
            公司、岗位、沟通和面试轮次都留在同一个地方。下一步该跟进谁，一眼就能看到。
          </p>
          <div class="tracker-flow" aria-label="投递流程">
            <span>感兴趣</span><i></i><span>已投递</span><i></i><span>面试中</span><i></i><span>Offer</span>
          </div>
          <button class="inline-link" type="button" @click="handleFeature('/app/deliveries')">
            整理我的投递 <ArrowRightOutlined />
          </button>
        </div>
        <figure class="workspace-figure">
          <img :src="workspaceImage" alt="桌面上的简历、笔记本与投递记录" loading="lazy" />
          <figcaption><strong>一周回顾</strong><span>哪些岗位有回应，哪些策略需要调整，一眼看清。</span></figcaption>
        </figure>
      </section>

      <section class="voices-section">
        <div class="voices-heading">
          <p class="eyebrow">真实的求职片段</p>
          <h2>每个人，都有自己的节奏。</h2>
        </div>
        <div class="voices-grid">
          <article v-for="t in testimonials" :key="t.name" class="voice-card">
            <div class="voice-stars" aria-label="五星评价">
              <StarFilled v-for="n in 5" :key="n" />
            </div>
            <blockquote>“{{ t.quote }}”</blockquote>
            <div class="voice-author">
              <img :src="t.photo" :alt="t.name" loading="lazy" />
              <div><strong>{{ t.name }}</strong><span>{{ t.role }}</span></div>
            </div>
          </article>
        </div>
      </section>

      <section id="steps" class="steps-section">
        <div class="steps-heading">
          <p class="eyebrow">从今天开始</p>
          <h2>先完成一件小事。</h2>
        </div>
        <ol class="steps-list">
          <li v-for="(step, index) in steps" :key="step.title">
            <span>0{{ index + 1 }}</span>
            <div><strong>{{ step.title }}</strong><p>{{ step.desc }}</p></div>
            <ArrowRightOutlined />
          </li>
        </ol>
      </section>

      <section class="final-cta">
        <p>你的下一份工作，<br />从一份更清晰的计划开始。</p>
        <button class="solid-button inverse" type="button" @click="handlePrimaryCta">
          建立求职档案 <ArrowRightOutlined />
        </button>
      </section>
    </main>

    <footer class="site-footer">
      <a class="wordmark footer-wordmark" href="#top" @click.prevent="scrollTo('top')">
        <img src="/brand/trae-logo.png" alt="Trae" class="wordmark-logo" /><span>职途</span>
      </a>
      <p>求职这件事，我们一起努力。</p>
      <span>© 2026 职途</span>
    </footer>

    <a-modal v-model:open="authModalVisible" :footer="null" :width="430" :destroy-on-close="false">
      <div class="auth-modal-body">
        <p class="eyebrow">{{ activeTab === 'login' ? '继续你的求职计划' : '建立你的求职档案' }}</p>
        <h3>{{ activeTab === 'login' ? '欢迎回来' : '从今天开始' }}</h3>
        <div class="auth-tabs">
          <button type="button" :class="{ active: activeTab === 'login' }" @click="activeTab = 'login'">登录</button>
          <button type="button" :class="{ active: activeTab === 'register' }" @click="activeTab = 'register'">注册</button>
        </div>
        <a-form v-if="activeTab === 'login'" :model="loginForm" :rules="loginRules" layout="vertical" @finish="handleLogin">
          <a-form-item label="邮箱" name="email"><a-input v-model:value="loginForm.email" size="large" placeholder="name@example.com"><template #prefix><MailOutlined /></template></a-input></a-form-item>
          <a-form-item label="密码" name="password"><a-input-password v-model:value="loginForm.password" size="large" placeholder="不少于 6 位"><template #prefix><LockOutlined /></template></a-input-password></a-form-item>
          <button class="solid-button auth-submit" type="submit" :disabled="loading">{{ loading ? '登录中…' : '登录' }}</button>
        </a-form>
        <a-form v-else :model="registerForm" :rules="registerRules" layout="vertical" @finish="handleRegister">
          <a-form-item label="邮箱" name="email"><a-input v-model:value="registerForm.email" size="large" placeholder="name@example.com"><template #prefix><MailOutlined /></template></a-input></a-form-item>
          <a-form-item label="昵称" name="nickname"><a-input v-model:value="registerForm.nickname" size="large" placeholder="怎么称呼你"><template #prefix><UserOutlined /></template></a-input></a-form-item>
          <a-form-item label="密码" name="password"><a-input-password v-model:value="registerForm.password" size="large" placeholder="不少于 6 位"><template #prefix><LockOutlined /></template></a-input-password></a-form-item>
          <a-form-item label="确认密码" name="confirmPassword"><a-input-password v-model:value="registerForm.confirmPassword" size="large" placeholder="再次输入密码"><template #prefix><LockOutlined /></template></a-input-password></a-form-item>
          <button class="solid-button auth-submit" type="submit" :disabled="loading">{{ loading ? '注册中…' : '创建账号' }}</button>
        </a-form>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { ArrowRightOutlined, CheckOutlined, MailOutlined, LockOutlined, UserOutlined, StarFilled } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'
import { useAuthStore } from '@/stores/auth'
import heroImage from '@/assets/home-editorial/hero-career-desk.webp'
import interviewImage from '@/assets/home-editorial/interview-practice.webp'
import workspaceImage from '@/assets/home-editorial/application-workspace.webp'
import resumePersonalImage from '@/assets/home-editorial/resume-personal-edit.webp'
import resumeDetailImage from '@/assets/home-editorial/resume-markup-detail.webp'

const router = useRouter()
const authStore = useAuthStore()
const authModalVisible = ref(false)
const activeTab = ref<'login' | 'register'>('login')
const loading = ref(false)
const loginForm = reactive({ email: '', password: '' })
const registerForm = reactive({ email: '', nickname: '', password: '', confirmPassword: '' })

const validateConfirmPassword = async (_rule: Rule, value: string) => {
  if (value !== registerForm.password) return Promise.reject('两次输入的密码不一致')
  return Promise.resolve()
}
const loginRules: Record<string, Rule[]> = {
  email: [{ required: true, message: '请输入邮箱' }, { type: 'email', message: '请输入有效的邮箱地址' }],
  password: [{ required: true, message: '请输入密码' }, { min: 6, message: '密码长度不能少于 6 位' }],
}
const registerRules: Record<string, Rule[]> = {
  email: loginRules.email,
  nickname: [{ required: true, message: '请输入昵称' }, { min: 2, max: 20, message: '昵称长度为 2-20 个字符' }],
  password: loginRules.password,
  confirmPassword: [{ required: true, message: '请再次输入密码' }, { validator: validateConfirmPassword }],
}

const proofItems = [
  { value: '3', label: '个关键求职阶段' },
  { value: '1', label: '个连续工作台' },
  { value: '∞', label: '次表达练习' },
]
const steps = [
  { title: '选一个目标岗位', desc: '先对准目标再下笔，从真实岗位要求开始。' },
  { title: '整理一版真实经历', desc: '保留事实，把重点说具体，少一些夸张包装。' },
  { title: '完成一次面试练习', desc: '用自己的语言回答，再根据记录慢慢调整。' },
]
// 保留原首页已有的用户头像与证言素材
const testimonials = [
  {
    quote: '从粘贴 JD 到生成简历只用了 30 秒，措辞比我手写的专业多了。',
    name: '张同学',
    role: '前端工程师 · 3 年经验',
    avatar: 'Z',
    color: '#007aff',
    photo:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=photorealistic%20headshot%20portrait%20of%20young%20Chinese%20woman%2C%20mid%20twenties%2C%20software%20engineer%2C%20soft%20natural%20light%2C%20blurred%20office%20background%2C%20subtle%20smile%2C%20smart%20casual%20blouse%2C%204k&image_size=square_hd',
  },
  {
    quote: '模拟面试的追问很真实，第一次面试被问懵的题，第二次在这里练熟了，正式面试拿下 Offer。',
    name: '李同学',
    role: '后端开发 · 应届生',
    avatar: 'L',
    color: '#34c759',
    photo:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=photorealistic%20headshot%20portrait%20of%20young%20Chinese%20man%2C%20early%20twenties%2C%20graduate%20developer%2C%20soft%20studio%20light%2C%20neutral%20background%2C%20friendly%20smile%2C%20collared%20shirt%2C%204k&image_size=square_hd',
  },
  {
    quote: '投递看板让我看清自己的漏斗，原来拒了 20 家只有 3 家面试，调整策略后面试率翻倍。',
    name: '王同学',
    role: '产品经理 · 转行求职',
    avatar: 'W',
    color: '#af52de',
    photo:
      'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=photorealistic%20headshot%20portrait%20of%20Chinese%20man%20around%20thirty%2C%20product%20manager%2C%20soft%20natural%20light%2C%20blurred%20office%20background%2C%20confident%20expression%2C%20business%20casual%2C%204k&image_size=square_hd',
  },
]

const openAuthModal = (tab: 'login' | 'register') => {
  activeTab.value = tab
  authModalVisible.value = true
}
const handlePrimaryCta = () => authStore.isAuthenticated ? router.push('/app') : openAuthModal('register')
const handleFeature = (path: string) => authStore.isAuthenticated ? router.push(path) : openAuthModal('login')
const scrollTo = (id: string) => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })

const handleLogin = async () => {
  loading.value = true
  try {
    if (await authStore.login(loginForm)) {
      authModalVisible.value = false
      router.push('/app')
    }
  } finally { loading.value = false }
}
const handleRegister = async () => {
  loading.value = true
  try {
    if (await authStore.register({ email: registerForm.email, nickname: registerForm.nickname, password: registerForm.password })) {
      loginForm.email = registerForm.email
      loginForm.password = ''
      activeTab.value = 'login'
      message.info('请使用注册的账号登录')
    }
  } finally { loading.value = false }
}
</script>

<style scoped>
.editorial-home {
  --ink: #151a23;
  --muted: #626a77;
  --paper: #f4f5f2;
  --blue: #1757d2;
  --line: #d8dce1;
  min-height: 100vh;
  color: var(--ink);
  background: var(--paper);
  font-family: "SF Pro Display", "PingFang SC", "Noto Sans CJK SC", sans-serif;
}
button, a { -webkit-tap-highlight-color: transparent; }
button { font: inherit; }
.site-header {
  position: sticky;
  top: 0;
  z-index: 100;
  width: 100%;
  height: 84px;
  margin: 0 auto;
  /* 不再以 1440px 容器为界无限扩大留白，任何屏幕都贴近浏览器边框 */
  padding-inline: clamp(20px, 3vw, 48px);
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  border-bottom: 0;
  box-shadow: 0 1px 0 rgba(21, 26, 35, .04);
  backdrop-filter: saturate(180%) blur(14px);
  -webkit-backdrop-filter: saturate(180%) blur(14px);
  background: rgba(244, 245, 242, 0.72);
}
.wordmark { display: inline-flex; align-items: center; gap: 10px; width: fit-content; color: var(--ink); font-size: 20px; font-weight: 750; text-decoration: none; letter-spacing: -.03em; }
.wordmark-logo { width: 34px; height: 34px; object-fit: contain; display: block; flex-shrink: 0; }
.site-nav { display: flex; gap: 36px; }
.site-nav a { color: #343b46; font-size: 14px; text-decoration: none; }
.site-nav a:hover { color: var(--blue); }
.header-actions { display: flex; justify-content: flex-end; align-items: center; gap: 14px; }
.text-button, .line-button, .inline-link { border: 0; background: none; color: var(--ink); cursor: pointer; }
.text-button { padding: 10px 8px; }
.solid-button { min-height: 52px; padding: 0 24px; border: 1px solid var(--blue); background: var(--blue); color: white; display: inline-flex; align-items: center; justify-content: center; gap: 12px; cursor: pointer; font-weight: 650; transition: transform .2s ease, background .2s ease; white-space: nowrap; }
.solid-button:hover { background: #0d47bb; }
.solid-button:active, .line-button:active { transform: translateY(1px); }
.solid-button.compact { min-height: 42px; padding: 0 18px; font-size: 14px; }
.hero-section { position: relative; isolation: isolate; width: 100%; min-height: calc(100vh - 84px); margin: 0 auto; display: flex; align-items: center; overflow: hidden; padding-block: 74px 112px; padding-inline: clamp(20px, 3vw, 48px); background: #f2f3f0; }
.hero-background { position: absolute; z-index: -3; inset: 0; background-position: 58% center; background-size: cover; filter: grayscale(.15) saturate(.55) contrast(.9); opacity: .24; transform: scale(1.01); }
.hero-wash { position: absolute; z-index: -2; inset: 0; background: linear-gradient(90deg, rgba(246,247,244,.99) 0%, rgba(246,247,244,.93) 42%, rgba(246,247,244,.48) 66%, rgba(246,247,244,.18) 100%); }
.hero-pattern { position: absolute; z-index: -1; left: 0; right: 0; bottom: 0; height: 30%; opacity: .36; background-color: rgba(255,255,255,.68); background-image: repeating-linear-gradient(12deg, transparent 0 14px, rgba(84,95,110,.14) 14px 15px, transparent 15px 28px); clip-path: polygon(0 52%, 62% 14%, 100% 0, 100% 100%, 0 100%); }
.hero-copy { position: relative; z-index: 1; max-width: 680px; }
.eyebrow { margin: 0 0 24px; color: var(--blue); font-size: 12px; font-weight: 750; letter-spacing: .14em; text-transform: uppercase; }
h1 { margin: 0; font-size: clamp(48px, 5.4vw, 82px); line-height: 1.08; letter-spacing: -.065em; font-weight: 730; }
h1 em { color: var(--blue); font-style: normal; }
.hero-lead { max-width: 520px; margin: 32px 0 0; color: var(--muted); font-size: 18px; line-height: 1.85; }
.hero-actions { display: flex; gap: 14px; margin-top: 36px; }
.line-button { min-height: 52px; padding: 0 22px; border: 1px solid transparent; background: rgba(21,26,35,0.04); font-weight: 650; transition: background 0.2s ease; }
.line-button:hover { background: rgba(21,26,35,0.08); }
.hero-caption { position: absolute; right: clamp(20px, 3vw, 48px); bottom: 38px; width: min(280px, 28%); padding-top: 14px; color: #313844; }
.hero-caption::before { content: ''; position: absolute; left: 0; right: 30%; top: 0; height: 2px; background: linear-gradient(90deg, rgba(21,26,35,.72), rgba(21,26,35,0)); border-radius: 2px; }
.hero-caption span { color: #697280; font-size: 11px; letter-spacing: .08em; }
.hero-caption p { margin: 8px 0 0; font-size: 13px; line-height: 1.55; }
.proof-strip { width: 100%; margin: 0 auto; padding-inline: clamp(20px, 3vw, 48px); min-height: 150px; border-top: 0; border-bottom: 0; display: grid; grid-template-columns: 1.5fr repeat(3, 1fr); align-items: stretch; }
.proof-intro { display: flex; align-items: center; margin: 0; max-width: 240px; font-size: 18px; line-height: 1.5; font-weight: 650; }
.proof-item { padding: 34px 34px 34px 48px; border-left: 0; display: flex; flex-direction: column; justify-content: center; }
.proof-item strong { font-size: 34px; letter-spacing: -.04em; }
.proof-item span { margin-top: 6px; color: var(--muted); font-size: 13px; }
.editorial-section { width: 100%; margin: 0 auto; padding: 200px clamp(20px, 3vw, 48px); }
.story-section { display: grid; grid-template-columns: 70px .9fr 1.1fr; gap: clamp(28px, 5vw, 78px); align-items: center; }
.section-number { align-self: start; color: #9ca3ad; font-size: 13px; letter-spacing: .1em; }
h2 { margin: 0; font-size: clamp(40px, 4vw, 64px); line-height: 1.12; letter-spacing: -.055em; font-weight: 710; }
.section-lead, .photo-story-copy > p:not(.eyebrow) { margin: 30px 0; color: var(--muted); font-size: 17px; line-height: 1.9; }
.plain-list { list-style: none; padding: 0; margin: 30px 0; border-top: 0; }
.plain-list li { display: flex; gap: 12px; align-items: center; padding: 14px 0; border-bottom: 0; color: #343b46; }
.plain-list li + li { margin-top: 4px; }
.plain-list :deep(svg) { color: var(--blue); }
.inline-link { display: inline-flex; align-items: center; gap: 12px; padding: 10px 0; border-bottom: 0; font-weight: 700; transition: gap 0.2s ease, color 0.2s ease; }
.inline-link:hover { color: var(--blue); gap: 16px; }
.resume-story { position: relative; min-width: 0; min-height: 650px; margin: 0; overflow: hidden; background: #dfe4ea; }
.resume-story-main { width: 100%; height: 650px; display: block; object-fit: cover; object-position: 45% center; filter: saturate(.82) contrast(.97); }
.resume-story::after { content: ''; position: absolute; inset: 0; pointer-events: none; background: linear-gradient(180deg, transparent 58%, rgba(17,24,34,.52) 100%); }
.resume-story-detail { position: absolute; z-index: 2; top: 30px; right: 30px; width: min(240px, 34%); padding: 8px 8px 12px; background: rgba(249,249,246,.94); box-shadow: 0 18px 42px rgba(26,33,43,.2); transform: rotate(2deg); }
.resume-story-detail img { width: 100%; aspect-ratio: 4 / 3; display: block; object-fit: cover; }
.resume-story-detail span { display: block; padding: 10px 5px 0; color: #58606b; font-size: 11px; letter-spacing: .05em; }
.resume-story figcaption { position: absolute; z-index: 2; left: 30px; right: 30px; bottom: 26px; display: flex; align-items: end; justify-content: space-between; gap: 28px; color: white; }
.resume-story figcaption span { flex: 0 0 auto; font-size: 11px; letter-spacing: .12em; }
.resume-story figcaption strong { max-width: 370px; font-size: 17px; line-height: 1.55; text-align: right; font-weight: 600; }
.photo-story { display: grid; grid-template-columns: 1.2fr .8fr; min-height: 920px; background: #171d27; color: white; }
.photo-story figure { min-width: 0; margin: 0; position: relative; overflow: hidden; }
.photo-story figure img { width: 100%; height: 100%; min-height: 920px; display: block; object-fit: cover; object-position: center; }
.photo-story figure figcaption { position: absolute; left: 34px; bottom: 30px; padding: 12px 16px; background: rgba(255,255,255,.9); color: var(--ink); font-size: 13px; }
.photo-story-copy { display: flex; flex-direction: column; justify-content: center; padding: 80px clamp(44px, 6vw, 100px); }
.photo-story-copy .section-number { margin-bottom: 70px; color: #748096; }
.photo-story-copy .eyebrow { color: #87aaf2; }
.photo-story-copy > p:not(.eyebrow) { color: #aeb7c5; }
.fact-row { display: grid; grid-template-columns: 36px 1fr; gap: 16px; padding: 18px 0; border-top: 0; }
.fact-row + .fact-row { margin-top: 4px; }
.fact-row > span { color: #6d788a; font-size: 12px; }.fact-row div { display: flex; flex-direction: column; gap: 5px; }.fact-row small { color: #909aa9; }
.inline-link.light { align-self: flex-start; margin-top: 24px; color: white; border-color: #7b8492; }
.tracker-section { display: grid; grid-template-columns: .85fr 1.15fr; gap: clamp(40px, 7vw, 110px); align-items: center; }
.tracker-copy .section-number { display: block; margin-bottom: 70px; }
.tracker-flow { display: grid; grid-template-columns: auto 1fr auto 1fr auto 1fr auto; align-items: center; margin: 44px 0; font-size: 12px; }
.tracker-flow i { height: 2px; margin: 0 14px; background: linear-gradient(90deg, rgba(175,181,189,0), rgba(175,181,189,1) 50%, rgba(175,181,189,0)); border-radius: 2px; }
.workspace-figure { position: relative; margin: 0; height: 760px; overflow: hidden; }
.workspace-figure img { width: 100%; height: 100%; object-fit: cover; display: block; }
.workspace-figure figcaption { position: absolute; right: 0; bottom: 0; width: min(330px, 70%); padding: 24px 28px; background: white; display: flex; flex-direction: column; gap: 6px; }
.workspace-figure figcaption span { color: var(--muted); font-size: 13px; line-height: 1.6; }
.voices-section { padding: 180px max(32px, calc((100vw - 1220px) / 2)); background: #171d27; color: white; }
.voices-heading { display: flex; justify-content: space-between; align-items: end; gap: 50px; margin-bottom: 72px; }
.voices-heading .eyebrow { flex: 0 0 auto; margin-bottom: 10px; color: #87aaf2; }
.voices-heading h2 { max-width: 680px; }
.voices-grid { display: grid; grid-template-columns: repeat(3, 1fr); border-top: 0; gap: 24px; }
.voice-card { min-height: 440px; padding: 52px 44px 48px; border-right: 0; display: flex; flex-direction: column; background: rgba(255,255,255,0.04); border-radius: 18px; backdrop-filter: blur(8px); }
.voice-card:last-child { border-right: 0; }
.voice-stars { display: flex; gap: 4px; color: #87aaf2; font-size: 12px; }
.voice-card blockquote { flex: 1; margin: 32px 0 48px; color: #e4e8ee; font-size: 18px; line-height: 1.8; }
.voice-author { display: flex; align-items: center; gap: 15px; }
.voice-author img { width: 52px; height: 52px; border-radius: 50%; object-fit: cover; background: #303846; }
.voice-author div { display: flex; flex-direction: column; gap: 5px; }
.voice-author span { color: #929dac; font-size: 12px; }
.steps-section { padding: 130px max(32px, calc((100vw - 1220px) / 2)); background: #e5e8ec; display: grid; grid-template-columns: .7fr 1.3fr; gap: 100px; }
.steps-heading { position: sticky; top: 40px; align-self: start; }
.steps-list { list-style: none; margin: 0; padding: 0; border-top: 0; }
.steps-list li { min-height: 150px; display: grid; grid-template-columns: 54px 1fr auto; gap: 24px; align-items: center; border-bottom: 0; padding: 24px 0; }
.steps-list li + li { margin-top: 8px; }
.steps-list li > span { color: var(--blue); font-size: 12px; }.steps-list strong { font-size: 21px; }.steps-list p { margin: 8px 0 0; color: var(--muted); line-height: 1.6; }
.final-cta { min-height: 450px; padding: 80px max(32px, calc((100vw - 1220px) / 2)); background: var(--blue); color: white; display: flex; justify-content: space-between; align-items: center; gap: 60px; }
.final-cta p { margin: 0; font-size: clamp(42px, 5vw, 72px); line-height: 1.15; letter-spacing: -.055em; font-weight: 700; }
.solid-button.inverse { flex: 0 0 auto; background: white; border-color: white; color: var(--blue); }
.solid-button.inverse:hover { background: #eef3ff; }
.site-footer { min-height: 170px; padding: 0 max(32px, calc((100vw - 1220px) / 2)); background: #10151d; color: #88919e; display: grid; grid-template-columns: 1fr 1fr auto; align-items: center; gap: 30px; }
.footer-wordmark { color: white; }.site-footer p { margin: 0; }
.auth-modal-body { --blue: #1757d2; --line: #d8dce1; --ink: #151a23; padding: 10px 10px 6px; color: var(--ink); }.auth-modal-body h3 { margin: 0 0 26px; font-size: 34px; letter-spacing: -.04em; }
.auth-tabs { display: grid; grid-template-columns: 1fr 1fr; margin-bottom: 26px; border-bottom: 0; gap: 8px; }
.auth-tabs button { padding: 12px; border: 0; border-bottom: 0; border-radius: 8px; background: none; cursor: pointer; }.auth-tabs button.active { border-color: var(--blue); color: var(--blue); font-weight: 700; background: rgba(23,87,210,0.06); }
.auth-submit { width: 100%; background: #1757d2; border-color: #1757d2; color: #fff; }.auth-submit:hover { background: #0d47bb; border-color: #0d47bb; }.auth-submit:disabled { opacity: .65; cursor: wait; }

@media (max-width: 980px) {
  .site-header { width: 100%; grid-template-columns: 1fr auto; }.site-nav { display: none; }
  .hero-section { width: 100%; min-height: 680px; padding-top: 64px; }.hero-background { background-position: 64% center; opacity: .2; }.hero-wash { background: linear-gradient(90deg, rgba(246,247,244,.98) 0%, rgba(246,247,244,.88) 58%, rgba(246,247,244,.4) 100%); }
  .proof-strip { width: calc(100% - 40px); grid-template-columns: 1fr 1fr; }.proof-intro { padding: 28px 0; }.proof-item { border-top: 0; padding-top: 20px; }
  .story-section { grid-template-columns: 45px 1fr; }.resume-story { grid-column: 2; }
  .photo-story { grid-template-columns: 1fr; }.photo-story figure img { min-height: 620px; }.photo-story-copy { padding: 80px 40px; }
  .tracker-section, .steps-section { grid-template-columns: 1fr; }.workspace-figure { height: 580px; }.voices-grid { grid-template-columns: 1fr; }.voice-card { min-height: 280px; border-right: 0; border-bottom: 0; }.steps-heading { position: static; }
  .final-cta { align-items: flex-start; flex-direction: column; }.site-footer { grid-template-columns: 1fr; padding-top: 46px; padding-bottom: 46px; }
}
@media (max-width: 640px) {
  .site-header { width: 100%; height: 72px; padding-inline: 16px; }.wordmark { font-size: 18px; }.header-actions .text-button { display: none; }.solid-button.compact { min-height: 38px; padding: 0 12px; }
  .hero-section { width: 100%; min-height: 680px; margin-top: 0; padding: 54px 22px 120px; }.hero-background { background-position: 66% center; opacity: .16; }.hero-wash { background: linear-gradient(90deg, rgba(246,247,244,.97), rgba(246,247,244,.72)); }h1 { font-size: 44px; }.hero-lead { font-size: 16px; }.hero-actions { flex-direction: column; align-items: stretch; }.hero-caption { left: 22px; right: 22px; bottom: 28px; width: auto; }
  .proof-strip { width: calc(100% - 32px); display: block; }.proof-intro { max-width: none; }.proof-item { padding: 24px 0; border-left: 0; }
  .editorial-section { width: calc(100% - 32px); padding: 90px 0; }.story-section { display: block; }.story-section > .section-number { margin-bottom: 50px; }.resume-story { min-height: 500px; margin-top: 50px; }.resume-story-main { height: 500px; }.resume-story-detail { top: 16px; right: 16px; width: 38%; }.resume-story figcaption { left: 18px; right: 18px; bottom: 18px; display: block; }.resume-story figcaption strong { display: block; margin-top: 8px; text-align: left; }
  .photo-story figure img { min-height: 0; height: 470px; }.photo-story-copy { padding: 70px 24px; }.photo-story-copy .section-number { margin-bottom: 50px; }
  .tracker-section { gap: 56px; }.tracker-copy .section-number { margin-bottom: 50px; }.tracker-flow { grid-template-columns: 1fr; gap: 8px; }.tracker-flow i { width: 1px; height: 14px; margin-left: 18px; }.workspace-figure { height: 500px; }
  .voices-section { padding: 90px 24px; }.voices-heading { display: block; margin-bottom: 50px; }.voices-heading .eyebrow { margin-bottom: 24px; }.voice-card { padding-left: 0; padding-right: 0; }
  .steps-section { padding: 90px 24px; gap: 54px; }.steps-list li { grid-template-columns: 36px 1fr; padding: 20px 0; }.steps-list li > :deep(svg) { display: none; }
  .final-cta { min-height: 390px; padding: 70px 24px; }.final-cta p { font-size: 42px; }.site-footer { padding-left: 24px; padding-right: 24px; }
}
@media (prefers-reduced-motion: reduce) { * { scroll-behavior: auto !important; transition: none !important; } }
</style>
