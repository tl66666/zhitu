import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, LoginRequest, RegisterRequest, ChangePasswordRequest } from '@/types/models'
import * as authApi from '@/api/auth'
import { message } from 'ant-design-vue'
import router from '@/router'

// 普通用户认证 store，持久化到 zhitu-auth 键名
// 管理员认证见 stores/admin.ts，两者完全隔离
export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref<string>('')
  const user = ref<User | null>(null)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)

  // 登录
  const login = async (credentials: LoginRequest) => {
    try {
      const response = await authApi.login(credentials)
      // 后端统一响应：{ code, message, data: { token, token_type, expires_in } }
      const authToken = response.data?.data?.token
      if (!authToken) {
        message.error('登录失败：未返回 token')
        return false
      }

      token.value = authToken
      // 后端 login 不返回 user 详情，用邮箱占位，后续可调 /api/auth/me 拉取
      user.value = {
        id: 0,
        email: credentials.email,
        nickname: credentials.email.split('@')[0],
        avatar: '',
        created_at: '',
        updated_at: '',
      }

      // 异步拉取真实用户信息（不阻塞跳转）
      fetchUser()

      message.success('登录成功')
      router.push('/app')

      return true
    } catch (error) {
      console.error('登录失败:', error)
      return false
    }
  }

  // 注册
  const register = async (data: RegisterRequest) => {
    try {
      const response = await authApi.register(data)
      const authToken = response.data?.data?.token
      if (!authToken) {
        message.error('注册失败：未返回 token')
        return false
      }

      token.value = authToken
      user.value = {
        id: 0,
        email: data.email,
        nickname: data.nickname || data.email.split('@')[0],
        avatar: '',
      }
      await fetchUser()
      message.success('注册成功')
      router.push('/app')
      return true
    } catch (error) {
      console.error('注册失败:', error)
      return false
    }
  }

  // 获取当前用户信息
  const fetchUser = async () => {
    try {
      const response = await authApi.getCurrentUser()
      const u = response.data?.data
      if (u) user.value = u
      return true
    } catch (error) {
      console.error('获取用户信息失败:', error)
      return false
    }
  }

  // 修改密码
  const changePassword = async (data: ChangePasswordRequest) => {
    try {
      await authApi.changePassword(data)
      message.success('密码修改成功')
      return true
    } catch (error) {
      console.error('修改密码失败:', error)
      return false
    }
  }

  // 退出登录
  const logout = () => {
    token.value = ''
    user.value = null
    router.push('/')
    message.success('已退出登录')
  }

  return {
    // 状态
    token,
    user,
    isAuthenticated,
    // 操作
    login,
    register,
    fetchUser,
    changePassword,
    logout,
  }
}, {
  // 持久化配置
  persist: {
    key: 'zhitu-auth',
    paths: ['token', 'user'],
  },
})
