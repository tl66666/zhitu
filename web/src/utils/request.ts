import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { message } from 'ant-design-vue'
import { useAuthStore } from '@/stores/auth'
import { useAdminAuthStore } from '@/stores/admin'
import router from '@/router'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  // 开发环境走 vite 代理（/api → :8080），生产环境由 .env 注入绝对 URL
  baseURL: import.meta.env.DEV ? '' : (import.meta.env.VITE_API_BASE_URL || ''),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// isAdminRequest 判断是否为管理端请求（/api/admin/*）
// 管理端请求注入管理员 token，其余注入用户 token
function isAdminRequest(url: string | undefined): boolean {
  return !!url && url.includes('/api/admin/')
}

// 请求拦截器：按请求归属注入对应身份的 token
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const url = config.url
    if (isAdminRequest(url)) {
      // 管理端请求 → 管理员 token
      const adminStore = useAdminAuthStore()
      if (adminStore.adminToken) {
        config.headers.Authorization = `Bearer ${adminStore.adminToken}`
      }
    } else {
      // 普通请求 → 用户 token
      const authStore = useAuthStore()
      if (authStore.token) {
        config.headers.Authorization = `Bearer ${authStore.token}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：统一错误处理，401 按请求归属跳转对应登录页
request.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error) => {
    if (error.response) {
      const { status, config, data } = error.response

      // 401 未授权，按请求归属清理对应身份并跳转
      if (status === 401) {
        if (isAdminRequest(config?.url)) {
          const adminStore = useAdminAuthStore()
          adminStore.adminToken = ''
          adminStore.adminUser = null
          router.push('/admin/login')
        } else {
          const authStore = useAuthStore()
          authStore.logout()
          router.push('/')
        }
        message.error('登录已过期，请重新登录')
      } else if (status === 403) {
        message.error('没有权限访问该资源')
      } else if (status === 404) {
        message.error('请求的资源不存在')
      } else if (status === 500) {
        message.error('服务器内部错误')
      } else {
        const errorMsg = data?.message || data?.error || '请求失败'
        message.error(errorMsg)
      }
    } else if (error.request) {
      message.error('网络错误，请检查网络连接')
    } else {
      message.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

// 封装 GET 请求
export const get = <T>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return request.get<T>(url, config)
}

// 封装 POST 请求
export const post = <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return request.post<T>(url, data, config)
}

// 封装 PUT 请求
export const put = <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return request.put<T>(url, data, config)
}

// 封装 PATCH 请求
export const patch = <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return request.patch<T>(url, data, config)
}

// 封装 DELETE 请求
export const del = <T>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<T>> => {
  return request.delete<T>(url, config)
}

export default request
