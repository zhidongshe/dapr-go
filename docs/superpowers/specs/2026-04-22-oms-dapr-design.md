# OMS订单管理系统 - Dapr架构设计文档

## 1. 概述

### 1.1 项目目标

构建一个基于Dapr的订单管理系统(OMS)，实现订单创建、支付和状态管理，通过消息队列将订单状态变化通知下游系统。

### 1.2 技术栈

- **语言**: Go 1.21+
- **微服务框架**: Dapr (Sidecar模式)
- **状态存储**: MySQL 8.0
- **消息队列**: Redis Pub/Sub
- **部署**: Docker Compose
- **API风格**: RESTful HTTP

## 2. 架构设计

### 2.1 服务拆分

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                           │
│                   (Nginx / Traefik / 直接访问)               │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┴─────────────────────┐
        ▼                                           ▼
┌───────────────┐                           ┌───────────────┐
│ Order Service │◄─────────────────────────►│Payment Service│
│   (订单服务)   │     Dapr Service Invocation│   (支付服务)   │
└───────┬───────┘                           └───────┬───────┘
        │                                           │
        │ Dapr Sidecar                        Dapr Sidecar
        │                                           │
        └─────────────────────┬─────────────────────┘
                              │
                    ┌─────────┴──────────┐
                    │   Dapr Runtime     │
                    │    Components      │
                    └─────────┬──────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    ┌──────────┐      ┌──────────┐         ┌──────────┐
    │  MySQL   │      │  Redis   │         │ (Other   │
    │ (State)  │      │(Pub/Sub) │         │ Consumers)│
    └──────────┘      └──────────┘         └──────────┘
```

### 2.2 服务职责

| 服务              | 职责                   | 端口                       |
| --------------- | -------------------- | ------------------------ |
| Order Service   | 订单CRUD、状态管理、发布状态变更事件 | 8080 (app) / 3500 (dapr) |
| Payment Service | 处理支付请求、更新订单支付状态      | 8081 (app) / 3501 (dapr) |

## 3. 数据模型

### 3.1 订单表 (orders)

```sql
CREATE TABLE orders (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no        VARCHAR(32) NOT NULL UNIQUE COMMENT '订单编号',
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    total_amount    DECIMAL(12,2) NOT NULL COMMENT '订单总金额',
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '订单状态',
    pay_status      TINYINT NOT NULL DEFAULT 0 COMMENT '支付状态',
    pay_time        DATETIME NULL COMMENT '支付时间',
    pay_method      VARCHAR(20) NULL COMMENT '支付方式',
    remark          VARCHAR(500) NULL COMMENT '备注',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

CREATE TABLE order_items (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id        BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    product_id      BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    product_name    VARCHAR(200) NOT NULL COMMENT '商品名称',
    unit_price      DECIMAL(10,2) NOT NULL COMMENT '单价',
    quantity        INT UNSIGNED NOT NULL COMMENT '数量',
    total_price     DECIMAL(12,2) NOT NULL COMMENT '小计金额',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    INDEX idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';
```

### 3.2 订单状态定义

```go
const (
    OrderStatusPending    = 0 // 待支付
    OrderStatusPaid       = 1 // 已支付
    OrderStatusProcessing = 2 // 处理中
    OrderStatusShipped    = 3 // 已发货
    OrderStatusCompleted  = 4 // 已完成
    OrderStatusCancelled  = 5 // 已取消
)

const (
    PayStatusUnpaid = 0 // 未支付
    PayStatusPaid   = 1 // 已支付
    PayStatusFailed = 2 // 支付失败
    PayStatusRefunded = 3 // 已退款
)
```

## 4. API设计

### 4.1 Order Service API

| 方法   | 路径                        | 描述     |
| ---- | ------------------------- | ------ |
| POST | /api/v1/orders            | 创建订单   |
| GET  | /api/v1/orders/:id        | 查询订单详情 |
| GET  | /api/v1/orders            | 查询订单列表 |
| POST | /api/v1/orders/:id/cancel | 取消订单   |

### 4.2 Payment Service API

| 方法   | 路径                        | 描述       |
| ---- | ------------------------- | -------- |
| POST | /api/v1/payments          | 发起支付     |
| POST | /api/v1/payments/callback | 支付回调(模拟) |

### 4.3 请求/响应示例

**创建订单：**

```http
POST /api/v1/orders
Content-Type: application/json

{
    "user_id": 10001,
    "items": [
        {
            "product_id": 101,
            "product_name": "iPhone 15",
            "unit_price": 5999.00,
            "quantity": 1
        }
    ],
    "remark": "请尽快发货"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "order_id": 1,
        "order_no": "202504220001",
        "total_amount": 5999.00,
        "status": 0,
        "created_at": "2026-04-22T10:30:00Z"
    }
}
```

**发起支付：**

```http
POST /api/v1/payments
Content-Type: application/json

{
    "order_no": "202504220001",
    "pay_method": "alipay"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "transaction_id": "TXN202504220001",
        "status": "success"
    }
}
```

## 5. Dapr组件配置

### 5.1 State Store (MySQL)

```yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: order-statestore
spec:
  type: state.mysql
  version: v1
  metadata:
  - name: connectionString
    value: "user:password@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
```

### 5.2 Pub/Sub (Redis)

```yaml
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
  - name: redisPassword
    value: ""
```

## 6. 消息事件设计

### 6.1 事件类型

| Topic                | 描述    | 生产者             | 消费者                |
| -------------------- | ----- | --------------- | ------------------ |
| order-created        | 订单创建  | Order Service   | 下游系统               |
| order-paid           | 订单已支付 | Payment Service | 下游系统、Order Service |
| order-cancelled      | 订单取消  | Order Service   | 下游系统               |
| order-status-changed | 状态变更  | Order Service   | 下游系统               |

### 6.2 事件格式 (CloudEvents)

```json
{
    "specversion": "1.0",
    "type": "order-paid",
    "source": "payment-service",
    "id": "event-001",
    "time": "2026-04-22T10:35:00Z",
    "datacontenttype": "application/json",
    "data": {
        "order_id": 1,
        "order_no": "202504220001",
        "user_id": 10001,
        "old_status": 0,
        "new_status": 1,
        "pay_time": "2026-04-22T10:35:00Z",
        "pay_method": "alipay"
    }
}
```

## 7. 服务间通信

### 7.1 Order Service → Payment Service

使用Dapr Service Invocation:

```go
// Order Service调用Payment Service发起支付
client.InvokeMethod(ctx, "payment-service", "v1/payments", "POST", paymentReq)
```

### 7.2 Payment Service → Order Service (状态更新)

Payment Service支付成功后：

1. 通过Dapr State API更新订单状态
2. 发布 `order-paid` 事件到Redis Pub/Sub
3. Order Service订阅事件并处理

## 8. 项目结构

```
dapr-oms/
├── docker-compose.yml          # 服务编排
├── Makefile                    # 构建脚本
├── README.md                   # 项目说明
├──
├── order-service/              # 订单服务
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   ├── handlers/
│   │   └── order_handler.go
│   ├── models/
│   │   └── order.go
│   ├── repository/
│   │   └── order_repo.go
│   ├── services/
│   │   └── order_service.go
│   └── dapr.yaml              # Dapr配置
│
├── payment-service/            # 支付服务
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   ├── handlers/
│   │   └── payment_handler.go
│   ├── models/
│   │   └── payment.go
│   ├── services/
│   │   └── payment_service.go
│   └── dapr.yaml
│
├── components/                 # Dapr组件配置
│   ├── statestore.yaml
│   └── pubsub.yaml
│
├── scripts/                    # 脚本
│   └── init-db.sql
│
└── shared/                     # 共享代码
    ├── events/
    │   └── order_events.go
    └── dto/
        └── common.go
```

## 9. 部署说明

### 9.1 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 9.2 端口映射

| 服务              | 端口   | 说明                |
| --------------- | ---- | ----------------- |
| Order Service   | 8080 | HTTP API          |
| Order Dapr      | 3500 | Dapr Sidecar HTTP |
| Payment Service | 8081 | HTTP API          |
| Payment Dapr    | 3501 | Dapr Sidecar HTTP |
| MySQL           | 3306 | 数据库               |
| Redis           | 6379 | 消息队列              |

## 10. 测试验证

### 10.1 创建订单

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":10001,"items":[{"product_id":101,"product_name":"iPhone 15","unit_price":5999,"quantity":1}]}'
```

### 10.2 支付订单

```bash
curl -X POST http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{"order_no":"202504220001","pay_method":"alipay"}'
```

## 11. 错误处理

### 11.1 统一响应格式

```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功, 非0=错误码
    Message string      `json:"message"` // 提示信息
    Data    interface{} `json:"data"`    // 数据
}
```

### 11.2 错误码定义

| 错误码  | 说明     |
| ---- | ------ |
| 0    | 成功     |
| 1001 | 参数错误   |
| 1002 | 订单不存在  |
| 1003 | 订单状态非法 |
| 1004 | 支付失败   |
| 5000 | 系统错误   |

## 12. 扩展性考虑

- **支付方式扩展**: Payment Service预留接口，可对接真实支付渠道
- **订单状态机**: 使用状态机模式管理订单流转
- **事件消费**: 消费者可独立部署，支持水平扩展
- **追踪**: 可接入Dapr分布式追踪（OpenTelemetry）

---

**设计日期**: 2026-04-22
**版本**: v1.0
