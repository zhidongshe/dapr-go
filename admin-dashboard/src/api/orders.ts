import request from './request'
import type { ApiResponse, Order, PaginatedResponse } from '@/types/api'

export interface OrderListParams {
  page?: number
  pageSize?: number
  order_no?: string
  status?: number
}

export const getOrders = (params?: OrderListParams) => {
  return request.get<ApiResponse<PaginatedResponse<Order>>>('/orders', { params })
}

export const getOrderDetail = (id: number | string) => {
  return request.get<ApiResponse<Order>>(`/orders/${id}`)
}
