import request from './request'
import type { ApiResponse, Product, PaginatedResponse } from '@/types/api'

export interface ProductListParams {
  product_name?: string
  status?: number
  page?: number
  pageSize?: number
}

export const getProducts = (params?: ProductListParams) =>
  request.get<ApiResponse<PaginatedResponse<Product>>>('/products', { params })

export const createProduct = (data: Partial<Product>) =>
  request.post<ApiResponse<Product>>('/products', data)

export const updateProductPrice = (id: number, original_price: number) =>
  request.put<ApiResponse<Product>>(`/products/${id}/price`, { original_price })

export const updateProductStatus = (id: number, status: number) =>
  request.put<ApiResponse<Product>>(`/products/${id}/status`, { status })
