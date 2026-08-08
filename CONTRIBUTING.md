# 开发协作指南

本文件定义团队的 Git 工作流、分支策略、测试规范与分工原则。所有成员在开始开发前请通读本文件。

## Git 工作流

### 分支策略

| 分支 | 用途 | 保护规则 |
|------|------|----------|
| `main` | 存放经过测试的稳定代码 | 禁止直接 push，只能通过 PR 合并 |
| `feature/*` | 功能开发分支 | 命名格式 `feature/功能名`，如 `feature/resume-preview` |
| `fix/*` | Bug 修复分支 | 命名格式 `fix/问题描述`，如 `fix/login-redirect` |

### 日常开发流程

```
1. git pull origin main              # 每日开工先拉取最新代码
2. git checkout -b feature/xxx       # 创建功能分支
3. ...开发与本地测试...
4. git add . && git commit -m "..."  # 提交代码
5. git push origin feature/xxx       # 推送到远程
6. 在 GitHub 上创建 PR → 通知队友 review
7. 队友 review + 本地测试 → 合并到 main
8. git checkout main && git pull     # 切回主分支并拉取最新
```

### Commit 信息规范

使用 conventional commits 格式：

```
<type>(<scope>): <description>

type:   feat | fix | docs | style | refactor | test | chore
scope:  模块名，如 resume | interview | dashboard | auth
```

示例：

```
feat(resume): 实现简历竞争力雷达的六维评分计算
fix(auth): 修复 GitHub OAuth 回调地址错误
docs(prd): 更新面试官画像库的字段规则
```

## 测试规范

### 本地测试流程

收到队友的 PR 后，按以下步骤进行本地测试：

```
1. git fetch origin pull/<PR编号>/head:pr-<PR编号>   # 拉取 PR 分支到本地
2. git checkout pr-<PR编号>                           # 切换到 PR 分支
3. 安装依赖（如有变更）
4. 运行测试套件
5. 手动验证功能点
6. 在 GitHub 上 review 代码并留下评论
7. 测试通过 → 合并 PR；测试失败 → 评论说明问题
```

### 测试命令

项目技术栈：后端 Go（Gin + GORM + SQLite），前端 Vue 3（TypeScript + Vite + Ant Design Vue + Tailwind CSS）。

**后端测试与构建检查：**

```bash
# 运行后端全部单元测试
cd server && go test ./...

# 构建检查（静态分析 + 编译验证）
cd server && go vet ./... && go build ./...
```

**前端类型检查与构建：**

```bash
# TypeScript 类型检查
cd web && npm run check

# 生产构建
cd web && npm run build
```

提交 PR 前请确保以上命令全部通过。

### 持续集成（CI）

项目使用 GitHub Actions 作为 CI 平台。每次向 `main` 发起 PR 时，GitHub Actions 会自动触发以下检查：

- 后端：`go vet` + `go build` + `go test ./...`
- 前端：`npm run check`（类型检查）+ `npm run build`（构建验证）

CI 全部通过是合并 PR 的前提条件。若 CI 失败，请在本地运行对应命令排查问题后重新推送。

## 分工原则

### 按完整功能划分

每个功能应包含前端 + 后端的完整实现，避免按层级切分（如「你做前端、我做后端」）。这样做的原因：

- 减少跨层沟通成本
- 每个人对完整功能负责，理解更深入
- 避免「等对方接口」的阻塞

### 功能拆分建议

| 功能 | 涉及模块 | 建议负责人 |
|------|----------|------------|
| 登录与认证 | 认证模块 | 后端为主 |
| 简历编辑器 | 简历实验室 | 前端为主 |
| 简历竞争力雷达 | 简历实验室 | 全栈 |
| JD 匹配优化 | 简历实验室 + Agent | 后端为主 |
| 语音面试交互 | 面试训练场 | 前端为主 |
| 面试官画像库 | 面试训练场 | 全栈 |
| 投递看板表格 | 投递看板 | 前端为主 |
| 求职策略洞察 | 投递看板 + Agent | 后端为主 |

### 避免冲突

- **不要同时修改同一文件**。开工前在群里说一声「我要改 xxx 文件」
- **数据库 schema 变更必须提前沟通**。任何涉及表结构、字段增减的改动，先讨论再动手
- **每日 `git pull`**。开工前拉取最新代码，避免基于过期代码开发

## 数据库变更规范

项目使用 GORM AutoMigrate 进行数据库表结构管理（后端 Go + GORM + SQLite），不使用独立的 migration 脚本工具。

数据库 schema 变更流程：

1. 在群里提出变更需求，说明原因和影响范围
2. 在 `server/internal/models/` 下新增或修改模型结构体（定义 GORM tag）
3. 在 `server/internal/database/database.go` 的 `autoMigrate` 函数中注册新增的模型
4. PR 中包含 model 更新和对应的测试代码
5. 合并后重启服务，GORM AutoMigrate 会自动同步表结构（建表、补字段）

注意事项：

- GORM AutoMigrate 只会**新增**列和表，**不会删除**列或修改列类型。如需删除字段，请在群里讨论后手动处理。
- 禁止直接在代码中硬编码 SQL 修改表结构，所有表结构变更通过 model 定义 + AutoMigrate 管理。
- 涉及外键关系的模型变更，注意 `autoMigrate` 中的注册顺序（被依赖的表先注册）。
