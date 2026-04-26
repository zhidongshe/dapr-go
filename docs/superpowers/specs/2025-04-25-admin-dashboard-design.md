# 电商管理后台设计文档

## 1. 项目概述

为现有的微服务架构（Order/Payment/Inventory）增加一个统一的 Vue 管理后台，提供数据查看和分析功能。

### 1.1 功能范围
- **首页看板**：订单、支付、库存统计数据展示
- **订单管理**：订单列表查看、订单详情查看（含商品明细）、搜索筛选、导出
- **库存概览**：库存列表、库存预警提示
- **支付统计**：支付金额、支付方式分布

### 1.2 非功能需求
- 管理后台**仅查看**，不支持订单操作（取消/修改状态）
- JWT 认证
- 响应式设计
- 暗色/亮色主题切换

---

## 2. 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Vue 3 + TypeScript | 框架 |
| 前端 | Element Plus | UI 组件库 |
| 前端 | Pinia | 状态管理 |
| 前端 | Vue Router | 路由 |
| 前端 | Axios | HTTP 请求 |
| 前端 | ECharts | 图表展示 |
| Gateway | Go + Gin | BFF 层 |
| Gateway | JWT-go | 认证 |

---

## 3. 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                     Vue Admin Dashboard                      │
│         (Vue 3 + TypeScript + Element Plus)                 │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP
┌─────────────────────────▼───────────────────────────────────┐
│                    API Gateway (BFF)                         │
│              (Go + Gin + JWT Middleware)                    │
│  ┌─────────────┬─────────────┬─────────────┐               │
│  │   Auth      │   Proxy     │   CORS      │               │
│  │ Middleware  │  Handler    │   Config    │               │
│  └─────────────┴─────────────┴─────────────┘               │
└──────┬──────────────┬──────────────┬────────────────────────┘
       │              │              │
       ▼              ▼              ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Order   │  │ Payment  │  │ Inventory│
│ Service  │  │ Service  │  │ Service  │
│  :8080   │  │  :8081   │  │  :8082   │
└──────────┘  └──────────┘  └──────────┘
```

### 3.1 Gateway 职责
- JWT 认证校验
- 请求路由转发到各服务
- 统一错误处理和日志

### 3.2 现有服务复用
- Order Service: 订单查询接口
- Payment Service: 支付统计接口
- Inventory Service: 库存查询接口

---

## 4. 目录结构

```
dapr-go/
├── admin-dashboard/           # Vue 前端项目
│   ├── src/
│   │   ├── api/              # API 接口封装
│   │   ├── assets/           # 静态资源
│   │   ├── components/       # 公共组件
│   │   ├── layouts/          # 布局组件
│   │   ├── router/           # 路由配置
│   │   ├── stores/           # Pinia 状态管理
│   │   ├── styles/           # 样式文件
│   │   ├── types/            # TypeScript 类型定义
│   │   ├── utils/            # 工具函数
│   │   └── views/            # 页面组件
│   │       ├── login/
│   │       ├── dashboard/
│   │       ├── orders/
│   │       └── inventory/
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── api-gateway/              # Go Gateway 服务
│   ├── main.go
│   ├── go.mod
│   ├── config/
│   │   └── config.go
│   ├── middleware/
│   │   ├── auth.go          # JWT 认证中间件
│   │   ├── cors.go          # 跨域中间件
│   │   └── logger.go        # 日志中间件
│   ├── handlers/
│   │   ├── auth_handler.go  # 认证接口
│   │   ├── order_handler.go # 订单接口
│   │   ├── payment_handler.go # 支付接口
│   │   ├── inventory_handler.go # 库存接口
│   │   └── dashboard_handler.go # 看板接口
│   ├── client/
│   │   └── service_client.go # 微服务调用客户端
│   └── utils/
│       ├── jwt.go           # JWT 工具
│       └── response.go      # 响应封装
│
├── order-service/            # 现有订单服务
├── payment-service/          # 现有支付服务
└── inventory-service/        # 现有库存服务
```

---

## 5. API 设计

### 5.1 认证接口

#### POST /api/auth/login
用户登录，返回 JWT Token。

**Request:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
}
```

#### POST /api/auth/logout
用户登出。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 5.2 看板接口

#### GET /api/dashboard/stats
获取首页统计数据。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "orders": {
      "total": 1000,
      "pending": 50,
      "paid": 200,
      "processing": 30,
      "shipped": 100,
      "completed": 590,
      "cancelled": 30
    },
    "payments": {
      "todayAmount": 15000.00,
      "todayCount": 45,
      "weekAmount": 98000.00,
      "monthAmount": 450000.00
    },
    "inventory": {
      "totalProducts": 500,
      "warningCount": 10
    }
  }
}
```

### 5.3 订单接口

#### GET /api/orders
获取订单列表。

**Query Parameters:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10，最大 100 |
| orderNo | string | 否 | 订单号筛选 |
| status | int | 否 | 订单状态筛选 |
| payStatus | int | 否 | 支付状态筛选 |
| startTime | string | 否 | 开始时间，格式: 2025-04-01 |
| endTime | string | 否 | 结束时间，格式: 2025-04-25 |

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "ORD20250425000001",
        "user_id": 100,
        "total_amount": 299.99,
        "status": 1,
        "pay_status": 1,
        "created_at": "2025-04-25T10:00:00Z"
      }
    ],
    "total": 1000,
    "page": 1,
    "pageSize": 10
  }
}
```

#### GET /api/orders/:id
获取订单详情。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "ORD20250425000001",
    "user_id": 100,
    "total_amount": 599.97,
    "status": 1,
    "pay_status": 1,
    "pay_time": "2025-04-25T10:05:00Z",
    "pay_method": "alipay",
    "remark": "请尽快发货",
    "created_at": "2025-04-25T10:00:00Z",
    "updated_at": "2025-04-25T10:05:00Z",
    "items": [
      {
        "id": 1,
        "product_id": 101,
        "product_name": "iPhone 15 手机壳",
        "unit_price": 199.99,
        "quantity": 3,
        "total_price": 599.97,
        "created_at": "2025-04-25T10:00:00Z"
      }
    ]
  }
}
```

#### GET /api/orders/export
导出订单数据（CSV/Excel）。

**Query Parameters:**
- 同列表接口

**Response:** 文件下载

### 5.4 支付接口

#### GET /api/payments/stats
获取支付统计数据。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "todayAmount": 15000.00,
    "todayCount": 45,
    "methodDistribution": [
      { "method": "alipay", "amount": 9000.00, "count": 25 },
      { "method": "wechat", "amount": 6000.00, "count": 20 }
    ],
    "trend": [
      { "date": "2025-04-19", "amount": 12000.00 },
      { "date": "2025-04-20", "amount": 15000.00 }
    ]
  }
}
```

### 5.5 库存接口

#### GET /api/inventory/stats
获取库存统计数据。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "totalProducts": 500,
    "totalStock": 5000,
    "warningCount": 10,
    "outOfStockCount": 2
  }
}
```

#### GET /api/inventory/warnings
获取库存预警列表。

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "product_id": 101,
      "product_name": "iPhone 15 手机壳",
      "current_stock": 5,
      "warning_threshold": 10
    }
  ]
}
```

#### GET /api/inventory
获取库存列表。

**Query Parameters:**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 10 |
| productName | string | 否 | 商品名称筛选 |
| lowStock | bool | 否 | 是否只显示低库存 |

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "product_id": 101,
        "product_name": "iPhone 15 手机壳",
        "stock": 50,
        "reserved": 10,
        "available": 40,
        "warning_threshold": 10
      }
    ],
    "total": 500,
    "page": 1,
    "pageSize": 10
  }
}
```

---

## 6. 前端设计

### 6.1 页面路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/login` | 登录页 | 账号密码登录 |
| `/` | 首页看板 | 数据统计卡片 + 图表 |
| `/orders` | 订单列表 | 表格 + 搜索 + 分页 |
| `/orders/:id` | 订单详情 | 订单信息 + 商品明细 |
| `/inventory` | 库存概览 | 库存列表 + 预警 |

### 6.2 布局设计

**整体布局:**
- 左侧：固定侧边栏（可折叠），包含菜单导航
- 顶部：Header（面包屑 + 用户信息 + 主题切换按钮）
- 内容区：白色卡片背景，可滚动

**侧边栏菜单:**
```
- 首页看板
- 订单管理
  - 订单列表
- 库存管理
  - 库存概览
```

### 6.3 页面详情

#### 登录页
- 居中卡片布局
- 账号、密码输入框
- 登录按钮
- 背景：渐变色或简洁图案

#### 首页看板
- **统计卡片区**: 4-6 个卡片横向排列
  - 今日订单数
  - 今日销售额
  - 待处理订单
  - 库存预警数
- **图表区**:
  - 近 7 天销售趋势（折线图）
  - 订单状态分布（饼图）
  - 支付方式分布（柱状图）

#### 订单列表页
- 搜索栏：订单号、状态、时间范围
- 操作按钮：导出 Excel
- 数据表格：列包括订单号、用户ID、金额、状态、支付状态、创建时间、操作（查看详情）
- 分页器

#### 订单详情页
- 上部卡片：订单基本信息
  - 订单号、用户ID
  - 订单状态（带颜色标签）
  - 支付状态、支付方式、支付时间
  - 订单金额、创建时间
- 中部卡片：商品明细表格
  - 商品ID、名称、单价、数量、总价
- 下部：订单时间线（可选）

#### 库存概览页
- 统计卡片：总商品数、库存预警数、缺货数
- 库存表格：商品信息 + 库存数量
- 预警提示区域

---

## 7. 数据模型

### 7.1 前端 TypeScript 类型

```typescript
// 通用响应类型
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 分页响应
interface PaginatedResponse<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}

// 订单
interface Order {
  id: number;
  orderNo: string;
  userId: number;
  totalAmount: number;
  status: OrderStatus;
  payStatus: PayStatus;
  payTime?: string;
  payMethod?: string;
  remark?: string;
  createdAt: string;
  updatedAt: string;
  items?: OrderItem[];
}

interface OrderItem {
  id: number;
  orderId: number;
  productId: number;
  productName: string;
  unitPrice: number;
  quantity: number;
  totalPrice: number;
  createdAt: string;
}

enum OrderStatus {
  Pending = 0,
  Paid = 1,
  Processing = 2,
  Shipped = 3,
  Completed = 4,
  Cancelled = 5
}

enum PayStatus {
  Unpaid = 0,
  Paid = 1,
  Failed = 2,
  Refunded = 3
}

// 看板统计
interface DashboardStats {
  orders: {
    total: number;
    pending: number;
    paid: number;
    processing: number;
    shipped: number;
    completed: number;
    cancelled: number;
  };
  payments: {
    todayAmount: number;
    todayCount: number;
    weekAmount: number;
    monthAmount: number;
  };
  inventory: {
    totalProducts: number;
    warningCount: number;
  };
}

// 库存
interface Inventory {
  productId: number;
  productName: string;
  stock: number;
  reserved: number;
  available: number;
  warningThreshold: number;
}

interface InventoryWarning {
  productId: number;
  productName: string;
  currentStock: number;
  warningThreshold: number;
}
```

---

## 8. 错误处理

### 8.1 HTTP 状态码

| 状态码 | 含义 | 处理 |
|--------|------|------|
| 200 | 成功 | 正常处理 |
| 400 | 请求参数错误 | 显示具体错误信息 |
| 401 | 未认证 | 跳转登录页 |
| 403 | 无权限 | 显示无权限提示 |
| 404 | 资源不存在 | 显示 404 页面 |
| 500 | 服务器错误 | 显示通用错误提示 |
| 503 | 服务不可用 | 显示服务不可用提示 |

### 8.2 业务错误码

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 资源不存在 |
| 2001 | 认证失败 |
| 2002 | Token 过期 |
| 5000 | 服务器内部错误 |

### 8.3 错误处理流程

**前端:**
1. Axios 拦截器统一处理 HTTP 错误
2. 401 自动跳转登录页
3. 业务错误显示 Message 提示
4. 网络错误显示重试按钮

**Gateway:**
1. 统一错误响应格式
2. 记录错误日志
3. 服务不可用时返回 503

---

## 9. 安全考虑

1. **JWT 认证**: 所有接口（除登录）需验证 Token
2. **Token 刷新**: 支持 Token 自动刷新机制
3. **HTTPS**: 生产环境强制 HTTPS
4. **密码加密**: 登录密码 bcrypt 加密存储
5. **CORS**: Gateway 配置跨域白名单
6. **请求限流**: Gateway 层限流防护

---

## 10. 开发计划

### Phase 1: Gateway 基础
- Gateway 项目初始化
- JWT 认证中间件
- 服务发现/代理

### Phase 2: 前端基础
- Vue 项目初始化
- 登录页面
- 布局框架
- Axios 封装

### Phase 3: 看板页面
- Gateway 看板接口
- 首页统计卡片
- 图表组件

### Phase 4: 订单管理
- Gateway 订单接口
- 订单列表页
- 订单详情页
- 导出功能

### Phase 5: 库存管理
- Gateway 库存接口
- 库存概览页

---

## 11. 附录

### 11.1 订单状态映射

| 值 | 状态 | 颜色 |
|----|------|------|
| 0 | 待支付 | orange |
| 1 | 已支付 | blue |
| 2 | 处理中 | cyan |
| 3 | 已发货 | purple |
| 4 | 已完成 | green |
| 5 | 已取消 | red |

### 11.2 支付状态映射

| 值 | 状态 | 颜色 |
|----|------|------|
| 0 | 未支付 | default |
| 1 | 已支付 | success |
| 2 | 支付失败 | danger |
| 3 | 已退款 | warning |
