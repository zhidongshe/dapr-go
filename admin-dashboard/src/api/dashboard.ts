import request from './request'
import type { ApiResponse, DashboardStats } from '@/types/api'

export const getDashboardStats = () => {
  return request.get<ApiResponse<DashboardStats>>('/dashboard/stats')
}
