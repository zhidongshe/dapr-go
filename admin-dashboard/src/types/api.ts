// 通用响应类型
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

// 分页响应
export interface PaginatedResponse<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 订单状态
export enum OrderStatus {
  Pending = 0,
  Paid = 1,
  Processing = 2,
  Shipped = 3,
  Completed = 4,
  Cancelled = 5
}

// 支付状态
export enum PayStatus {
  Unpaid = 0,
  Paid = 1,
  Failed = 2,
  Refunded = 3
}

// 订单
export interface Order {
  id: number
  orderNo: string
  userId: number
  totalAmount: number
  status: OrderStatus
  payStatus: PayStatus
  payTime?: string
  payMethod?: string
  remark?: string
  createdAt: string
  updatedAt: string
  items?: OrderItem[]
}

// 订单项
export interface OrderItem {
  id: number
  orderId: number
  productId: number
  productName: string
  unitPrice: number
  quantity: number
  totalPrice: number
  createdAt: string
}

// 看板统计
export interface DashboardStats {
  orders: {
    total: number
    pending: number
    paid: number
    processing: number
    shipped: number
    completed: number
    cancelled: number
  }
  payments: {
    todayAmount: number
    todayCount: number
    weekAmount: number
    monthAmount: number
  }
  inventory: {
    totalProducts: number
    warningCount: number
  }
}

// 登录
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expiresIn: number
  tokenType: string
}

// 用户
export interface UserInfo {
  userId: number
  username: string
}

// 商品状态
export enum ProductStatus {
  OffSale = 0,
  OnSale = 1
}

// 商品
export interface Product {
  product_id: number
  product_name: string
  original_price: number
  status: ProductStatus
  created_at: string
  updated_at: string
}

// 商品列表参数
export interface ProductListParams {
  product_name?: string
  status?: number
  page?: number
  pageSize?: number
}
