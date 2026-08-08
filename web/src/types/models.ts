// ==================== 通用类型 ====================

// 后端统一响应结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页请求参数
export interface PaginationParams {
  page?: number
  page_size?: number
}

// 分页响应数据（管理端用）
export interface PaginationData<T> {
  total: number
  list: T[]
}

// ==================== 用户认证相关 ====================

// 用户状态
export type UserStatus = 'active' | 'disabled'

// 用户信息（GET /api/auth/me 返回）
export interface User {
  id: number
  email: string
  nickname: string
  avatar?: string
  is_admin?: boolean
  status?: UserStatus
  created_at?: string
  updated_at?: string
}

// 登录请求
export interface LoginRequest {
  email: string
  password: string
}

// 登录返回的 token 数据（位于 ApiResponse.data 中）
export interface LoginTokenData {
  token: string
  token_type: string
  expires_in: number
}

// 登录响应 = 后端统一响应包裹的 token 数据
export type LoginResponse = ApiResponse<LoginTokenData>

// 注册请求
export interface RegisterRequest {
  email: string
  password: string
  nickname?: string
}

// 修改密码请求
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

// ==================== 用户档案相关 ====================

// 性别枚举
export type Gender = 'male' | 'female' | 'other'

// 求职状态
export type JobStatus = 'fresh' | 'graduated' | 'employed' | 'resigned'

// 用户档案主表
export interface Profile {
  user_id: number
  real_name: string
  gender: Gender | ''
  birth_date: string | null
  phone: string
  target_position: string
  target_city: string
  expected_salary: string
  job_status: JobStatus | ''
  self_introduction: string
  completion_pct: number
  updated_at: string
}

// 完整档案（含子资源）
export interface FullProfile extends Profile {
  educations: Education[]
  works: Work[]
  projects: Project[]
  skills: Skill[]
  honors: Honor[]
  practices: Practice[]
}

// 更新档案基础信息请求（仅允许白名单字段）
export interface UpdateProfileRequest {
  real_name?: string
  gender?: Gender
  birth_date?: string
  phone?: string
  target_position?: string
  target_city?: string
  expected_salary?: string
  job_status?: JobStatus
  self_introduction?: string
}

// 教育背景
export interface Education {
  id?: number
  user_id?: number
  school: string
  major: string
  degree: string
  start_date: string
  end_date: string
  gpa: string
  courses: string
  exchange: string
  created_at?: string
  updated_at?: string
}

// 工作/实习经历
export interface Work {
  id?: number
  user_id?: number
  company: string
  position: string
  start_date: string
  end_date: string
  description: string
  leave_reason: string
  created_at?: string
  updated_at?: string
}

// 项目经历
export interface Project {
  id?: number
  user_id?: number
  name: string
  role: string
  start_date: string
  end_date: string
  description: string
  tech_stack: string // JSON 数组字符串 ["Go","React"]
  url: string
  created_at?: string
  updated_at?: string
}

// 技能
export interface Skill {
  id?: number
  user_id?: number
  category: string
  name: string
  proficiency: string
  created_at?: string
  updated_at?: string
}

// 荣誉奖项
export interface Honor {
  id?: number
  user_id?: number
  name: string
  issuer: string
  award_date: string
  level: string
  created_at?: string
  updated_at?: string
}

// 校内外实践
export interface Practice {
  id?: number
  user_id?: number
  title: string
  organization: string
  start_date: string
  end_date: string
  description: string
  created_at?: string
  updated_at?: string
}

// 简历解析结果
export interface ParseResumeResult {
  [key: string]: unknown
}

// ==================== 简历相关 ====================

export type CopilotTask = 'jd_match' | 'project_optimize' | 'interview_predict' | 'career_chat'

export interface CopilotMessage {
  role: 'user' | 'assistant'
  content: string
  created_at: string
  result?: CopilotResponse
}

export interface CopilotRequirement {
  title: string
  priority: string
  status: string
  evidence: string[]
  gap: string
}

export interface CopilotMatchResult {
  match_score: number
  strengths: string[]
  missing_capabilities: string[]
  requirement_map: CopilotRequirement[]
  recommendations: string[]
}

export interface CopilotProjectResult {
  current_issues: string[]
  star_analysis: Record<string, string>
  technical_highlights: string[]
  missing_evidence: string[]
  rewritten_description: string
  rewritten_tech_stack: string[]
}

export interface CopilotInterviewQuestion {
  question: string
  type: string
  priority: string
  reason: string
  answer_plan: string
}

export interface CopilotPredictionResult {
  risk_points: string[]
  resume_triggers: string[]
  questions: CopilotInterviewQuestion[]
}

export interface CopilotProposal {
  id: string
  kind: string
  title: string
  rationale: string
  project_index?: number
  replacement_description?: string
  replacement_tech_stack?: string[]
}

export interface CopilotResponse {
  task: CopilotTask
  reply: string
  context: {
    resume_id: number
    version_id: number
    resume_name: string
    target_position: string
    jd?: string
    project_index?: number
    using_draft: boolean
  }
  match?: CopilotMatchResult
  project?: CopilotProjectResult
  prediction?: CopilotPredictionResult
  proposals?: CopilotProposal[]
  memory_summary?: string
}

export interface CopilotSession {
  id: string
  resume_id: number | null
  version_id: number | null
  task: CopilotTask
  jd: string
  draft_content?: string
  project_index?: number
  messages: CopilotMessage[]
  summary: string
  created_at: string
  updated_at: string
}

export interface CopilotChatRequest {
  task: CopilotTask
  resume_id: number
  version_id?: number
  jd?: string
  project_index?: number
  draft_content?: string
  messages: Array<{ role: 'user' | 'assistant'; content: string }>
}

// 简历场景
export type ResumeScene = 'manual' | 'jd' | 'scenario'

// 简历主表
export interface Resume {
  id: number
  user_id: number
  name: string
  target_company: string
  target_position: string
  target_jd: string
  scene: ResumeScene
  current_version_id: number
  created_at: string
  updated_at: string
}

// 创建简历请求
export interface CreateResumeRequest {
  name: string
  target_company?: string
  target_position?: string
  target_jd?: string
  scene?: ResumeScene
  initial_content?: string
}

// 更新简历请求
export interface UpdateResumeRequest {
  name?: string
  target_company?: string
  target_position?: string
  target_jd?: string
  scene?: ResumeScene
}

// 简历版本
export interface ResumeVersion {
  id: number
  resume_id: number
  version_label: string
  content: string
  change_note: string
  created_at: string
}

// 创建版本请求
export interface CreateVersionRequest {
  content: string
  change_note?: string
}

// AI 生成简历入参
export interface GenerateInput {
  target_jd?: string
  scene?: string
  module_hints?: string
}

// AI 润色入参
export interface PolishInput {
  module: 'work' | 'project' | 'all'
  jd?: string
}

// AI 评分入参
export interface ScoreInput {
  jd?: string
}

// AI JD 匹配入参
export interface JDMatchInput {
  jd: string
}

// AI 评分结果
export interface ScoreResult {
  [key: string]: unknown
}

// AI JD 匹配结果
export interface JDMatchResult {
  [key: string]: unknown
}

// ==================== 模拟面试相关 ====================

// 面试场景
export type InterviewScene =
  | 'tech'
  | 'behavior'
  | 'hr'
  | 'teaching'
  | 'corporate'
  | 'group'
  | 'defense'
  | 'client'
  | 'pressure'
  | 'public'
  | 'medical'
  | 'media'
  | 'remote'
  | 'system'
  | 'aviation'

// 面试模式
export type InterviewMode = 'text' | 'voice' | 'hybrid'

// 面试状态
export type InterviewStatus = 'ongoing' | 'completed' | 'cancelled'
  | 'preparing'
  | 'starting'
  | 'reviewing'
  | 'report_failed'

// 面试难度
export type InterviewDifficulty = 'junior' | 'mid' | 'senior' | 'mixed'

// 消息角色
export type MessageRole = 'assistant' | 'user'

// 面试会话
export interface Interview {
  id: number
  user_id: number
  scene: InterviewScene
  target_company: string
  target_position: string
  target_jd: string
  resume_id: number
  resume_version_id: number
  resume_snapshot: string
  resume_name: string
  examiner_style: string
  training_focus: string
  difficulty: InterviewDifficulty | string
  total_questions: number
  mode: InterviewMode
  status: InterviewStatus
  status_message: string
  current_question_no: number
  started_at: string | null
  ended_at: string | null
  created_at: string
  updated_at: string
}

// 创建面试请求
export interface CreateInterviewRequest {
  scene: InterviewScene
  target_company?: string
  target_position: string
  target_jd: string
  resume_id: number
  version_id?: number
  difficulty?: InterviewDifficulty
  total_questions?: number
  mode?: InterviewMode
  examiner_style?: string
  training_focus?: string
}

// 面试发送简历请求
export interface AttachResumeRequest {
  resume_id: number
  version_id?: number
}

// 面试消息
export interface InterviewMessage {
  id: number
  interview_id: number
  role: MessageRole
  content: string
  audio_url: string
  input_mode: InterviewMode | string
  question_type: string
  question_no: number
  duration_sec: number
  created_at: string
}

// 面试详情响应（GET /api/v1/interviews/:id）
export interface InterviewDetail {
  interview: Interview
  messages: InterviewMessage[]
}

// 面试评分维度
export type InterviewDimension =
  | 'professional'
  | 'expression'
  | 'logic'
  | 'adaptability'
  | 'pace'

// 面试评分明细
export interface InterviewScore {
  id: number
  interview_id: number
  dimension: InterviewDimension | string
  score: number
  comment: string
  created_at: string
}

// 面试复盘报告
export interface InterviewReport {
  id: number
  interview_id: number
  overall_score: number
  summary: string
  highlights: string // JSON 数组字符串
  improvements: string // JSON 数组字符串
  recommendations: string // JSON 数组字符串
  word_cloud: string // JSON 对象字符串
  question_feedback: string // JSON 数组字符串，每题的逐题评价
  created_at: string
}

// 面试复盘 - 逐题评价项
export interface QuestionFeedbackItem {
  question_no: number
  question: string
  answer: string
  score: number
  comment: string
  suggestion: string
}

// ==================== SSE 事件相关 ====================

// SSE 事件类型
export type SSEEventType =
  | 'delta'
  | 'status'
  | 'done'
  | 'started'
  | 'interview_ended'
  | 'error'

// SSE 事件基础结构
export interface SSEEvent {
  type: SSEEventType
  content?: string
  message?: string | InterviewMessage | CopilotMessage
  interview?: Interview
  version?: ResumeVersion
  result?: unknown
  proposals?: CopilotProposal[]
  memory_summary?: string
}

// SSE 回调接口
export interface SSECallbacks {
  onDelta?: (delta: string) => void
  onStatus?: (message: string) => void
  onDone?: (data: { message?: InterviewMessage; version?: ResumeVersion }) => void
  onStarted?: (data: { message?: InterviewMessage; interview?: Interview }) => void
  onInterviewEnded?: (data: {
    message?: InterviewMessage
    interview?: Interview
  }) => void
  onError?: (message: string) => void
  onCopilotDone?: (data: {
    message?: CopilotMessage
    result?: CopilotResponse
    proposals?: CopilotProposal[]
    memory_summary?: string
  }) => void
}

// ==================== 投递看板相关 ====================

// 投递状态
export type DeliveryStatus =
  | 'pending'
  | 'written_test'
  | 'interview'
  | 'waiting_offer'
  | 'offer'
  | 'rejected'

// 投递渠道
export type DeliveryChannel =
  | 'boss'
  | 'official'
  | 'referral'
  | 'campus'
  | 'headhunt'
  | 'other'

// 投递优先级
export type DeliveryPriority = 'high' | 'medium' | 'low'

// 面试轮次类型
export type DeliveryRoundType =
  | 'written_test'
  | 'first_tech'
  | 'second_tech'
  | 'third_tech'
  | 'cross'
  | 'hr'
  | 'additional'
  | 'final'

// 面试形式
export type DeliveryRoundFormat = 'onsite' | 'video' | 'phone'

// 轮次结果
export type DeliveryRoundResult = 'pass' | 'pending' | 'rejected'

// 投递记录主表
export interface Delivery {
  id: number
  user_id: number
  company: string
  position: string
  channel: DeliveryChannel | string
  apply_date: string // YYYY-MM-DD
  status: DeliveryStatus
  priority: DeliveryPriority | string
  jd_text: string
  resume_version_id: number
  hr_contact: string // JSON: {name, wechat, phone, email}
  next_step: string // JSON: {todo, deadline}
  offer_detail: string // JSON: {salary_base, annual_bonus, stock, benefits, deadline}
  remark: string
  created_at: string
  updated_at: string
}

// 面试轮次记录
export interface DeliveryRound {
  id: number
  delivery_id: number
  round_type: DeliveryRoundType | string
  interview_time: string | null
  format: DeliveryRoundFormat | string
  interviewer_name: string
  interviewer_title: string
  question_summary: string
  feedback: string
  result: DeliveryRoundResult | string
  created_at: string
  updated_at: string
}

// HR 反馈记录
export interface DeliveryFeedback {
  id: number
  delivery_id: number
  contact_time: string // YYYY-MM-DD HH:MM
  method: string // wechat/phone/email
  summary: string
  next_action: string
  created_at: string
}

// 投递详情响应（GET /api/v1/deliveries/:id）
export interface DeliveryDetail {
  delivery: Delivery
  rounds: DeliveryRound[]
  feedbacks: DeliveryFeedback[]
}

// 创建投递请求
export interface CreateDeliveryRequest {
  company: string
  position: string
  channel: DeliveryChannel
  apply_date: string
  priority?: DeliveryPriority
  jd_text?: string
  resume_version_id?: number
  hr_contact?: string
  next_step?: string
  remark?: string
}

// 变更状态请求
export interface ChangeStatusRequest {
  status: DeliveryStatus
}

// 创建轮次请求（直接使用 DeliveryRound 结构）
export type CreateRoundRequest = Partial<DeliveryRound> & {
  round_type: DeliveryRoundType | string
}

// 创建反馈请求（直接使用 DeliveryFeedback 结构）
export type CreateFeedbackRequest = Partial<DeliveryFeedback> & {
  contact_time: string
}

// 投递统计
export interface DeliveryStats {
  total: number
  in_progress: number
  offer_count: number
  rejected: number
}

// 投递漏斗
export interface DeliveryFunnel {
  applied: number
  written_test_pass: number
  first_pass: number
  second_pass: number
  offer_count: number
  written_test_rate: number
  first_rate: number
  second_rate: number
  offer_rate: number
}

// 投递列表查询参数
export interface DeliveryListParams {
  status?: string
  channel?: string
}

// ==================== 管理端相关 ====================

// 管理端用户列表项（同 User，含 status）
export type AdminUser = User

// 管理端用户详情
export interface AdminUserDetail extends User {
  resume_count: number
  delivery_count: number
  interview_count: number
}

// 管理端用户列表响应
export interface AdminUserListResponse {
  total: number
  list: AdminUser[]
}

// 管理端用户列表查询参数
export interface AdminUserListParams extends PaginationParams {
  keyword?: string
  status?: string
}

// 管理端仪表盘统计
export interface AdminDashboardStats {
  total_users: number
  active_users: number
  disabled_users: number
  total_resumes: number
  total_deliveries: number
  offer_count: number
  rejected_count: number
  total_interviews: number
}

// 管理端投递列表响应
export interface AdminDeliveryListResponse {
  total: number
  list: Delivery[]
}

// 管理端投递列表查询参数
export interface AdminDeliveryListParams extends PaginationParams {
  status?: string
  company?: string
  user_email?: string
}

// 管理端投递漏斗（同用户漏斗结构）
export type AdminDeliveryFunnel = DeliveryFunnel

// 切换用户状态请求
export interface ToggleUserStatusRequest {
  status: UserStatus
}

// 重置密码请求
export interface ResetPasswordRequest {
  new_password: string
}
