package repository

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/dapr-oms/order-service/models"
    _ "github.com/go-sql-driver/mysql"
)

type OrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(dsn string) (*OrderRepository, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &OrderRepository{db: db}, nil
}

func (r *OrderRepository) Close() error {
    return r.db.Close()
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    result, err := tx.Exec(
        `INSERT INTO orders (order_no, user_id, total_amount, status, pay_status, remark, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        order.OrderNo, order.UserID, order.TotalAmount, order.Status, order.PayStatus,
        order.Remark, order.CreatedAt, order.UpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("insert order failed: %w", err)
    }

    orderID, _ := result.LastInsertId()
    order.ID = uint64(orderID)

    for i := range order.Items {
        item := &order.Items[i]
        item.OrderID = order.ID
        item.TotalPrice = int64(item.Quantity) * item.UnitPrice

        _, err = tx.Exec(
            `INSERT INTO order_items (order_id, product_id, product_name, unit_price, quantity, total_price, created_at)
             VALUES (?, ?, ?, ?, ?, ?, ?)`,
            item.OrderID, item.ProductID, item.ProductName, item.UnitPrice,
            item.Quantity, item.TotalPrice, time.Now(),
        )
        if err != nil {
            return fmt.Errorf("insert order item failed: %w", err)
        }
    }

    return tx.Commit()
}

func (r *OrderRepository) GetOrderByID(orderID uint64) (*models.Order, error) {
    order := &models.Order{}
    var payMethod, remark sql.NullString
    err := r.db.QueryRow(
        `SELECT id, order_no, user_id, total_amount, status, pay_status, pay_time,
                pay_method, remark, created_at, updated_at
         FROM orders WHERE id = ?`, orderID,
    ).Scan(&order.ID, &order.OrderNo, &order.UserID, &order.TotalAmount,
        &order.Status, &order.PayStatus, &order.PayTime, &payMethod,
        &remark, &order.CreatedAt, &order.UpdatedAt)
    order.PayMethod = payMethod.String
    order.Remark = remark.String

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    rows, err := r.db.Query(
        `SELECT id, order_id, product_id, product_name, unit_price, quantity, total_price, created_at
         FROM order_items WHERE order_id = ?`, orderID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var item models.OrderItem
        err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
            &item.UnitPrice, &item.Quantity, &item.TotalPrice, &item.CreatedAt)
        if err != nil {
            return nil, err
        }
        order.Items = append(order.Items, item)
    }

    return order, nil
}

func (r *OrderRepository) GetOrderByNo(orderNo string) (*models.Order, error) {
    order := &models.Order{}
    var payMethod, remark sql.NullString
    err := r.db.QueryRow(
        `SELECT id, order_no, user_id, total_amount, status, pay_status, pay_time,
                pay_method, remark, created_at, updated_at
         FROM orders WHERE order_no = ?`, orderNo,
    ).Scan(&order.ID, &order.OrderNo, &order.UserID, &order.TotalAmount,
        &order.Status, &order.PayStatus, &order.PayTime, &payMethod,
        &remark, &order.CreatedAt, &order.UpdatedAt)
    order.PayMethod = payMethod.String
    order.Remark = remark.String

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return order, nil
}

func (r *OrderRepository) UpdateOrderStatus(orderID uint64, status int) error {
    _, err := r.db.Exec(
        `UPDATE orders SET status = ? WHERE id = ?`,
        status, orderID,
    )
    return err
}

func (r *OrderRepository) UpdatePayStatus(orderID uint64, payStatus int, payTime time.Time, payMethod string) error {
    _, err := r.db.Exec(
        `UPDATE orders SET pay_status = ?, pay_time = ?, pay_method = ? WHERE id = ?`,
        payStatus, payTime, payMethod, orderID,
    )
    return err
}

func (r *OrderRepository) ListOrders(userID uint64, status *int, limit, offset int) ([]models.Order, int64, error) {
    where := "1=1"
    args := []interface{}{}
    if userID > 0 {
        where += " AND user_id = ?"
        args = append(args, userID)
    }
    if status != nil {
        where += " AND status = ?"
        args = append(args, *status)
    }

    var total int64
    countArgs := make([]interface{}, len(args))
    copy(countArgs, args)
    err := r.db.QueryRow("SELECT COUNT(*) FROM orders WHERE "+where, countArgs...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    queryArgs := append(args, limit, offset)
    rows, err := r.db.Query(
        "SELECT id, order_no, user_id, total_amount, status, pay_status, created_at FROM orders WHERE "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
        queryArgs...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var orders []models.Order
    for rows.Next() {
        var order models.Order
        err := rows.Scan(&order.ID, &order.OrderNo, &order.UserID, &order.TotalAmount,
            &order.Status, &order.PayStatus, &order.CreatedAt)
        if err != nil {
            return nil, 0, err
        }
        orders = append(orders, order)
    }

    return orders, total, nil
}

type OrderStatusCount struct {
    Status int   `json:"status"`
    Count  int64 `json:"count"`
    Amount int64 `json:"amount"` // 单位：分
}

func (r *OrderRepository) ListPendingOrders() ([]models.Order, error) {
    rows, err := r.db.Query(
        `SELECT id, order_no, user_id, total_amount, status, pay_status, created_at
         FROM orders WHERE status = 0 ORDER BY created_at`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []models.Order
    for rows.Next() {
        var o models.Order
        if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.TotalAmount, &o.Status, &o.PayStatus, &o.CreatedAt); err != nil {
            return nil, err
        }
        orders = append(orders, o)
    }
    return orders, nil
}

func (r *OrderRepository) GetOrderStats() ([]OrderStatusCount, error) {
    rows, err := r.db.Query(
        `SELECT status, COUNT(*) as cnt, COALESCE(SUM(total_amount), 0) as amount
         FROM orders GROUP BY status`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var stats []OrderStatusCount
    for rows.Next() {
        var s OrderStatusCount
        if err := rows.Scan(&s.Status, &s.Count, &s.Amount); err != nil {
            return nil, err
        }
        stats = append(stats, s)
    }
    return stats, nil
}
