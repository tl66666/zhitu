# 职途 前端 Web

「职途」求职全流程工具的 Web 前端，提供简历实验室、模拟面试训练场、投递看板、求职 Copilot 等可视化交互界面，与后端 `/api` 接口对接。

## 技术栈

| 类别 | 选型 | 说明 |
|---|---|---|
| 框架 | Vue 3 | Composition API + `<script setup>` 单文件组件 |
| 语言 | TypeScript | 全量类型约束 |
| 构建工具 | Vite | 快速冷启动 + HMR |
| 组件库 | Ant Design Vue | 企业级 UI 组件 |
| 样式方案 | Tailwind CSS | 原子化样式，配合组件库做布局与定制 |
| 状态管理 | Pinia | 轻量状态管理，支持持久化插件 |
| 路由 | Vue Router | 前端路由 |
| HTTP | Axios | 请求封装与拦截器 |
| 其他 | dayjs、lucide-vue-next、clsx、tailwind-merge | 时间处理、图标、类名合并 |

## 目录结构

```
web/
├── src/
│   ├── api/                       # API 请求封装
│   │   ├── auth.ts
│   │   ├── profile.ts
│   │   ├── resume.ts
│   │   ├── interview.ts
│   │   ├── delivery.ts
│   │   ├── admin.ts
│   │   └── browserState.ts        # 浏览器状态 API
│   ├── stores/                    # Pinia 状态管理
│   │   ├── auth.ts
│   │   ├── profile.ts
│   │   ├── resume.ts
│   │   ├── interview.ts
│   │   ├── delivery.ts
│   │   └── copilot.ts
│   ├── views/                     # 页面组件
│   │   ├── Home.vue               # 首页（登录/注册）
│   │   ├── Dashboard.vue          # 工作台首页
│   │   ├── ResumeList.vue         # 简历列表
│   │   ├── ResumeEditor.vue       # 简历编辑器
│   │   ├── InterviewSceneSelect.vue # 面试场景选择
│   │   ├── InterviewList.vue      # 面试列表
│   │   ├── InterviewRoom.vue      # 面试间
│   │   ├── DeliveryKanban.vue     # 投递看板
│   │   ├── Copilot.vue            # 求职 Copilot
│   │   └── Profile.vue            # 个人档案
│   ├── components/                # 公共组件
│   ├── composables/               # 组合式函数
│   ├── router/                    # 路由配置
│   ├── utils/                     # 工具函数
│   │   ├── request.ts             # Axios 请求封装
│   │   ├── sse.ts                 # SSE 工具
│   │   └── browserScope.ts        # 浏览器工作区管理
│   └── types/                     # TypeScript 类型
├── package.json
├── vite.config.ts                 # Vite 配置（含 API 代理）
├── tailwind.config.js
└── tsconfig.json
```

## 快速开始

### 1. 环境要求

- Node.js 18+
- npm（或 pnpm / yarn）

### 2. 安装依赖

```bash
cd web
npm install
```

### 3. 常用命令

```bash
npm run dev          # 开发模式，默认 :5173，自动打开浏览器
npm run check        # TypeScript 类型检查
npm run build        # 生产构建
npm run preview      # 预览生产构建
```

### 4. 环境变量

- 开发环境：`.env.development`
- 生产环境：`.env.production`（配置生产环境 API 地址）

## Vite 配置要点

- **开发模式 API 代理**：`/api` → `http://localhost:8080`，前端请求 `/api/*` 会被代理到后端，避免跨域问题。
- **自动打开浏览器**：`server.open: true`，启动开发服务器后自动唤起浏览器。
- **路径别名**：`@` 指向 `src/`，import 时可使用 `@/api/auth` 等简写。
- **生产环境 API 地址**：通过 `.env.production` 中的变量配置，构建时注入。

## 关键模块说明

### 请求封装（utils/request.ts）

基于 Axios 封装统一请求实例：

- 请求拦截器自动注入 `Authorization: Bearer <token>` 与 `X-Browser-Token` 头。
- 响应拦截器统一处理后端 `{ code, message, data }` 格式，按业务码抛错或返回 `data`。
- 401 时清理登录态并跳转登录页。

### SSE 工具（utils/sse.ts）

封装 `fetch` + `ReadableStream` 的流式读取，用于简历 AI 生成、求职 Copilot 等流式接口，逐块解析并回调渲染。

### 浏览器工作区（utils/browserScope.ts）

- 为当前浏览器生成并持久化一个 `X-Browser-Token`。
- 所有业务请求自动携带该 token，配合后端实现「用户 + 浏览器」双维度数据隔离。
- 多窗口 / 多标签页各自独立，互不干扰。

### 状态管理（stores/）

按业务域拆分 Pinia store（auth / profile / resume / interview / delivery / copilot），负责对应模块的数据缓存与跨组件共享，登录态支持持久化。
