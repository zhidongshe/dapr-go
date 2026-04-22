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
