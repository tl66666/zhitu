# 更新日志

本文件记录"职途"项目的所有重要变更。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [0.6.1] - 2026-08-09

### PR #14 (commit aa72f68)

#### 变更
- `.gitignore` 新增 `*.docx` 忽略规则
- `CONTRIBUTING.md` 修正 commit 示例（移除不存在的 GitHub OAuth 引用），功能拆分表更新为实际产品模块
- `server/README.md` 补充面试路由缺失文档（DELETE/start/mode/cancel），标注 SSE 端点
- `server/configs/config.example.yaml` 移除过时的 5173 端口引用
- `web/README.md` 修正 API 代理端口说明 5173 → 3000
- `web/vite.config.ts` 补充 `open: true` 配置，启动时自动打开浏览器

## [0.6.0] - 2026-08-09

### PR #12 (commit 93e20a7)

#### 变更
- Copilot 流式响应优化：`emitCopilotReply` 按 rune 分块推送，避免长文本一次性输出造成卡顿
- 面试删除取消状态限制：允许删除任意状态的面试记录（此前仅允许 `completed`/`cancelled` 状态删除）

## [0.5.0] - 2026-08-05

### PR #11 (commit d08e226)

#### 新增
- 浏览器工作区隔离机制：通过 `X-Browser-Token` 的 SHA-256 哈希映射 `workspace_id`，实现多工作区数据隔离
- Copilot SSE 流式回复：`ChatStream` 逐 token 返回，提升对话体验
- 面试记录删除功能：事务级联删除 + 状态保护 + `user_id` 越权防护
- 投递创建/更新时增加简历版本归属校验
- 前端新增浏览器工作区管理工具
- 前端新增简历选择重命名功能
- 前端新增面试删除入口
- 6 项回归测试

#### 安全
- `BrowserStateService` 增加字段白名单、大小限制与 JSON 校验
- 移除公共静态文件服务，减少攻击面

## [0.4.0] - 2026-07-29

### PR #10 (commit e421020)

#### 新增
- 注册成功后自动登录：复用 API 返回的 token，无需再次输入账号密码

#### 变更
- 密码最小长度从 6 位提升到 8 位

#### 移除
- 移除注册后手动登录的冗余流程

## [0.3.0] - 2026-07-24

### PR #9 (commit 80bbbf6)

#### 新增
- 面试状态机升级为 7 状态：`preparing` / `starting` / `ongoing` / `reviewing` / `completed` / `report_failed` / `cancelled`
- `starting` 状态租约机制：2 分钟超时自动回收，防止状态卡死
- 面试取消 API：`POST /api/v1/interviews/:id/cancel`
- 5 项回归测试

#### 变更
- 前端面试间 UI 大幅增强
- Copilot 页面重构

#### 修复
- 投递删除跨用户安全修复：增加 `user_id` 条件校验

## [0.2.0] - 2026-07-18

### PR #8 (commit 74463dc)

#### 新增
- SSE 终止事件检测 + 前端用户反馈
- 4 项回归测试

#### 变更
- `WriteTimeout` 从 15s 提升到 5min，防止 AI/SSE 长请求被截断

#### 修复
- 投递删除增加 `user_id` 条件，防止跨用户删除
- Copilot JSON 容错：提取重试 + schema 校验

## [0.1.0] - 2026-07-12

### PR #7

#### 新增
- Copilot 对话接入服务端模型分析
- MiMo TTS 面试语音播放

#### 变更
- 统一简历/面试/Copilot 页面前端布局
- 对话历史移至右侧抽屉
- 优化 Markdown 消息渲染
- 更新生产环境 API 配置

#### 修复
- 修复面试输入法问题
