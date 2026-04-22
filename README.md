# Dapr OMS 订单管理系统

基于Dapr构建的订单管理系统，包含订单服务和支付服务，使用MySQL存储数据，Redis Pub/Sub发布订单状态变更事件。

## 架构

- **Order Service**: 订单管理服务 (端口: 8080)
- **Payment Service**: 支付处理服务 (端口: 8081)
- **Dapr Sidecar**: 每个服务配一个Dapr sidecar提供状态管理和消息发布
- **MySQL**: 订单数据持久化
- **Redis**: 消息队列 (Pub/Sub)

## 快速开始

### 启动服务

```bash
make up
```

### 测试API

```bash
# 创建订单
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":10001,"items":[{"product_id":101,"product_name":"iPhone 15","unit_price":5999,"quantity":1}]}'

# 支付订单
curl -X POST http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{"order_no":"ORDxxxxxxxx","pay_method":"alipay"}'

# 查询订单
curl http://localhost:8080/api/v1/orders/1

# 取消订单
curl -X POST http://localhost:8080/api/v1/orders/1/cancel
```

## API文档

### Order Service

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/orders | 创建订单 |
| GET | /api/v1/orders/:id | 查询订单详情 |
| GET | /api/v1/orders?user_id=xxx | 查询订单列表 |
| POST | /api/v1/orders/:id/cancel | 取消订单 |

### Payment Service

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/payments | 发起支付 |
| POST | /api/v1/payments/callback | 支付回调 |

## 订单状态

- 0: 待支付
- 1: 已支付
- 2: 处理中
- 3: 已发货
- 4: 已完成
- 5: 已取消

## Dapr组件

- **State Store**: MySQL - 存储订单状态
- **Pub/Sub**: Redis - 发布订单状态变更事件

## 开发

```bash
# 本地构建
make build

# 查看日志
make logs

# 停止并清理
make down
```
