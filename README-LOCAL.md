# Dapr OMS 本地开发指南

本指南介绍如何在本地（不使用 Docker）运行 Dapr OMS 订单管理系统。

## 前置依赖

### 1. 安装必要软件

```bash
# macOS (使用 Homebrew)
brew install go mysql redis dapr/tap/dapr-cli

# Ubuntu/Debian
wget -q https://raw.githubusercontent.com/dapr/cli/master/install/install.sh -O - | /bin/bash
sudo apt-get install golang mysql-server redis-server
```

### 2. 启动 MySQL

```bash
# macOS
brew services start mysql

# 设置 root 密码为 rootpassword（或修改配置使用你自己的密码）
mysql_secure_installation

# 创建数据库
mysql -u root -p < scripts/init-db.sql
```

### 3. 启动 Redis

```bash
# macOS
brew services start redis

# 验证 Redis 运行
redis-cli ping
# 应返回: PONG
```

### 4. 初始化 Dapr

```bash
# 初始化 Dapr（只运行一次）
dapr init

# 验证安装
dapr --version
```

## 快速启动

### 方式一：使用一键启动脚本（推荐）

```bash
# 给脚本执行权限
chmod +x start-local.sh stop-local.sh

# 启动所有服务
./start-local.sh
```

脚本会自动：
1. 检查所有依赖是否安装
2. 启动 MySQL 和 Redis（如果未运行）
3. 初始化数据库
4. 安装 Go 依赖
5. 启动 4 个进程（2 个服务 + 2 个 Dapr Sidecar）

如果使用 tmux，会打开 4 个窗口；否则在后台运行。

### 方式二：手动启动（用于调试）

打开 4 个终端窗口，分别执行：

**Terminal 1 - Order Service Dapr Sidecar:**
```bash
dapr run \
  --app-id order-service \
  --app-port 8080 \
  --dapr-http-port 3500 \
  --components-path ./components
```

**Terminal 2 - Order Service:**
```bash
cd order-service
export MYSQL_DSN="root:rootpassword@tcp(localhost:3306)/oms_db?charset=utf8mb4&parseTime=true"
go run main.go
```

**Terminal 3 - Payment Service Dapr Sidecar:**
```bash
dapr run \
  --app-id payment-service \
  --app-port 8081 \
  --dapr-http-port 3501 \
  --components-path ./components
```

**Terminal 4 - Payment Service:**
```bash
cd payment-service
export ORDER_SERVICE_URL="http://localhost:8080"
go run main.go
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Order Service | 8080 | 订单服务 API |
| Order Dapr | 3500 | Order Service 的 Dapr Sidecar |
| Payment Service | 8081 | 支付服务 API |
| Payment Dapr | 3501 | Payment Service 的 Dapr Sidecar |
| MySQL | 3306 | 数据库 |
| Redis | 6379 | 消息队列 |

## 测试 API

```bash
# 1. 创建订单
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
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
  }'

# 2. 查询订单列表
curl "http://localhost:8080/api/v1/orders?user_id=10001"

# 3. 查询订单详情
curl http://localhost:8080/api/v1/orders/1

# 4. 支付订单（使用返回的 order_no）
curl -X POST http://localhost:8081/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_no": "ORDxxxxxxxxxxxxx",
    "pay_method": "alipay"
  }'

# 5. 取消订单
curl -X POST http://localhost:8080/api/v1/orders/1/cancel \
  -H "Content-Type: application/json" \
  -d '{"reason": "测试取消"}'
```

## 停止服务

```bash
# 使用停止脚本
./stop-local.sh

# 或者手动停止
# 按 Ctrl+C 停止各个进程
```

## 常见问题

### 1. MySQL 连接失败

检查 MySQL 是否运行：
```bash
mysql -u root -p -e "SELECT 1"
```

修改密码后，更新以下地方：
- `order-service/services/order_service.go` 中的默认 DSN
- 启动脚本中的 `MYSQL_DSN` 环境变量

### 2. Dapr 端口被占用

检查并杀死占用进程：
```bash
lsof -i :3500
lsof -i :3501
```

### 3. 依赖下载失败

确保 Go 模块代理设置正确：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 4. 端口冲突

如果默认端口被占用，修改：
- Dapr sidecar 端口: `--dapr-http-port` 参数
- 服务端口: 修改 `main.go` 中的默认端口或设置 `APP_PORT` 环境变量

## 开发调试

### 热重载

使用 `air` 工具实现代码修改后自动重启：

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 在 order-service 目录下运行
cd order-service
air
```

### 查看 Dapr 日志

Dapr 日志包含服务间调用和消息发布的信息，有助于调试：

```bash
# 查看 Order Service 的 Dapr 日志
curl http://localhost:3500/v1.0/metadata

# 查看发布订阅状态
curl http://localhost:3500/v1.0/subscribe
```

### 数据库调试

```bash
# 进入 MySQL
mysql -u root -p oms_db

# 查看订单
SELECT * FROM orders;
SELECT * FROM order_items;
```
