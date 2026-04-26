import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, logout as logoutApi } from '@/api/auth'
import type { LoginRequest } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<{ userId: number; username: string } | null>(null)

  // Getters
  const isLoggedIn = computed(() => !!token.value)

  // Actions
  const login = async (credentials: LoginRequest) => {
    const response = await loginApi(credentials)
    const data = response.data.data
    token.value = data.token
    localStorage.setItem('token', data.token)
    return data
  }

  const logout = async () => {
    try {
      await logoutApi()
    } finally {
      token.value = ''
      userInfo.value = null
      localStorage.removeItem('token')
    }
  }

  const setUserInfo = (info: { userId: number; username: string }) => {
    userInfo.value = info
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    login,
    logout,
    setUserInfo,
  }
})
