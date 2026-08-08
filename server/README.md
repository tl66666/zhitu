# 职途 后端服务

基于 Go + Gin + GORM + 纯 Go SQLite 驱动的求职全流程工具后端，覆盖用户档案、简历实验室、模拟面试、投递看板、求职 Copilot 等核心业务，并集成 LLM 能力提供 AI 生成、润色、评分、语音面试与复盘报告。

## 技术栈

| 层 | 选型 | 说明 |
|---|---|---|
| HTTP 框架 | `github.com/gin-gonic/gin` | 轻量高性能 |
| ORM | `gorm.io/gorm` | 主流 Go ORM |
| SQLite 驱动 | `github.com/glebarez/sqlite` | **纯 Go 实现**，底层 `modernc.org/sqlite`，无需 CGO |
| JWT | `github.com/golang-jwt/jwt/v5` | 标准 JWT 签发与校验 |
| 配置 | `gopkg.in/yaml.v3` | YAML 配置文件 |
| 密码哈希 | `golang.org/x/crypto/bcrypt` | 业界标准 |
| LLM 集成 | OpenAI / Anthropic / MiMo 协议 | 统一 `LLMService` 接口，支持文本对话、语音识别（ASR）、语音合成（TTS） |
| 流式响应 | SSE（Server-Sent Events） | 简历 AI 生成、求职 Copilot 流式回复 |

> 不引入缓存层、不接入第三方 OAuth，符合方案要求。

## 目录结构

```
server/
├── cmd/server/main.go              # 程序入口
├── configs/
│   ├── config.yaml                 # 配置文件
│   └── config.example.yaml         # 配置模板
└── internal/
    ├── config/                     # 配置加载
    ├── database/                   # 数据库初始化 + GORM AutoMigrate
    ├── models/                     # 数据模型
    │   ├── user.go
    │   ├── profile.go
    │   ├── resume.go
    │   ├── interview.go
    │   ├── delivery.go
    │   └── browser_workspace.go    # 浏览器工作区隔离
    ├── handlers/                   # HTTP 处理器
    │   ├── auth_handler.go
    │   ├── profile_handler.go
    │   ├── resume_handler.go
    │   ├── interview_handler.go
    │   ├── delivery_handler.go
    │   ├── copilot_handler.go      # 求职 Copilot
    │   ├── browser_state_handler.go # 浏览器状态
    │   └── admin_handler.go
    ├── services/                   # 业务逻辑
    │   ├── auth_service.go
    │   ├── jwt_service.go
    │   ├── profile_service.go
    │   ├── resume_service.go
    │   ├── resume_ai_service.go    # 简历 AI 操作
    │   ├── interview_service.go    # 面试管理
    │   ├── interview_flow.go       # 面试流程（状态机）
    │   ├── interview_report.go     # 面试复盘报告
    │   ├── delivery_service.go     # 投递管理
    │   ├── llm_service.go          # LLM 统一接口
    │   ├── llm_anthropic.go        # Anthropic/Claude 适配
    │   ├── llm_audio.go            # 语音识别与合成
    │   ├── resume_copilot_service.go # 求职 Copilot
    │   └── browser_state_service.go # 浏览器状态服务
    ├── middleware/
    │   ├── jwt.go                  # JWT 认证
    │   ├── cors.go                 # 跨域
    │   ├── recovery.go             # panic 恢复
    │   └── browser_workspace.go    # 浏览器工作区隔离
    ├── routers/router.go           # 路由注册
    └── utils/
        ├── response.go             # 统一响应封装
        ├── password.go             # 密码哈希
        └── file.go                 # 文件工具
```

## 分层架构

```
HTTP Request
    │
    ▼
[middleware]  Recovery → Logger → CORS → JWTAuth → BrowserWorkspace → (RequireAdmin)
    │
    ▼
[handlers]    解析请求 / 校验入参 / 调用 service / 组装响应（含 SSE 流式）
    │
    ▼
[services]    业务逻辑（状态机、AI 编排、简历解析、投递漏斗等，不依赖 gin）
    │
    ▼
[models]      GORM 模型
    │
    ▼
[database]    SQLite (glebarez/sqlite, 纯 Go, WAL 模式)
```

- `handlers` 只负责 HTTP 协议层：参数绑定、错误码映射、调用 service，AI 类接口负责 SSE 流式输出。
- `services` 不依赖 gin，方便后续被定时任务 / RPC 等复用；状态机、AI 编排、评分复盘等核心逻辑均在此层。
- `models` 仅描述数据结构，不含业务方法。
- `middleware` 提供横切关注点（认证、跨域、恢复、浏览器工作区隔离）。
- `database` 负责连接初始化、启用 WAL 与外键约束，并通过 `AutoMigrate` 自动建表与补字段。

## 用户体系说明

| 角色 | 凭据来源 | 登录入口 | 权限 |
|---|---|---|---|
| 管理员 | `configs/config.yaml` 的 `admin` 字段 | `POST /api/auth/admin/login` | 可访问 `/api/admin/*` |
| 普通用户 | 数据库 `users` 表 | `POST /api/auth/login` | 仅能访问 `/api/v1/*` 业务接口 |

- 多用户，**无用户组**，**无角色细分**：除管理员外所有普通用户权限一致。
- 管理员**不入库**，凭据完全由配置文件维护；如需修改，编辑 `config.yaml` 后重启服务。
- `config.yaml` 中管理员密码支持**明文**或 **bcrypt 哈希**（以 `$2a$` / `$2b$` / `$2y$` 开头时自动按哈希校验）。
- 普通用户的业务数据（档案、简历、面试、投递等）在登录用户维度之上，再按**浏览器工作区**隔离，详见下文「浏览器工作区」。

## 业务模块

### 认证模块（auth）

- **注册**：邮箱 + 密码注册，成功后**自动登录**并签发 Token，前端无需再次调用登录。
- **登录**：邮箱 + 密码，校验通过后签发 JWT。
- **改密**：校验旧密码后更新为新密码。
- **管理员登录**：凭配置文件中的管理员凭据签发管理员 Token（使用独立签名密钥）。
- **当前用户**：`GET /api/auth/me` 返回登录者信息。

### 用户档案（profile）

- 基础信息（姓名、性别、联系方式、求职意向等）的获取与更新。
- **6 类子资源 CRUD**：教育经历（educations）、工作经历（works）、项目经历（projects）、技能（skills）、荣誉（honors）、实践经历（practices），均按标准 REST 风格提供列表 / 创建 / 更新 / 删除。
- **简历解析自动填充**：上传简历后调用解析服务，自动回填档案字段，减少重复录入。
- **完成度计算**：根据各模块填写情况综合计算档案完成度，引导用户补全信息。

### 简历模块（resume）

- **多份简历**：一个用户可维护多份独立简历，支持 CRUD。
- **版本管理**：每次保存生成一个版本，支持版本列表、查看指定版本、回滚到历史版本。
- **AI 操作**（基于 LLM）：
  - `generate`：按指令生成简历内容，**SSE 流式**逐字返回。
  - `polish`：对现有内容润色优化。
  - `score`：多维度评分并给出改进建议。
  - `jd-match`：输入 JD 进行匹配度分析与定向优化建议。
- **同步档案**：将简历内容反向同步回用户档案，保持数据一致。

### 求职 Copilot

- 提供 **4 种 AI 任务**（如简历诊断、岗位分析、面试预测、求职策略等），覆盖求职关键决策点。
- 全部以 **SSE 流式回复**，边生成边呈现，体验流畅。
- 与浏览器工作区打通，对话上下文按浏览器隔离，多窗口互不干扰。

### 面试模块（interview）

- **7 状态状态机**（`interview_flow.go`）：管理面试从创建到结束的完整生命周期，状态流转受校验约束。
- **多种交互模式**：文字问答、语音问答、文字 + 语音混合模式。
  - 语音回答：前端录制音频上传，后端调用 ASR 转写为文字再进入对话。
  - 语音提问：AI 面试官的回答可通过 TTS 生成音频，前端按需播放。
- **评分复盘**：面试结束后生成多维度评分（`interview_report.go`）与可读复盘报告，支持取消与删除。
- 可关联指定简历版本作为面试上下文。

### 投递模块（delivery）

- **全流程追踪**：从投递到Offer的完整状态流转。
- **轮次记录**：一笔投递可记录多轮面试（一面 / 二面 / HR 面等），支持增删改。
- **HR 反馈**：记录每轮 HR 反馈与时间节点。
- **统计漏斗**：按状态聚合提供投递漏斗与统计数据，量化求职进展。
- **简历版本归属校验**：投递关联具体简历版本，确保回溯时能还原当时投递的简历内容。

### 浏览器工作区

- 通过请求头 `X-Browser-Token` 标识浏览器实例，后端使用 **SHA-256 哈希**映射为工作区标识。
- 普通用户的业务数据在「用户 + 浏览器工作区」两个维度隔离，同一用户在不同浏览器 / 窗口下的求职 Copilot 对话、浏览器状态等互不干扰。
- 工作区中间件统一注入工作区标识，业务层透明使用。

## 路由总览

| 分组 | 前缀 | 鉴权 | 说明 |
|---|---|---|---|
| 健康检查 | `/health` | 无 | 服务探活 |
| 认证 | `/api/auth/*` | 部分需登录 | 注册 / 登录 / 改密 / me |
| 业务模块 | `/api/v1/*` | 登录 + 浏览器工作区 | 档案 / 简历 / 面试 / 投递 / Copilot |
| 管理端 | `/api/admin/*` | 管理员身份 | 用户管理 / 投递管理 / 统计 |
| 静态资源 | `/static/*` | 无 | 访问上传的音频、TTS、简历文件 |

### 认证路由（`/api/auth`）

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/auth/register` | 无 | 注册（成功后自动登录） |
| POST | `/api/auth/login` | 无 | 登录 |
| POST | `/api/auth/admin/login` | 无 | 管理员登录 |
| GET | `/api/auth/me` | 登录 | 获取当前登录者信息 |
| POST | `/api/auth/change-password` | 登录 | 修改密码 |

### 用户档案路由（`/api/v1/profile`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/profile` | 获取档案 |
| PUT | `/api/v1/profile` | 更新档案 |
| GET | `/api/v1/profile/completion` | 档案完成度 |
| POST | `/api/v1/profile/parse-resume` | 解析简历自动填充 |
| GET / POST / PUT / DELETE | `/api/v1/profile/{educations,works,projects,skills,honors,practices}` | 6 类子资源 CRUD |

### 简历路由（`/api/v1/resumes`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET / POST | `/api/v1/resumes` | 列表 / 创建 |
| GET / PUT / DELETE | `/api/v1/resumes/:id` | 查看 / 更新 / 删除 |
| GET / POST | `/api/v1/resumes/:id/versions` | 版本列表 / 新建版本 |
| GET | `/api/v1/resumes/:id/versions/:vid` | 查看指定版本 |
| POST | `/api/v1/resumes/:id/rollback/:vid` | 回滚到指定版本 |
| POST | `/api/v1/resumes/:id/ai/generate` | AI 生成（SSE 流式） |
| POST | `/api/v1/resumes/:id/ai/polish` | AI 润色 |
| POST | `/api/v1/resumes/:id/ai/score` | AI 评分 |
| POST | `/api/v1/resumes/:id/ai/jd-match` | AI JD 匹配 |
| POST | `/api/v1/resumes/:id/sync-profile` | 同步至档案 |

### 面试路由（`/api/v1/interviews`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET / POST | `/api/v1/interviews` | 列表 / 创建 |
| GET | `/api/v1/interviews/:id` | 详情 |
| POST | `/api/v1/interviews/:id/resume` | 关联简历 |
| POST | `/api/v1/interviews/:id/messages` | 发送文字消息 |
| POST | `/api/v1/interviews/:id/transcribe` | 语音转文字 |
| POST | `/api/v1/interviews/:id/voice` | 发送语音回答 |
| GET | `/api/v1/interviews/:id/tts/:msgId` | 获取 TTS 音频 |
| POST | `/api/v1/interviews/:id/end` | 结束面试 |
| GET | `/api/v1/interviews/:id/report` | 复盘报告 |
| GET | `/api/v1/interviews/:id/scores` | 多维评分 |

### 投递路由（`/api/v1/deliveries`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET / POST | `/api/v1/deliveries` | 列表 / 创建 |
| GET | `/api/v1/deliveries/stats` | 统计 |
| GET | `/api/v1/deliveries/funnel` | 漏斗 |
| GET / PUT / DELETE | `/api/v1/deliveries/:id` | 查看 / 更新 / 删除 |
| PATCH | `/api/v1/deliveries/:id/status` | 变更状态 |
| GET / POST | `/api/v1/deliveries/:id/rounds` | 轮次列表 / 新增 |
| PUT / DELETE | `/api/v1/deliveries/:id/rounds/:rid` | 更新 / 删除轮次 |
| GET / POST | `/api/v1/deliveries/:id/feedbacks` | HR 反馈列表 / 新增 |
| DELETE | `/api/v1/deliveries/:id/feedbacks/:fid` | 删除反馈 |

### 管理端路由（`/api/admin`）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/ping` | 管理员身份探活 |
| GET | `/api/admin/stats` | 仪表盘统计 |
| GET | `/api/admin/users` | 用户列表 |
| GET | `/api/admin/users/:id` | 用户详情 |
| PATCH | `/api/admin/users/:id/status` | 启用 / 禁用用户 |
| POST | `/api/admin/users/:id/reset-password` | 重置用户密码 |
| GET | `/api/admin/deliveries` | 投递列表 |
| GET | `/api/admin/deliveries/funnel` | 投递漏斗 |

## 中间件设计

| 中间件 | 作用 |
|---|---|
| `Recovery` | 捕获 panic，输出 500 JSON，避免进程崩溃 |
| `gin.Logger` | 标准 access log |
| `CORS` | 按 `allow_origins` 配置放行跨域，OPTIONS 预检直接 204 |
| `JWTAuth` | 解析 `Authorization: Bearer`，校验签名/过期，注入 `user_id` / `email` / `is_admin` 到 gin.Context |
| `BrowserWorkspace` | 解析 `X-Browser-Token` 头，SHA-256 哈希后注入工作区标识，业务数据按浏览器隔离 |
| `RequireAdmin` | 在 `JWTAuth` 之后使用，校验 `is_admin=true`，否则 403 |

## 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

业务码：`0` 成功；`4xxxx` 客户端错误；`5xxxx` 服务端错误（详见 `utils/response.go`）。

SSE 流式接口（简历 AI 生成、求职 Copilot）按 `text/event-stream` 输出，逐块推送生成内容，结束后以统一事件携带上述结构作为收尾。

## 快速开始

### 1. 安装 Go

需要 Go 1.22+。

### 2. 安装依赖

```bash
cd server
go mod tidy
```

### 3. 修改配置

复制 `configs/config.example.yaml` 为 `configs/config.yaml` 并按需修改：

- `jwt.secret` / `jwt.admin_secret` 改为强随机字符串
- `admin.email` / `admin.password` 改为自定义管理员凭据
- `llm` 段填入 LLM 供应商的 `base_url` / `api_key` / 模型名称（AI 功能依赖此项）
- 如需对接前端，把前端地址加入 `server.allow_origins`

### 4. 运行

```bash
go run ./cmd/server
# 或编译后运行
go build -o bin/server ./cmd/server && ./bin/server
```

默认监听 `:8080`，数据库文件 `./data/zhitu.db` 自动创建并执行 AutoMigrate。

### 5. 接口验证

```bash
# 注册（成功后自动登录，返回 token）
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"abc12345"}'

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"abc12345"}'

# 用返回的 token 调用 /me
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <token>"

# 修改密码
curl -X POST http://localhost:8080/api/auth/change-password \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"abc12345","new_password":"xyz67890"}'

# 管理员登录
curl -X POST http://localhost:8080/api/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@zhitu.com","password":"admin123456"}'
```
