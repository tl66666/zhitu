# 职途

场景驱动的求职全流程工具，围绕「写简历、练面试、追投递」三个高频求职场景设计。

---

> **产品需求文档（PRD）在线查看：[点击进入 PRD 文档](https://tl66666.github.io/zhitu/docs/jobhunter-prd.html)**
>
> 含可交互原型 Demo，浏览器直接打开即可体验全部功能设计。

> **线上 Demo：[https://zhitu.kralai.tech/](https://zhitu.kralai.tech/)**

---

## 产品定位

面向应届生和初级求职者，解决求职过程中工具零散、信息不对称、缺乏沉浸式练习环境的痛点。产品以简历版本和投递记录为数据纽带，将三大核心模块串联成闭环：简历写完直接关联投递 → 投递后针对岗位练习面试 → 面试结果反馈到投递记录。数据在模块间自动流转，用户无需重复输入。

## 模块概览

| 模块 | 核心能力 | 差异化设计 |
|------|----------|------------|
| 登录与认证 | 邮箱密码注册登录，注册即登录，JWT 令牌管理 | — |
| 简历实验室 | 所见即所得编辑器，模块化管理、实时预览、版本控制、AI 驱动的生成/润色/评分/JD 匹配 | **SSE 流式 AI 生成** — 逐字呈现，体验流畅 |
| 面试训练场 | 文字/语音/混合模式模拟面试，7 状态状态机，AI 面试官实时提问，多维度评分复盘 | **语音交互** — ASR 转写 + TTS 播报，沉浸式面试体验 |
| 投递看板 | 全流程状态追踪，面试轮次与 HR 反馈记录，统计漏斗分析 | **简历版本关联** — 投递绑定具体简历版本，可回溯投递时的简历内容 |
| 求职 Copilot | 结合简历上下文的 AI 求职顾问，4 种 AI 任务（JD 匹配/项目优化/面试预测/自由问答） | **浏览器工作区隔离** — 按浏览器实例隔离对话，多窗口互不干扰 |

## 技术栈

| 层 | 技术 | 说明 |
|---|------|------|
| 后端 | Go + Gin + GORM | 轻量高性能，纯 Go SQLite 驱动（无需 CGO） |
| 前端 | Vue 3 + TypeScript + Vite | Composition API，类型安全 |
| 组件库 | Ant Design Vue + Tailwind CSS | 企业级 UI + 原子化样式 |
| 状态管理 | Pinia | Vue 3 官方推荐 |
| 认证 | JWT | 用户/管理员双密钥体系，独立签名 |
| 部署 | Cloudflare Pages + 自建服务器 | 前端 CDN，后端 API 服务 |
| 数据库 | SQLite | 嵌入式数据库，零配置部署 |
| AI 能力 | LLM API（OpenAI/Anthropic/MiMo） | 统一接口适配多模型，支持文本/语音 |
| 实时交互 | SSE（Server-Sent Events） | 简历生成、Copilot 流式响应 |

## 项目结构

```
zhitu/
├── server/                          # Go 后端
│   ├── cmd/server/main.go           # 程序入口
│   ├── configs/
│   │   ├── config.yaml              # 配置文件（不入库）
│   │   └── config.example.yaml      # 配置模板
│   └── internal/
│       ├── config/                  # 配置加载
│       ├── database/                # 数据库初始化 + GORM AutoMigrate
│       ├── models/                  # 数据模型
│       ├── handlers/                # HTTP 处理器
│       ├── services/                # 业务逻辑（状态机、AI 编排、评分复盘等）
│       ├── middleware/              # 中间件（JWT、CORS、浏览器工作区隔离）
│       ├── routers/                 # 路由注册
│       └── utils/                   # 工具函数
├── web/                             # Vue 3 前端
│   ├── src/
│   │   ├── api/                     # API 请求封装
│   │   ├── stores/                  # Pinia 状态管理
│   │   ├── views/                   # 页面组件
│   │   ├── components/              # 公共组件
│   │   ├── composables/             # 组合式函数
│   │   ├── router/                  # 路由配置
│   │   ├── utils/                   # 工具函数（请求封装、SSE、浏览器工作区）
│   │   └── types/                   # TypeScript 类型定义
│   ├── package.json
│   ├── vite.config.ts               # Vite 配置（含 API 代理、自动打开浏览器）
│   └── tailwind.config.js
├── docs/                            # 产品需求文档
│   ├── jobhunter-prd.html           # PRD（含可交互原型）
│   └── _shared/fonts/               # 字体文件
├── api.md                           # 后端 API 文档
├── CONTRIBUTING.md                  # 开发协作指南
├── CHANGELOG.md                     # 更新日志
├── .nojekyll                        # 绕过 Jekyll 处理
└── .gitignore
```

## 快速启动

### 后端（Go）

```bash
cd server
go mod tidy
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml 配置 LLM API Key
go run ./cmd/server
# 默认监听 :8080
```

### 前端（Vue）

```bash
cd web
npm install
npm run dev
# 默认监听 :3000，自动打开浏览器
# 开发模式下 /api 请求代理到后端 :8080
```

## 配置说明

核心配置项位于 `server/configs/config.yaml`（从 `config.example.yaml` 复制）：

| 配置项 | 说明 |
|-------|------|
| `server.port` | 后端服务端口（默认 8080） |
| `server.allow_origins` | 跨域允许的前端地址 |
| `jwt.secret` | 用户 JWT 签名密钥 |
| `jwt.admin_secret` | 管理员 JWT 独立签名密钥 |
| `admin.email` / `admin.password` | 管理员凭据（不入库） |
| `llm.provider` | LLM 供应商（openai/anthropic/mimo） |
| `llm.api_key` | LLM API 密钥 |
| `llm.chat_model` | 文字模型 |
| `llm.whisper_model` | 语音识别模型 |
| `llm.tts_model` | 语音合成模型 |

## 安全特性

- **浏览器工作区隔离**：通过 `X-Browser-Token` 的 SHA-256 哈希映射 `workspace_id`，实现多工作区数据隔离
- **权限控制**：所有数据操作包含 `user_id` 条件，防止越权访问
- **管理员隔离**：管理员 JWT 使用独立签名密钥，与用户体系完全隔离
- **输入验证**：所有用户输入经过严格校验，防止注入攻击

## 部署信息

| 环境 | 地址 | 说明 |
|------|------|------|
| 线上 Demo | https://zhitu.kralai.tech/ | Cloudflare Pages 部署 |
| PRD 文档 | https://tl66666.github.io/zhitu/docs/jobhunter-prd.html | GitHub Pages 部署 |
| 后端 API | https://api.zhitu.kralai.tech/ | 自建服务器 |

## 开发协作

详细规范见 [CONTRIBUTING.md](CONTRIBUTING.md)，更新日志见 [CHANGELOG.md](CHANGELOG.md)，API 文档见 [api.md](api.md)。

核心要点：

- `main` 分支只存放经过测试的稳定代码
- 功能开发在独立 `feature/` 分支进行
- 提交 PR 需通过 CI 检查并经队友 review
- 数据库 schema 变更需提前沟通
- 每日开工前 `git pull origin main` 保持本地代码同步
