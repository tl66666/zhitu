import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAdminAuthStore } from '@/stores/admin'
import AppLayout from '@/components/layout/AppLayout.vue'
import AdminLayout from '@/components/layout/AdminLayout.vue'

// 路由配置
const routes: RouteRecordRaw[] = [
  // 公开首页 - 项目介绍 + 登录注册
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { requiresAuth: false, isPublic: true },
  },
  // 兼容旧路径 - 登录页重定向到首页
  {
    path: '/login',
    name: 'Login',
    redirect: '/',
    meta: { requiresAuth: false },
  },
  // 兼容旧路径 - 注册页重定向到首页
  {
    path: '/register',
    name: 'Register',
    redirect: '/',
    meta: { requiresAuth: false },
  },
  // 管理端登录
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/AdminLogin.vue'),
    meta: { requiresAuth: false, isAdmin: false },
  },
  // 简历实验室公开预览：用于本地设计验收，不依赖后端登录态
  {
    path: '/resume-lab-preview',
    name: 'ResumeLabPreview',
    component: () => import('@/views/ResumeEditor.vue'),
    meta: { requiresAuth: false, isPublic: false },
  },
  // 管理端路由
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAuth: true, isAdmin: true },
    children: [
      // 管理端仪表盘
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/AdminDashboard.vue'),
      },
      // 用户管理
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/AdminUsers.vue'),
      },
      // 投递管理
      {
        path: 'deliveries',
        name: 'AdminDeliveries',
        component: () => import('@/views/admin/AdminDeliveries.vue'),
      },
    ],
  },
  // 用户控制台（需要认证）
  {
    path: '/app',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      // 登录后默认进入简历实验室
      {
        path: '',
        redirect: '/app/resumes',
      },
      // 简历列表
      {
        path: 'resumes',
        name: 'ResumeList',
        component: () => import('@/views/ResumeList.vue'),
      },
      // 创建简历前先选择文本化模板
      {
        path: 'resumes/new',
        name: 'ResumeTemplateSelect',
        component: () => import('@/views/ResumeTemplateSelect.vue'),
      },
      // 简历编辑器
      {
        path: 'resumes/:id',
        name: 'ResumeEditor',
        component: () => import('@/views/ResumeEditor.vue'),
        props: true,
      },
      {
        path: 'copilot',
        name: 'Copilot',
        component: () => import('@/views/Copilot.vue'),
      },
      // 面试列表
      {
        path: 'interviews',
        name: 'InterviewList',
        component: () => import('@/views/InterviewList.vue'),
      },
      {
        path: 'interviews/new',
        name: 'InterviewSceneSelect',
        component: () => import('@/views/InterviewSceneSelect.vue'),
      },
      // 面试房间
      {
        path: 'interviews/:id',
        name: 'InterviewRoom',
        component: () => import('@/views/InterviewRoom.vue'),
        props: true,
      },
      // 投递看板
      {
        path: 'deliveries',
        name: 'DeliveryKanban',
        component: () => import('@/views/DeliveryKanban.vue'),
      },
      // 修改密码
      {
        path: 'settings/password',
        name: 'ChangePassword',
        component: () => import('@/views/Settings/ChangePassword.vue'),
      },
    ],
  },
  // 404 页面
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
  },
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (_to.hash) return { el: _to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const adminStore = useAdminAuthStore()
  const isAuthenticated = authStore.isAuthenticated
  const isAdminAuthenticated = adminStore.isAuthenticated
  const requiresAuth = to.meta.requiresAuth !== false
  const requiresAdmin = to.meta.isAdmin === true
  const isPublic = to.meta.isPublic === true

  // 管理端路由：使用独立的管理员身份校验
  if (requiresAdmin) {
    if (!isAdminAuthenticated) {
      next({
        path: '/admin/login',
        query: { redirect: to.fullPath },
      })
      return
    }
  }

  // 普通用户受保护路由：未登录跳转公开首页
  if (requiresAuth && !requiresAdmin && !isAuthenticated) {
    next({
      path: '/',
      query: to.fullPath !== '/app' ? { redirect: to.fullPath } : {},
    })
    return
  }

  // 已登录用户访问公开首页或旧登录/注册页，跳转控制台
  if (isPublic && isAuthenticated) {
    next({ path: '/app' })
    return
  }

  // 已登录管理员访问管理端登录页，跳转管理后台
  if (!requiresAuth && isAdminAuthenticated && to.path === '/admin/login') {
    next({ path: '/admin' })
    return
  }

  next()
})

export default router
