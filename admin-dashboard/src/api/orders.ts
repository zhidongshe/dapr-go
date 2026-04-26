import request from './request'
import type { ApiResponse, Order, PaginatedResponse } from '@/types/api'

export interface OrderListParams {
  page?: number
  pageSize?: number
  orderNo?: string
  status?: number
  payStatus?: number
  startTime?: string
  endTime?: string
}

export const getOrders = (params?: OrderListParams) => {
  return request.get<ApiResponse<PaginatedResponse<Order>>>('/orders', { params })
}

export const getOrderDetail = (id: number | string) => {
  return request.get<ApiResponse<Order>>(`/orders/${id}`)
}
