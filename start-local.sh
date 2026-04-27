#!/bin/bash

# Dapr OMS 本地开发启动脚本
# 需要前置依赖: MySQL, Redis, Dapr CLI, Go

set -e

echo "=== Dapr OMS 本地开发启动脚本 ==="
echo ""

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查依赖
check_dependency() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}错误: $1 未安装${NC}"
        return 1
    fi
    echo -e "${GREEN}✓ $1 已安装${NC}"
}

echo "1. 检查依赖..."
check_dependency go
check_dependency dapr
check_dependency mysql
check_dependency redis-cli
echo ""

# 检查 MySQL 运行状态
echo "2. 检查 MySQL 状态..."
if mysql -u root -prootpassword -e "SELECT 1" &> /dev/null; then
    echo -e "${GREEN}✓ MySQL 运行中${NC}"
else
    echo -e "${YELLOW}警告: MySQL 未运行，尝试启动...${NC}"
    if command -v brew &> /dev/null; then
        brew services start mysql
    else
        echo -e "${RED}请手动启动 MySQL${NC}"
        exit 1
    fi
fi
echo ""

# 检查 Redis 运行状态
echo "3. 检查 Redis 状态..."
if redis-cli ping &> /dev/null; then
    echo -e "${GREEN}✓ Redis 运行中${NC}"
else
    echo -e "${YELLOW}警告: Redis 未运行，尝试启动...${NC}"
    if command -v brew &> /dev/null; then
        brew services start redis
    else
        echo -e "${RED}请手动启动 Redis${NC}"
        exit 1
    fi
fi
echo ""

# 初始化数据库
echo "4. 初始化数据库..."
mysql -u root -prootpassword < scripts/init-db.sql 2>/dev/null || {
    echo -e "${YELLOW}数据库已存在或初始化失败（可能已初始化过）${NC}"
}
echo -e "${GREEN}✓ 数据库就绪${NC}"
echo ""

# 安装 Go 依赖
echo "5. 安装 Go 依赖..."
cd shared && go mod tidy 2>/dev/null || true
cd ../order-service && go mod tidy
cd ../payment-service && go mod tidy
cd ../product-service && go mod tidy
cd ..
echo -e "${GREEN}✓ 依赖安装完成${NC}"
echo ""

# 创建 logs 目录
mkdir -p logs

echo "=== 启动服务 ==="
echo ""

# 使用 tmux 或新窗口启动（如果可用）
if command -v tmux &> /dev/null; then
    echo "使用 tmux 启动 6 个窗口..."

    # 创建新的 tmux 会话
    tmux new-session -d -s dapr-oms -n "order-dapr"

    # Window 1: Order Service Dapr Sidecar
    tmux send-keys -t dapr-oms:0 "dapr run --app-id order-service --app-port 8080 --dapr-http-port 3500 --components-path ./components --config components/statestore-local.yaml --config components/pubsub-local.yaml 2>&1 | tee logs/order-dapr.log" C-m

    # Window 2: Order Service
    tmux new-window -t dapr-oms -n "order-svc"
    tmux send-keys -t dapr-oms:1 "cd order-service && export MYSQL_DSN='root:rootpassword@tcp(localhost:3306)/oms_db?charset=utf8mb4&parseTime=true' && export PRODUCT_SERVICE_URL='http://localhost:8083' && go run main.go 2>&1 | tee ../logs/order-service.log" C-m

    # Window 3: Payment Service Dapr Sidecar
    tmux new-window -t dapr-oms -n "payment-dapr"
    tmux send-keys -t dapr-oms:2 "dapr run --app-id payment-service --app-port 8081 --dapr-http-port 3501 --components-path ./components --config components/statestore-local.yaml --config components/pubsub-local.yaml 2>&1 | tee logs/payment-dapr.log" C-m

    # Window 4: Payment Service
    tmux new-window -t dapr-oms -n "payment-svc"
    tmux send-keys -t dapr-oms:3 "cd payment-service && export ORDER_SERVICE_URL='http://localhost:8080' && go run main.go 2>&1 | tee ../logs/payment-service.log" C-m

    # Window 5: Product Service Dapr Sidecar
    tmux new-window -t dapr-oms -n "product-dapr"
    tmux send-keys -t dapr-oms:4 "dapr run --app-id product-service --app-port 8083 --dapr-http-port 3503 --components-path ./components --config components/statestore-local.yaml --config components/pubsub-local.yaml 2>&1 | tee logs/product-dapr.log" C-m

    # Window 6: Product Service
    tmux new-window -t dapr-oms -n "product-svc"
    tmux send-keys -t dapr-oms:5 "cd product-service && export MYSQL_DSN='root:rootpassword@tcp(localhost:3306)/oms_db?charset=utf8mb4&parseTime=true' && go run main.go 2>&1 | tee ../logs/product-service.log" C-m

    # 附加到会话
    echo -e "${GREEN}服务已启动！正在附加到 tmux 会话...${NC}"
    echo "tmux 快捷键: Ctrl+B 然后按数字切换窗口, Ctrl+B D 分离"
    sleep 2
    tmux attach -t dapr-oms

else
    # 没有 tmux，使用后台进程
    echo "后台启动服务（日志保存在 logs/ 目录）..."
    echo ""

    # 启动 Order Service Dapr Sidecar
    echo "启动 Order Service Dapr Sidecar (3500)..."
    dapr run --app-id order-service --app-port 8080 --dapr-http-port 3500 \
        --components-path ./components 2>&1 > logs/order-dapr.log &
    ORDER_DAPR_PID=$!

    sleep 2

    # 启动 Order Service
    echo "启动 Order Service (8080)..."
    cd order-service
    export MYSQL_DSN="root:rootpassword@tcp(localhost:3306)/oms_db?charset=utf8mb4&parseTime=true"
    export PRODUCT_SERVICE_URL="http://localhost:8083"
    go run main.go 2>&1 > ../logs/order-service.log &
    ORDER_SVC_PID=$!
    cd ..

    sleep 2

    # 启动 Payment Service Dapr Sidecar
    echo "启动 Payment Service Dapr Sidecar (3501)..."
    dapr run --app-id payment-service --app-port 8081 --dapr-http-port 3501 \
        --components-path ./components 2>&1 > logs/payment-dapr.log &
    PAYMENT_DAPR_PID=$!

    sleep 2

    # 启动 Payment Service
    echo "启动 Payment Service (8081)..."
    cd payment-service
    export ORDER_SERVICE_URL="http://localhost:8080"
    go run main.go 2>&1 > ../logs/payment-service.log &
    PAYMENT_SVC_PID=$!
    cd ..

    sleep 2

    # 启动 Product Service Dapr Sidecar
    echo "启动 Product Service Dapr Sidecar (3503)..."
    dapr run --app-id product-service --app-port 8083 --dapr-http-port 3503 \
        --components-path ./components 2>&1 > logs/product-dapr.log &
    PRODUCT_DAPR_PID=$!

    sleep 2

    # 启动 Product Service
    echo "启动 Product Service (8083)..."
    cd product-service
    export MYSQL_DSN="root:rootpassword@tcp(localhost:3306)/oms_db?charset=utf8mb4&parseTime=true"
    go run main.go 2>&1 > ../logs/product-service.log &
    PRODUCT_SVC_PID=$!
    cd ..

    echo ""
    echo -e "${GREEN}所有服务已后台启动！${NC}"
    echo ""
    echo "服务状态:"
    echo "  Order Service:       http://localhost:8080"
    echo "  Order Dapr Sidecar:  http://localhost:3500"
    echo "  Payment Service:     http://localhost:8081"
    echo "  Payment Dapr Sidecar:http://localhost:3501"
    echo "  Product Service:     http://localhost:8083"
    echo "  Product Dapr Sidecar:http://localhost:3503"
    echo ""
    echo "查看日志:"
    echo "  tail -f logs/order-service.log"
    echo "  tail -f logs/payment-service.log"
    echo "  tail -f logs/product-service.log"
    echo ""
    echo "停止服务:"
    echo "  kill $ORDER_DAPR_PID $ORDER_SVC_PID $PAYMENT_DAPR_PID $PAYMENT_SVC_PID $PRODUCT_DAPR_PID $PRODUCT_SVC_PID"
    echo ""
    echo "或者运行: ./stop-local.sh"
fi
