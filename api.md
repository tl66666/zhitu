# 职途 后端 API 文档

> 本文档描述后端已实现的所有接口、请求/响应格式、调用方法与注意事项。
>
> - 后端框架：Go + Gin + GORM + 纯 Go SQLite
> - 认证方式：JWT（`Authorization: Bearer <token>`）
> - 配置文件：`server/configs/config.yaml`（管理员凭据直接写在 config 中，不入库）
> - 默认服务端口：`8080`

---

## 目录

1. [通用约定](#1-通用约定)
2. [认证模块 /api/auth](#2-认证模块-apiauth)
3. [用户档案模块 /api/v1/profile](#3-用户档案模块-apiv1profile)
4. [简历模块 /api/v1/resumes](#4-简历模块-apiv1resumes)
5. [模拟面试模块 /api/v1/interviews](#5-模拟面试模块-apiv1interviews)
6. [投递看板模块 /api/v1/deliveries](#6-投递看板模块-apiv1deliveries)
7. [管理端模块 /api/admin](#7-管理端模块-apiadmin)
8. [静态资源与文件上传](#8-静态资源与文件上传)
9. [枚举值速查表](#9-枚举值速查表)
10. [注意事项](#10-注意事项)

---

## 1. 通用约定

### 1.1 统一响应结构

所有接口均返回如下 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code`：业务错误码，`0` 表示成功，非 0 表示失败
- `message`：人类可读的提示信息
- `data`：业务数据，失败时可能省略

### 1.2 业务错误码

| code   | HTTP 状态 | 含义         |
|--------|----------|--------------|
| 0      | 200      | 成功         |
| 40001  | 400      | 参数错误     |
| 40101  | 401      | 未认证/token 失效 |
| 40301  | 403      | 无权限       |
| 40401  | 404      | 资源不存在   |
| 40901  | 409      | 冲突（如邮箱已注册） |
| 50001  | 500      | 服务器内部错误 |

### 1.3 认证方式

除 `/api/auth/register`、`/api/auth/login`、`/api/auth/admin/login` 外，所有接口均需在请求头携带 JWT：

```
Authorization: Bearer <token>
```

- 普通用户 token：通过 `/api/auth/login` 或 `/api/auth/register` 获取
- 管理员 token：通过 `/api/auth/admin/login` 获取（管理员 JWT 与用户 JWT 使用**独立签名密钥**，互不混用）
- 管理端接口 `/api/admin/*` 额外校验 `is_admin=true`，普通用户 token 无法访问

### 1.4 路由前缀分层

| 前缀              | 说明                                  |
|------------------|---------------------------------------|
| `/api/auth/*`    | 认证相关（注册/登录/改密/me），不含 v1 |
| `/api/v1/*`      | 业务模块（需登录，普通用户与管理员均可访问） |
| `/api/admin/*`   | 管理端（需管理员身份）                 |
| `/health`、`/`   | 健康检查                              |
| `/static/*`      | 静态文件服务（上传的音频/TTS/简历）    |

> ⚠️ **重要**：认证模块路径是 `/api/auth/*`，**不带 v1**。业务模块才是 `/api/v1/*`。

---

## 2. 认证模块 /api/auth

### 2.1 用户注册

```
POST /api/auth/register
```

**请求体**

```json
{
  "email": "user@example.com",
  "password": "12345678",
  "nickname": "张三"
}
```

| 字段       | 类型   | 必填 | 校验规则                |
|-----------|--------|------|------------------------|
| email     | string | 是   | 合法邮箱格式            |
| password  | string | 是   | 长度 8-64              |
| nickname  | string | 否   | 最大 50 字符，缺省取邮箱前缀 |

**响应 data**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 604800
}
```

- 注册成功即返回 token，无需再次登录
- `expires_in` 单位为秒（默认 7 天）

**错误**

| code  | message              |
|-------|----------------------|
| 40001 | invalid request body |
| 40901 | email already registered |

---

### 2.2 用户登录

```
POST /api/auth/login
```

**请求体**

```json
{
  "email": "user@example.com",
  "password": "12345678"
}
```

**响应**：同注册接口，返回 `token / token_type / expires_in`。

> ⚠️ **重要**：登录接口**只返回 token，不返回 user 详情**。前端获取用户信息需调用 [`GET /api/auth/me`](#24-获取当前用户信息)。

**错误**：账号不存在或密码错误返回 `40101 invalid email or password`。用户被禁用（`status=disabled`）也返回 401。

---

### 2.3 管理员登录

```
POST /api/auth/admin/login
```

**请求体**

```json
{
  "email": "admin@zhitu.com",
  "password": "admin@zhitu.com"
}
```

- 管理员凭据来自 `config.yaml` 的 `admin.email` / `admin.password`，**不入库**
- 管理员 JWT 使用独立的签名密钥（`jwt.admin_secret`），与普通用户密钥隔离

**响应 data**：同登录接口，返回 token。`is_admin=true` 写入 JWT claims。

**错误**：凭据错误返回 `40101 invalid admin credentials`。

---

### 2.4 获取当前用户信息

```
GET /api/auth/me
```

**响应 data**

```json
{
  "id": 1,
  "email": "user@example.com",
  "nickname": "张三",
  "avatar": "",
  "is_admin": false
}
```

- 管理员 token 调用时返回 `id=0, nickname="admin", is_admin=true`
- 普通用户 token 调用时返回真实用户信息

---

### 2.5 修改密码

```
POST /api/auth/change-password
```

**请求体**

```json
{
  "old_password": "12345678",
  "new_password": "87654321"
}
```

| 字段          | 类型   | 必填 | 校验规则       |
|--------------|--------|------|---------------|
| old_password | string | 是   | -             |
| new_password | string | 是   | 长度 8-64     |

**响应**：`data` 为 `null`，`message` 为 `"password changed"`。

**错误**：旧密码错误返回 `40001 old password mismatch`。

---

## 3. 用户档案模块 /api/v1/profile

用户档案包含基础信息 + 6 类子资源（教育、工作、项目、技能、荣誉、实践）。所有接口需登录。

### 3.1 获取完整档案

```
GET /api/v1/profile
```

**响应 data**：返回完整档案对象，含基础信息与所有子资源列表。

```json
{
  "id": 1,
  "user_id": 1,
  "real_name": "张三",
  "gender": "male",
  "birth_date": "2000-01-01",
  "phone": "13800138000",
  "target_position": "后端开发",
  "target_city": "北京",
  "expected_salary": "20-30k",
  "job_status": "active",
  "self_introduction": "...",
  "educations": [...],
  "works": [...],
  "projects": [...],
  "skills": [...],
  "honors": [...],
  "practices": [...]
}
```

---

### 3.2 更新基础信息

```
PUT /api/v1/profile
```

**请求体**：仅允许更新以下字段（其他字段会被过滤忽略）

```json
{
  "real_name": "张三",
  "gender": "male",
  "birth_date": "2000-01-01",
  "phone": "13800138000",
  "target_position": "后端开发",
  "target_city": "北京",
  "expected_salary": "20-30k",
  "job_status": "active",
  "self_introduction": "..."
}
```

**允许字段白名单**：`real_name, gender, birth_date, phone, target_position, target_city, expected_salary, job_status, self_introduction`

**响应 data**：更新后的完整 profile 对象。

> 若 body 不含任何允许字段，返回 `40001 no updatable fields provided`。

---

### 3.3 获取档案完成度

```
GET /api/v1/profile/completion
```

**响应 data**

```json
{
  "completion_pct": 65
}
```

- 完成度为 0-100 的整数百分比
- 计算规则覆盖基础信息字段填充率 + 各子资源是否有数据

---

### 3.4 上传简历解析

```
POST /api/v1/profile/parse-resume
```

**请求**：`multipart/form-data`，字段 `file` 为简历文件

| 字段 | 类型 | 必填 | 校验规则                         |
|------|------|------|---------------------------------|
| file | file | 是   | 扩展名 `.pdf`/`.docx`/`.txt`，≤20MB |

**响应 data**：解析结果对象，已自动合并到用户档案。

> 解析使用 LLM，若 LLM 未配置则返回 500 错误。

---

### 3.5 子资源通用 CRUD

6 类子资源遵循相同的路由模式：

| 资源         | 路径                            |
|-------------|---------------------------------|
| 教育背景     | `/api/v1/profile/educations`    |
| 工作经历     | `/api/v1/profile/works`         |
| 项目经历     | `/api/v1/profile/projects`      |
| 技能         | `/api/v1/profile/skills`        |
| 荣誉奖项     | `/api/v1/profile/honors`        |
| 校内外实践   | `/api/v1/profile/practices`     |

每个资源支持 4 个标准操作：

| 方法     | 路径               | 说明           |
|---------|--------------------|---------------|
| GET     | `/{resource}`      | 列表           |
| POST    | `/{resource}`      | 创建           |
| PUT     | `/{resource}/:id`  | 更新（全字段 map） |
| DELETE  | `/{resource}/:id`  | 删除           |

#### 各资源创建请求体（POST）

**Education**

```json
{
  "school": "清华大学",        // 必填
  "major": "计算机科学",
  "degree": "本科",
  "start_date": "2018-09",
  "end_date": "2022-06",
  "gpa": "3.8/4.0",
  "courses": "数据结构, 操作系统",
  "exchange": ""
}
```

**Work**

```json
{
  "company": "字节跳动",       // 必填
  "position": "后端开发工程师",
  "start_date": "2022-07",
  "end_date": "2024-08",
  "description": "负责...",
  "leave_reason": "个人发展"
}
```

**Project**

```json
{
  "name": "分布式存储系统",    // 必填
  "role": "核心开发",
  "start_date": "2023-01",
  "end_date": "2023-06",
  "description": "...",
  "tech_stack": ["Go", "Redis"],   // 数组
  "url": "https://github.com/..."
}
```

**Skill**

```json
{
  "category": "编程语言",      // 必填
  "name": "Go",              // 必填
  "proficiency": "熟练"
}
```

**Honor**

```json
{
  "name": "国家奖学金",        // 必填
  "issuer": "教育部",
  "award_date": "2020-10",
  "level": "国家级"
}
```

**Practice**

```json
{
  "title": "校园技术沙龙",     // 必填
  "organization": "计算机协会",
  "start_date": "2019-09",
  "end_date": "2020-06",
  "description": "..."
}
```

#### 更新（PUT）

请求体为 `map[string]interface{}`，可传任意字段子集，如 `{"description": "新描述"}`。

#### 删除（DELETE）

响应 `data` 为 `null`，`message` 为 `"deleted"`。

---

## 4. 简历模块 /api/v1/resumes

所有接口需登录。简历支持多份、版本管理、AI 操作。

### 4.1 简历主表 CRUD

#### 列表

```
GET /api/v1/resumes
```

**响应 data**：当前用户的简历数组（无分页）。

#### 创建

```
POST /api/v1/resumes
```

**请求体**

```json
{
  "name": "字节后端简历",          // 必填
  "target_company": "字节跳动",
  "target_position": "后端开发",
  "target_jd": "岗位描述...",
  "scene": "manual",             // manual/jd/scenario，缺省 manual
  "initial_content": ""           // 可选初始内容 JSON
}
```

- 创建后自动生成 v1.0 版本

#### 获取

```
GET /api/v1/resumes/:id
```

**响应 data**：resume 对象，含 `current_version_id` 指向当前版本。

#### 更新

```
PUT /api/v1/resumes/:id
```

**请求体**：`map[string]interface{}`，可更新 `name / target_company / target_position / target_jd / scene` 等字段。

#### 删除

```
DELETE /api/v1/resumes/:id
```

---

### 4.2 版本管理

简历内容存在 `ResumeVersion` 表中，每次修改生成新版本，支持回滚。

#### 列出版本

```
GET /api/v1/resumes/:id/versions
```

#### 创建版本

```
POST /api/v1/resumes/:id/versions
```

**请求体**

```json
{
  "content": "{...}",          // 必填，简历内容 JSON 字符串
  "change_note": "润色了工作经历"
}
```

- 版本号自动递增（v1.0 → v1.1 → ...）
- 新版本自动设为 `current_version_id`

#### 获取版本

```
GET /api/v1/resumes/:id/versions/:vid
```

#### 回滚到指定版本

```
POST /api/v1/resumes/:id/rollback/:vid
```

- 基于 `:vid` 内容创建一个新版本（不删除中间版本），并把 `current_version_id` 指向新版本

---

### 4.3 AI 操作

> 所有 AI 接口依赖 LLM 配置。若 `config.yaml` 未配置 `llm.api_key`，接口返回 500 错误，提示"llm not configured"。

#### AI 生成简历（SSE 流式）

```
POST /api/v1/resumes/:id/ai/generate
```

**请求体**（可为空对象）

```json
{
  "target_jd": "岗位描述...",    // 可空，空则纯档案生成
  "scene": "jd",              // 可空，生成场景描述
  "module_hints": "重点突出项目经历"  // 可空，模块生成提示
}
```

**响应**：SSE（Server-Sent Events）流式推送

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

事件格式（每行 `data: {json}\n\n`）：

```json
// 进度增量
{"type": "delta", "content": "生成的文本片段..."}

// 完成
{"type": "done", "version": { /* 新版本对象 */ }}

// 错误
{"type": "error", "message": "错误信息"}
```

- 生成完成后会自动创建新版本并返回

#### AI 润色

```
POST /api/v1/resumes/:id/ai/polish
```

**请求体**

```json
{
  "module": "work",     // work/project/all
  "jd": "岗位描述..."   // 可空，润色参考
}
```

**响应 data**：润色后的新版本对象（非流式）。

#### AI 评分

```
POST /api/v1/resumes/:id/ai/score
```

**请求体**

```json
{
  "jd": "岗位描述..."   // 可空
}
```

**响应 data**：评分结果对象，含各维度分数与建议。

#### AI JD 匹配度分析

```
POST /api/v1/resumes/:id/ai/jd-match
```

**请求体**

```json
{
  "jd": "岗位描述..."   // 必填
}
```

**响应 data**：匹配度分析结果，含匹配项、缺失项、建议。

---

### 4.4 同步档案

```
POST /api/v1/resumes/:id/sync-profile
```

将当前简历版本内容反向同步回用户档案（如从简历中提取的工作经历补全档案）。

**响应**：`data` 为 `null`，`message` 为 `"synced to profile"`。

---

### 4.5 求职 Copilot

Copilot 是浏览器侧保留会话、服务端按任务读取简历上下文的轻量 Agent 工作流。服务端不落库聊天记录；浏览器用 Cookie 保存当前会话 ID，用 IndexedDB 保存消息（不支持 IndexedDB 时降级到 `localStorage`）。每次请求都会重新校验简历和版本归属。

支持四种任务：

| task | 用途 | 是否需要 JD |
|------|------|------------|
| `jd_match` | 简历-JD 匹配分数、优势、缺失能力和修改建议 | 是 |
| `project_optimize` | 选中一个项目，按 STAR 和实习校招场景多轮优化 | 否 |
| `interview_predict` | 预测岗位高频问题和简历追问风险 | 是 |
| `career_chat` | 结合简历上下文的求职自由问答 | 否 |

#### 对话

```
POST /api/v1/copilot/chat
```

**请求体**

```json
{
  "task": "jd_match",
  "resume_id": 12,                  // 已保存简历 ID；没有时传 0
  "version_id": 34,                 // 可选，缺省读取当前版本
  "jd": "岗位描述全文",              // jd_match/interview_predict 必填
  "project_index": 0,               // project_optimize 必填，从 0 开始
  "draft_content": "{...}",         // 可选，编辑器未保存内容或粘贴的简历
  "messages": [
    {"role": "user", "content": "先分析匹配度"},
    {"role": "assistant", "content": "..."}
  ]
}
```

当 `resume_id=0` 时，必须提供 `draft_content`。它可以是现有简历 JSON，也可以是纯文本；纯文本会作为临时简历上下文使用，不会创建简历记录。

**响应**为 SSE。先返回 `status`，完成时返回：

```json
{
  "type": "done",
  "message": {"role": "assistant", "content": "匹配度为 78 分……"},
  "result": {
    "task": "jd_match",
    "reply": "匹配度为 78 分……",
    "context": {"resume_id": 12, "version_id": 34, "using_draft": false},
    "match": {
      "match_score": 78,
      "strengths": ["有 Go 服务开发项目"],
      "missing_capabilities": ["缺少可验证的性能指标"],
      "requirement_map": [],
      "recommendations": ["补充个人负责范围和结果"]
    },
    "proposals": []
  }
}
```

`project_optimize` 会额外返回 `proposals`。提案只包含候选文案，不会自动修改简历；用户确认后调用应用接口。

#### 应用项目优化提案

```
POST /api/v1/copilot/apply
```

**请求体**

```json
{
  "resume_id": 12,
  "base_version_id": 34,
  "content": "{...}",
  "project_index": 0,
  "replacement_description": "使用 Go 重构支付接口，降低延迟",
  "replacement_tech_stack": ["Go", "MySQL"],
  "change_note": "Copilot 项目优化"
}
```

服务端会再次确认 `base_version_id` 仍为简历当前版本，只允许修改指定项目索引，并通过现有 `ResumeService.CreateVersion` 生成新版本。若期间简历已变化，返回 `409`，需要刷新上下文后重新生成提案。

---

## 5. 模拟面试模块 /api/v1/interviews

所有接口需登录。支持文字/语音/混合模式，AI 自动发问，结束后生成复盘报告。

### 5.1 创建面试会话

```
POST /api/v1/interviews
```

**请求体**

```json
{
  "scene": "tech",                  // 必填，见枚举
  "target_company": "字节跳动",
  "target_position": "后端开发",      // 必填
  "target_jd": "岗位描述...",         // 必填，面试官据此设题
  "resume_id": 12,                    // 必填，创建前选择简历
  "version_id": 34,                   // 可空，缺省使用该简历当前版本
  "difficulty": "mid",               // 可空，junior/mid/senior/mixed，默认 mid
  "total_questions": 5,              // 可空，5-15，默认 8
  "mode": "hybrid",                 // 可空，text/voice/hybrid，默认 hybrid
  "examiner_style": "追问深入",      // 可空
  "training_focus": "项目深挖"       // 可空
}
```

**scene 枚举**：`tech` / `behavior` / `pressure` / `hr` / `group` / `teaching` / `corporate` / `defense` / `client` / `public` / `medical` / `media` / `remote` / `system` / `aviation`

**mode 枚举**：`text`(纯文字打字) / `voice`(语音面对面，不展示文字输入框) / `hybrid`(默认语音输入、AI 自动播放语音，也可切换键盘输入)

- 创建时校验并固化所选简历版本快照与 JD，面试状态初始为 `preparing`，不会提前生成题目。
- 创建成功后调用 `POST /api/v1/interviews/:id/start`，生成基于 JD 和简历的第一道题并进入 `ongoing`。

**响应 data**：interview 对象。首题通过 `start` SSE 的 `started` 事件返回。

---

### 5.2 启动面试并生成首题（SSE）

```
POST /api/v1/interviews/:id/start
```

面试房间进入时调用。接口幂等，重复调用不会重复生成首题。

**响应事件**：

```json
{"type":"delta","content":"AI 提问片段..."}
{"type":"started","message":{ /* 首条 AI 提问 */ },"interview":{ /* ongoing 会话 */ }}
{"type":"error","message":"错误信息"}
```

首题 Prompt 同时包含创建时固化的 JD、简历快照和训练设置。

### 5.3 切换面试模式

```
PATCH /api/v1/interviews/:id/mode
```

**请求体**：`{"mode":"text"}`

可在 `preparing` 或 `ongoing` 状态下自由切换 `text / voice / hybrid`，不会重置已有对话。

### 5.4 列表

```
GET /api/v1/interviews
```

**响应 data**：当前用户的面试会话数组（无分页）。

---

### 5.5 获取面试详情

```
GET /api/v1/interviews/:id
```

**响应 data**

```json
{
  "interview": { /* 面试会话对象 */ },
  "messages": [ /* 消息列表 */ ]
}
```

---

### 5.6 发送文字回答（SSE 流式）

```
POST /api/v1/interviews/:id/messages
```

**请求体**

```json
{
  "content": "我的回答是..."   // 必填
}
```

**响应**：SSE 流式推送

事件格式：

```json
{"type": "delta", "content": "AI 回复片段..."}
{"type": "done", "message": { /* AI 完整消息对象 */ }}
{"type": "interview_ended", "message": {...}, "interview": {...}}  // 面试自动结束
{"type": "error", "message": "错误信息"}
```

- AI 回复完成后，若已达 `total_questions` 题数，面试自动结束并推送 `interview_ended` 事件
- 仅 `text` 和 `hybrid` 模式允许文字回答；`preparing` 状态需先调用 `start`。

---

### 5.7 发送语音回答（SSE 流式）

```
POST /api/v1/interviews/:id/voice
```

**请求**：`multipart/form-data`，字段 `audio` 为音频文件，`duration_sec` 为录音时长（可选）

| 字段   | 类型 | 必填 | 校验规则                                          |
|-------|------|------|--------------------------------------------------|
| audio | file | 是   | 扩展名 `mp3/wav/m4a/webm/ogg`，≤25MB              |

**响应**：SSE 流式推送

事件格式：

```json
{"type": "status", "message": "正在转写语音..."}
{"type": "delta", "content": "AI 回复片段..."}
{"type": "done", "message": { /* AI 完整消息对象 */ }}
{"type": "interview_ended"}
{"type": "error", "message": "错误信息"}
```

- 流程：录音结束后自动上传 → Whisper 转写 → 立即作为用户回答 → AI 流式回复。不会返回录音草稿，也不需要用户确认“转文字”或重录。
- 仅 `voice` 和 `hybrid` 模式允许语音回答。
- 依赖 LLM 的 Whisper 模型配置

---

### 5.8 获取 AI 提问的 TTS 音频

```
GET /api/v1/interviews/:id/tts/:msgId
```

- `voice/hybrid` 模式的前端会在收到 AI 文字后自动请求并播放 TTS；语音模式下文字仅作为状态数据，不展示文字输入框。
- 首次请求时懒生成并缓存到本地文件，后续直接返回缓存

**响应**：二进制音频流

```
Content-Type: audio/wav
Content-Disposition: inline; filename="tts_xxx.wav"
```

- 依赖 LLM 的 TTS 模型配置

---

### 5.9 结束面试并生成复盘

```
POST /api/v1/interviews/:id/end
```

- 主动结束面试，AI 自动生成复盘报告
- 若已达题数自动结束，则无需调用此接口

**响应 data**：复盘报告对象，含整体评价、各维度评分、改进建议。

---

### 5.10 获取复盘报告

```
GET /api/v1/interviews/:id/report
```

**错误**：若面试未结束或报告未生成，返回 `40401 report not generated yet, please end the interview first`。

---

### 5.11 获取评分明细

```
GET /api/v1/interviews/:id/scores
```

**响应 data**：评分明细数组，每项含维度名、分数、评语。

---

## 6. 投递看板模块 /api/v1/deliveries

所有接口需登录。投递看板管理求职进度，含面试轮次与 HR 反馈。

### 6.1 投递主表

#### 列表

```
GET /api/v1/deliveries?status=&channel=
```

| 参数     | 类型   | 说明                          |
|---------|--------|-------------------------------|
| status  | string | 按状态过滤，见枚举             |
| channel | string | 按渠道过滤                     |

**响应 data**：投递数组（无分页）。

#### 创建

```
POST /api/v1/deliveries
```

**请求体**

```json
{
  "company": "字节跳动",          // 必填
  "position": "后端开发",         // 必填
  "channel": "boss",            // 必填，见枚举
  "apply_date": "2024-08-01",   // 必填，YYYY-MM-DD
  "priority": "high",           // 可空，默认 medium
  "jd_text": "岗位描述...",
  "resume_version_id": 12,       // 关联简历版本快照
  "hr_contact": "{\"name\":\"HR张三\"}",   // JSON 字符串
  "next_step": "{\"todo\":\"二面\",\"deadline\":\"2024-08-10\"}",
  "remark": "备注"
}
```

- 创建后状态默认为 `pending`

#### 获取详情

```
GET /api/v1/deliveries/:id
```

**响应 data**

```json
{
  "delivery": { /* 投递主表 */ },
  "rounds": [ /* 面试轮次列表 */ ],
  "feedbacks": [ /* HR 反馈列表 */ ]
}
```

#### 更新

```
PUT /api/v1/deliveries/:id
```

**请求体**：`map[string]interface{}`，可更新任意字段。

#### 删除

```
DELETE /api/v1/deliveries/:id
```

#### 变更状态

```
PATCH /api/v1/deliveries/:id/status
```

**请求体**

```json
{
  "status": "interview"   // 必填
}
```

**状态流转规则**（非法流转返回 `40001 invalid status transition`）：

```
pending → written_test / interview / rejected
written_test → interview / waiting_offer / rejected
interview → waiting_offer / rejected
waiting_offer → offer / rejected
offer → rejected （放弃 Offer）
rejected → interview （复活）
```

---

### 6.2 面试轮次

#### 列表

```
GET /api/v1/deliveries/:id/rounds
```

#### 创建

```
POST /api/v1/deliveries/:id/rounds
```

**请求体**（`DeliveryRound` 模型）

```json
{
  "round_type": "first_tech",      // 必填，见枚举
  "interview_time": "2024-08-05T14:00:00Z",
  "format": "video",             // onsite/video/phone
  "interviewer_name": "李四",
  "interviewer_title": "技术专家",
  "question_summary": "问了分布式锁、MySQL索引...",
  "feedback": "回答得不错",
  "result": "pass"               // pass/pending/rejected，默认 pending
}
```

#### 更新

```
PUT /api/v1/deliveries/:id/rounds/:rid
```

**请求体**：`map[string]interface{}`。

#### 删除

```
DELETE /api/v1/deliveries/:id/rounds/:rid
```

---

### 6.3 HR 反馈

#### 列表

```
GET /api/v1/deliveries/:id/feedbacks
```

#### 创建

```
POST /api/v1/deliveries/:id/feedbacks
```

**请求体**（`DeliveryFeedback` 模型）：HR 反馈内容对象。

- 创建反馈时会自动将 `next_step` 同步到投递主表

#### 删除

```
DELETE /api/v1/deliveries/:id/feedbacks/:fid
```

---

### 6.4 统计与漏斗

#### 统计

```
GET /api/v1/deliveries/stats
```

**响应 data**：当前用户的投递统计，含总数、进行中、Offer 数、被拒数等。

#### 漏斗

```
GET /api/v1/deliveries/funnel
```

**响应 data**：各状态节点数量，用于漏斗图展示。

> ⚠️ **路由顺序**：`/stats` 和 `/funnel` 注册在 `/:id` 之前，避免被当作 id 匹配。

---

## 7. 管理端模块 /api/admin

所有接口需**管理员 token**（`is_admin=true`），普通用户 token 返回 `40301 admin only`。

### 7.1 仪表盘统计

```
GET /api/admin/stats
```

**响应 data**：全局统计数据，含用户总数、简历数、面试数、投递数等。

---

### 7.2 用户管理

#### 用户列表

```
GET /api/admin/users?page=1&page_size=20&keyword=&status=
```

| 参数       | 类型   | 默认 | 说明                |
|-----------|--------|------|---------------------|
| page      | int    | 1    | 页码                |
| page_size | int    | 20   | 每页条数            |
| keyword   | string | -    | 按邮箱/昵称模糊搜索 |
| status    | string | -    | active/disabled     |

#### 用户详情

```
GET /api/admin/users/:id
```

**响应 data**：用户详细信息。

#### 切换用户状态

```
PATCH /api/admin/users/:id/status
```

**请求体**

```json
{
  "status": "disabled"   // active/disabled
}
```

- 禁用后用户无法登录

#### 重置用户密码

```
POST /api/admin/users/:id/reset-password
```

**请求体**

```json
{
  "new_password": "newpass123"   // 至少 8 位
}
```

---

### 7.3 投递管理

#### 全局投递列表

```
GET /api/admin/deliveries?page=1&page_size=20&status=&company=&user_email=
```

| 参数       | 类型   | 说明                |
|-----------|--------|---------------------|
| page      | int    | 页码，默认 1        |
| page_size | int    | 每页条数，默认 20   |
| status    | string | 按状态过滤          |
| company   | string | 按公司模糊搜索      |
| user_email| string | 按用户邮箱过滤      |

#### 全局投递漏斗

```
GET /api/admin/deliveries/funnel
```

**响应 data**：全局各状态投递数量。

---

## 8. 静态资源与文件上传

### 8.1 静态文件服务

```
GET /static/*
```

- 映射到 `config.storage.base_dir` 目录
- 用于访问上传的音频、TTS、简历文件
- 目录结构：`{base_dir}/{audio,tts,resume}/...`

### 8.2 文件上传限制

| 场景           | 接口                            | 允许扩展名                    | 大小限制 |
|---------------|---------------------------------|------------------------------|---------|
| 简历解析       | POST /api/v1/profile/parse-resume | pdf/docx/txt                | 20MB    |
| 面试语音回答   | POST /api/v1/interviews/:id/voice | mp3/wav/m4a/webm/ogg        | 25MB    |

---

## 9. 枚举值速查表

### 9.1 投递状态（Delivery.status）

| 值             | 含义     |
|----------------|---------|
| pending        | 待响应   |
| written_test   | 笔试中   |
| interview      | 面试中   |
| waiting_offer  | 待 Offer |
| offer          | 已 Offer |
| rejected       | 已拒绝   |

### 9.2 投递渠道（Delivery.channel）

| 值        | 含义     |
|----------|---------|
| boss     | BOSS 直聘 |
| official | 官方投递 |
| referral | 内推     |
| campus   | 校园招聘 |
| headhunt | 猎头     |
| other    | 其他     |

### 9.3 投递优先级（Delivery.priority）

| 值     | 含义 |
|-------|-----|
| high  | 高   |
| medium| 中   |
| low   | 低   |

### 9.4 面试轮次类型（DeliveryRound.round_type）

| 值            | 含义       |
|--------------|-----------|
| written_test | 笔试       |
| first_tech   | 一面技术   |
| second_tech  | 二面技术   |
| third_tech   | 三面技术   |
| cross        | 交叉面     |
| hr           | HR 面      |
| additional   | 加面       |
| final        | 终面       |

### 9.5 面试形式（DeliveryRound.format）

| 值      | 含义   |
|--------|--------|
| onsite | 现场   |
| video  | 视频   |
| phone  | 电话   |

### 9.6 轮次结果（DeliveryRound.result）

| 值       | 含义   |
|---------|--------|
| pass    | 通过   |
| pending | 待定   |
| rejected| 淘汰   |

### 9.7 模拟面试场景（Interview.scene）

| 值        | 含义   |
|----------|--------|
| tech     | 技术面 |
| behavior | 行为面 |
| pressure | 压力面 |
| hr       | HR 面  |
| group    | 群面   |

### 9.8 模拟面试模式（Interview.mode）

| 值     | 含义                  |
|-------|----------------------|
| text  | 纯文字               |
| voice | 语音                 |
| hybrid| 系统语音 + 用户文字   |

### 9.9 面试状态（Interview.status）

| 值         | 含义   |
|-----------|--------|
| ongoing   | 进行中 |
| completed | 已完成 |
| cancelled | 已取消 |

### 9.10 用户状态（User.status）

| 值       | 含义   |
|---------|--------|
| active  | 正常   |
| disabled| 禁用   |

### 9.11 简历场景（Resume.scene）

| 值       | 含义       |
|---------|-----------|
| manual  | 手动新增   |
| jd      | 针对 JD    |
| scenario| 场景定制   |

---

## 10. 注意事项

### 10.1 路由前缀

- **认证模块路径为 `/api/auth/*`，不带 v1**。业务模块才是 `/api/v1/*`。前端常因写成 `/api/v1/auth/login` 导致 404。
- 管理端路径为 `/api/admin/*`，需管理员 token。

### 10.2 JWT 双密钥隔离

- 普通用户 JWT 使用 `jwt.secret` 签名
- 管理员 JWT 使用 `jwt.admin_secret` 签名
- 两者互不混用，前端应使用独立的 localStorage 键名存储（如 `zhitu-auth` 与 `zhitu-admin-auth`）

### 10.3 登录响应只返回 token

- `/api/auth/login` 和 `/api/auth/admin/login` **只返回 token**，不返回 user 详情
- 获取用户信息需调用 `GET /api/auth/me`
- 前端登录成功后建议异步调用 `/api/auth/me` 拉取真实用户信息

### 10.4 SSE 流式接口

以下接口返回 SSE 流，需用 `EventSource` 或 `fetch` + `ReadableStream` 接收：

- `POST /api/v1/resumes/:id/ai/generate`
- `POST /api/v1/interviews/:id/messages`
- `POST /api/v1/interviews/:id/voice`

事件格式统一为 `data: {"type": "...", ...}\n\n`，type 取值：

| type              | 含义                     |
|-------------------|--------------------------|
| delta             | 流式文本增量             |
| status            | 状态提示（如"正在转写"） |
| done              | 正常完成，携带最终数据   |
| interview_ended   | 面试自动结束             |
| error             | 错误                     |

### 10.5 AI 接口依赖 LLM 配置

所有 AI 相关接口（简历生成/润色/评分/JD 匹配、面试发问/转写/TTS、简历解析）均依赖 `config.yaml` 中的 `llm` 配置：

```yaml
llm:
  provider: openai          # OpenAI 兼容接口
  base_url: https://api.openai.com/v1
  api_key: sk-xxx
  chat_model: gpt-4o        # 对话模型
  whisper_model: whisper-1  # 语音转写
  tts_model: tts-1          # 文字转语音
```

- 未配置 `api_key` 时，AI 接口返回 500 错误，提示 "llm not configured"
- **服务启动不依赖 LLM**，未配置时业务模块仍可正常注册路由、启动服务，仅 AI 接口报错

### 10.6 路由参数顺序

`/api/v1/deliveries` 下，`/stats` 和 `/funnel` 必须注册在 `/:id` 之前，否则会被 Gin 当作 `:id` 匹配导致 404。后端已正确处理，前端调用时注意路径不要冲突。

### 10.7 子资源更新为 map 模式

用户档案子资源的 `PUT` 接口接收 `map[string]interface{}`，可传任意字段子集进行局部更新，无需传完整对象。

### 10.8 文件上传字段名

| 接口                            | 字段名 |
|---------------------------------|--------|
| POST /api/v1/profile/parse-resume | file   |
| POST /api/v1/interviews/:id/voice | audio  |

### 10.9 分页参数

仅管理端用户列表与投递列表支持分页（`page` / `page_size`）。业务模块列表接口（简历、面试、投递、档案子资源）均**无分页**，返回当前用户全量数据。

### 10.10 用户隔离

- 业务模块接口均按 JWT 中的 `user_id` 过滤数据，用户只能访问自己的数据
- 管理端接口可访问全部用户数据，但需管理员身份

### 10.11 静态文件

上传的文件存储在 `config.storage.base_dir` 下，通过 `/static/*` 访问。生产环境建议配置 Nginx 直接托管静态文件，减轻后端压力。

### 10.12 健康检查

```
GET /health
GET /
```

返回 `{"status": "ok", "service": "zhitu-server"}`，可用于负载均衡健康探测。
