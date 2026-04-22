#!/bin/bash

echo "=== 停止 Dapr OMS 本地服务 ==="

# 停止 tmux 会话（如果存在）
if tmux has-session -t dapr-oms 2>/dev/null; then
    echo "停止 tmux 会话..."
    tmux kill-session -t dapr-oms
    echo "✓ tmux 会话已停止"
fi

# 停止 Dapr 进程
echo "停止 Dapr 进程..."
pkill -f "dapr run --app-id order-service" 2>/dev/null || true
pkill -f "dapr run --app-id payment-service" 2>/dev/null || true

# 停止 Go 服务
echo "停止 Go 服务..."
pkill -f "order-service/main.go" 2>/dev/null || true
pkill -f "payment-service/main.go" 2>/dev/null || true
pkill -f "./order-service" 2>/dev/null || true
pkill -f "./payment-service" 2>/dev/null || true

echo "✓ 所有服务已停止"
