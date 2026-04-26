import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types/api'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse<unknown>

    if (data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }

    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
      ElMessage.error('登录已过期，请重新登录')
    } else {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
