<template src="./AppLayout.template.html"></template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import UserProfileModal from '@/components/UserProfileModal.vue'
import {
  UserOutlined,
  FileTextOutlined,
  CommentOutlined,
  SendOutlined,
  DownOutlined,
  LogoutOutlined,
  LockOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BulbOutlined,
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 侧边栏折叠状态
const collapsed = ref(false)
const autoCollapsedForViewport = ref(false)

// 个人资料弹窗
const showProfileModal = ref(false)

const syncSidebarToViewport = () => {
  const shouldCollapse = window.innerWidth <= 1100
  if (shouldCollapse === autoCollapsedForViewport.value) return
  autoCollapsedForViewport.value = shouldCollapse
  collapsed.value = shouldCollapse
}

onMounted(() => {
  syncSidebarToViewport()
  window.addEventListener('resize', syncSidebarToViewport)
})

onBeforeUnmount(() => window.removeEventListener('resize', syncSidebarToViewport))

// 菜单选中状态
const selectedKeys = ref<string[]>(['resumes'])
const openKeys = ref<string[]>([])

// 当前路由名称
const currentRouteName = computed(() => route.name as string)

// 当前路由标题
const currentRouteTitle = computed(() => {
  const titleMap: Record<string, string> = {
    ResumeList: '简历实验室',
    ResumeTemplateSelect: '选择简历模板',
    ResumeEditor: '简历实验室',
    Copilot: '求职 Copilot',
    InterviewList: '面试训练场',
    InterviewSceneSelect: '选择训练场景',
    InterviewRoom: '面试训练场',
    DeliveryKanban: '投递看板',
    ChangePassword: '修改密码',
  }
  return titleMap[currentRouteName.value] || ''
})

// 用户名首字母
const userInitial = computed(() => {
  const nickname = authStore.user?.nickname || '用户'
  return nickname.charAt(0).toUpperCase()
})

// 监听路由变化，更新菜单选中状态
watch(
  () => route.path,
  (path) => {
    if (path === '/app' || path === '/app/') {
      selectedKeys.value = ['resumes']
    } else {
      // 形如 /app/profile -> 取第三段
      const seg = path.split('/')[2]
      if (seg) selectedKeys.value = [seg]
    }
  },
  { immediate: true }
)

// 导航到指定路径
const navigateTo = (path: string) => {
  router.push(path)
}

// 退出登录
const handleLogout = () => {
  authStore.logout()
}
</script>

<style scoped src="./styles/app-layout.css"></style>
<style scoped src="./styles/app-layout-theme-bridge.css"></style>
