import request from './request'
import type { ApiResponse, LoginRequest, LoginResponse } from '@/types/api'

export const login = (data: LoginRequest) => {
  return request.post<ApiResponse<LoginResponse>>('/auth/login', data)
}

export const logout = () => {
  return request.post<ApiResponse<null>>('/auth/logout')
}
