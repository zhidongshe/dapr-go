# Dapr OMS 订单管理系统 - 项目总览

## 1. 项目简介

这是一个基于 **Dapr (Distributed Application Runtime)** 构建的微服务订单管理系统，演示了如何使用 Dapr 简化微服务开发。

### 核心功能
- 订单管理（创建、查询、列表、取消）
- 支付处理（模拟支付）
- 订单状态流转与事件通知

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        服务层 (Go + Gin)                      │
├──────────────┬──────────────────┬───────────────────────────┤
│              │                  │                           │
│  Order       │   Dapr Sidecar   │   Payment                 │
│  Service     │   (每个服务一个)  │   Service                 │
│  :8080       │   :3500/:3501    │   :8081                   │
│              │                  │                           │
└──────┬───────┴────────┬─────────┴───────────┬───────────────┘
       │                │                     │
       │                │  Dapr Pub/Sub       │
       │                │  (Redis)            │
       │                │                     │
       └────────────────┴─────────────────────┘
                │
       ┌────────┴────────┐
       │                 │
   MySQL (状态存储)    Redis (消息队列)
   :3306               :6379
```

---

## 3. 服务说明

### Order Service (订单服务)
**职责**: 订单生命周期管理

**端口**: 8080 (业务), 3500 (Dapr)

**核心功能**:
| API | 方法 | 说明 |
|-----|------|------|
| `/api/v1/orders` | POST | 创建订单 |
| `/api/v1/orders/:id` | GET | 查询订单详情 |
| `/api/v1/orders?user_id=xxx` | GET | 查询用户订单列表 |
| `/api/v1/orders/:id/cancel` | POST | 取消订单 |

**代码结构**:
```
order-service/
├── main.go              # 服务入口
├── handlers/            # HTTP 请求处理器
│   └── order_handler.go
├── services/            # 业务逻辑
│   └── order_service.go
├── models/              # 数据模型
│   ├── order.go
│   └── order_test.go
└── repository/          # 数据访问层
    └── order_repo.go
```

---

### Payment Service (支付服务)
**职责**: 处理支付请求

**端口**: 8081 (业务), 3501 (Dapr)

**核心功能**:
| API | 方法 | 说明 |
|-----|------|------|
| `/api/v1/payments` | POST | 发起支付 |
| `/api/v1/payments/callback` | POST | 支付回调 |

**代码结构**:
```
payment-service/
├── main.go
├── handlers/
│   └── payment_handler.go
├── services/
│   └── payment_service.go
└── models/
    ├── payment.go
    └── payment_test.go
```

---

### Shared (共享模块)
**职责**: 跨服务共享代码

```
shared/
├── dto/                 # 数据传输对象
│   ├── response.go      # 统一响应格式
│   └── response_test.go
└── events/              # 事件定义
    ├── order_events.go
    └── order_events_test.go
```

---

## 4. Dapr 组件配置

### 为什么用 Dapr？
Dapr 是一个微服务运行时，提供服务间通信、状态管理、发布订阅等能力，让开发者专注于业务代码。

### 本项目使用的 Dapr 功能

| 功能 | 组件 | 用途 |
|------|------|------|
| **Pub/Sub** | Redis | 服务间异步通信（订单事件） |

### 组件配置

```yaml
# components/pubsub.yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: order-pubsub
spec:
  type: pubsub.redis
  version: v1
  metadata:
  - name: redisHost
    value: redis:6379
```

---

## 5. 事件驱动架构

### 事件类型

```go
// 订单创建事件
TopicOrderCreated = "order-created"

// 订单支付事件
TopicOrderPaid = "order-paid"

// 订单取消事件
TopicOrderCancelled = "order-cancelled"

// 订单状态变更事件
TopicOrderStatusChanged = "order-status-changed"
```

### 事件流转示例（支付流程）

```
1. 用户调用 Payment Service 支付
         │
         ▼
2. Payment Service 发布 "order-paid" 事件
         │
         ▼
3. Order Service 订阅 "order-paid" 事件
         │
         ▼
4. Order Service 更新订单状态为 "已支付"
```

**代码示例** (Payment Service 发布事件):
```go
event := events.OrderPaidEvent{
    OrderID:   order.ID,
    OrderNo:   order.OrderNo,
    PayMethod: req.PayMethod,
    PayTime:   time.Now(),
}
s.daprClient.PublishEvent(ctx, "order-pubsub", events.TopicOrderPaid, eventData)
```

**代码示例** (Order Service 订阅事件):
```go
func (h *OrderHandler) DaprSubscribe(c *gin.Context) {
    c.JSON(http.StatusOK, []daprSubscription{
        {
            PubsubName: "order-pubsub",
            Topic:      events.TopicOrderPaid,
            Route:      "/events/order-paid",
        },
    })
}
```

---

## 6. 订单状态流转

```
┌──────────┐    支付成功     ┌──────────┐
│  待支付   │ ──────────────► │  已支付   │
│   (0)    │                 │   (1)    │
└────┬─────┘                 └────┬─────┘
     │                            │
     │ 取消订单                    │ 发货
     ▼                            ▼
┌──────────┐                 ┌──────────┐
│  已取消   │                 │  已发货   │
│   (5)    │                 │   (3)    │
└──────────┘                 └────┬─────┘
                                  │ 收货
                                  ▼
                            ┌──────────┐
                            │  已完成   │
                            │   (4)    │
                            └──────────┘
```

**状态码**:
```go
OrderStatusPending    = 0  // 待支付
OrderStatusPaid       = 1  // 已支付
OrderStatusProcessing = 2  // 处理中
OrderStatusShipped    = 3  // 已发货
OrderStatusCompleted  = 4  // 已完成
OrderStatusCancelled  = 5  // 已取消
```

---

## 7. 数据存储

### MySQL 表结构

**orders 表**:
```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(32) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status TINYINT DEFAULT 0,
    pay_status TINYINT DEFAULT 0,
    pay_time DATETIME,
    pay_method VARCHAR(20),
    remark TEXT,
    created_at DATETIME,
    updated_at DATETIME
);
```

**order_items 表**:
```sql
CREATE TABLE order_items (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    created_at DATETIME
);
```

---

## 8. 快速开始

### 启动所有服务
```bash
make up
```

### 查看日志
```bash
make logs
```

### 停止服务
```bash
make down
```

### 运行测试
```bash
make test
```

---

## 9. API 测试示例

### 1. 创建订单
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1001,
    "items": [
      {
        "product_id": 1,
        "product_name": "iPhone 16",
        "unit_price": 5999,
        "quantity": 1
      }
    ]
  }'
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": 1,
    "order_no": "ORD202504220001",
    "total_amount": 5999,
    "status": 0
  }
}
```

### 2. 支付订单
```bash
curl -X POST http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_no": "ORD202504220001",
    "pay_method": "alipay"
  }'
```

### 3. 查询订单
```bash
curl http://localhost:8080/api/v1/orders/1
```

### 4. 取消订单
```bash
curl -X POST http://localhost:8080/api/v1/orders/1/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "测试取消"}'
```

---

## 10. 项目目录结构

```
dapr_go/
├── README.md                    # 项目说明
├── PROJECT_OVERVIEW.md          # 本文档
├── Makefile                     # 常用命令
├── docker-compose.yml           # 服务编排
├── start-local.sh               # 本地启动脚本
├── stop-local.sh                # 本地停止脚本
│
├── components/                  # Dapr 组件配置 (Docker)
│   ├── pubsub.yaml             # Redis Pub/Sub 配置
│   └── statestore.yaml         # MySQL 状态存储配置
│
├── components-local/           # Dapr 组件配置 (本地开发)
│   ├── pubsub-local.yaml
│   └── statestore-local.yaml
│
├── order-service/              # 订单服务
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── handlers/
│   ├── services/
│   ├── models/
│   └── repository/
│
├── payment-service/           # 支付服务
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── handlers/
│   ├── services/
│   └── models/
│
├── shared/                    # 共享模块
│   ├── dto/
│   └── events/
│
├── scripts/                   # 初始化脚本
│   └── init-db.sql           # MySQL 初始化
│
└── docs/                      # 文档
    └── superpowers/
        ├── plans/            # 开发计划
        └── specs/            # 设计文档
```

---

## 11. 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.26 |
| Web 框架 | Gin |
| 微服务运行时 | Dapr 1.12 |
| 数据库 | MySQL 8.0 |
| 消息队列 | Redis 7 |
| 容器化 | Docker & Docker Compose |

---

## 12. 学习路径

### 第1步：理解 Dapr 基础
- Dapr Sidecar 模式
- Pub/Sub 发布订阅
- Service Invocation 服务调用

### 第2步：阅读代码
1. 先看 `shared/events/` 理解事件定义
2. 再看 `order-service/services/` 理解业务逻辑
3. 最后看 `payment-service/services/` 理解支付流程

### 第3步：动手实验
1. 启动服务：`make up`
2. 创建订单 → 支付 → 查询，观察状态变化
3. 查看日志理解事件流转

---

## 13. 常见问题

**Q: Dapr Sidecar 是什么？**
A: 每个服务旁边运行的代理，处理服务发现、消息传递、状态管理等，业务代码只需调用 Dapr SDK。

**Q: 为什么用 Redis 做 Pub/Sub？**
A: 轻量级、性能好，适合开发和测试环境。生产环境可替换为 Kafka、RabbitMQ 等。

**Q: 服务间如何通信？**
A: Payment Service 通过 HTTP 调用 Order Service 查询订单，通过 Dapr Pub/Sub 发送支付事件。

---

**项目地址**: https://github.com/zhidongshe/dapr-go
