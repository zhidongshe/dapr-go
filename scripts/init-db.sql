CREATE DATABASE IF NOT EXISTS oms_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE oms_db;

CREATE TABLE IF NOT EXISTS orders (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no        VARCHAR(32) NOT NULL UNIQUE COMMENT '订单编号',
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    total_amount    DECIMAL(12,2) NOT NULL COMMENT '订单总金额',
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '订单状态: 0待支付 1已支付 2处理中 3已发货 4已完成 5已取消',
    pay_status      TINYINT NOT NULL DEFAULT 0 COMMENT '支付状态: 0未支付 1已支付 2支付失败 3已退款',
    pay_time        DATETIME NULL COMMENT '支付时间',
    pay_method      VARCHAR(20) NULL COMMENT '支付方式',
    remark          VARCHAR(500) NULL COMMENT '备注',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

CREATE TABLE IF NOT EXISTS order_items (
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

-- Inventory master table
CREATE TABLE IF NOT EXISTS inventory (
    product_id      BIGINT PRIMARY KEY COMMENT '商品ID',
    product_name    VARCHAR(200) NOT NULL COMMENT '商品名称',
    available_stock INT NOT NULL DEFAULT 0 COMMENT '可用库存',
    reserved_stock  INT NOT NULL DEFAULT 0 COMMENT '已预占库存',
    version         INT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存主表';

-- Insert sample data
INSERT INTO inventory (product_id, product_name, available_stock, reserved_stock) VALUES
(1, 'iPhone 16', 100, 0),
(2, 'AirPods Pro', 200, 0),
(3, 'MacBook Pro', 50, 0),
(4, 'iPad Pro', 80, 0),
(5, 'Apple Watch', 150, 0)
ON DUPLICATE KEY UPDATE product_name = VALUES(product_name);

-- Inventory reservation table
CREATE TABLE IF NOT EXISTS inventory_reservation (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_no        VARCHAR(32) NOT NULL COMMENT '订单号',
    product_id      BIGINT NOT NULL COMMENT '商品ID',
    quantity        INT NOT NULL COMMENT '预占数量',
    status          TINYINT NOT NULL DEFAULT 0 COMMENT '0预占 1已扣减 2已释放',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_product (order_no, product_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预占记录表';

-- Message idempotency table for consumers
CREATE TABLE IF NOT EXISTS processed_messages (
    message_id      VARCHAR(64) PRIMARY KEY COMMENT '消息唯一ID',
    topic           VARCHAR(64) NOT NULL COMMENT '消息主题',
    processed_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_processed_at (processed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='已处理消息表';
