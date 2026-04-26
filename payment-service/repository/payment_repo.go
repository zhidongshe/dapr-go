package repository

import (
	"database/sql"
	"time"

	"github.com/dapr-oms/payment-service/models"
	_ "github.com/go-sql-driver/mysql"
)

// PaymentRepository 支付仓库
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository 创建支付仓库
func NewPaymentRepository(dsn string) (*PaymentRepository, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PaymentRepository{db: db}, nil
}

// Close 关闭数据库连接
func (r *PaymentRepository) Close() error {
	return r.db.Close()
}

// CreatePayment 创建支付记录
func (r *PaymentRepository) CreatePayment(payment *models.Payment) error {
	result, err := r.db.Exec(
		`INSERT INTO payments (order_no, order_id, transaction_id, amount, pay_method, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		payment.OrderNo, payment.OrderID, payment.TransactionID, payment.Amount, payment.PayMethod, payment.Status,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	payment.ID = id
	return nil
}

// GetPaymentByTransactionID 根据交易号查询支付记录
func (r *PaymentRepository) GetPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	payment := &models.Payment{}
	var payTime sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, order_no, order_id, transaction_id, amount, pay_method, status, pay_time, fail_reason, created_at, updated_at
		 FROM payments WHERE transaction_id = ?`,
		transactionID,
	).Scan(&payment.ID, &payment.OrderNo, &payment.OrderID, &payment.TransactionID, &payment.Amount,
		&payment.PayMethod, &payment.Status, &payTime, &payment.FailReason,
		&payment.CreatedAt, &payment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if payTime.Valid {
		payment.PayTime = &payTime.Time
	}

	return payment, nil
}

// GetPaymentsByOrderNo 根据订单号查询支付记录
func (r *PaymentRepository) GetPaymentsByOrderNo(orderNo string) ([]*models.Payment, error) {
	rows, err := r.db.Query(
		`SELECT id, order_no, order_id, transaction_id, amount, pay_method, status, pay_time, fail_reason, created_at, updated_at
		 FROM payments WHERE order_no = ? ORDER BY created_at DESC`,
		orderNo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.Payment
	for rows.Next() {
		payment := &models.Payment{}
		var payTime sql.NullTime

		err := rows.Scan(&payment.ID, &payment.OrderNo, &payment.OrderID, &payment.TransactionID, &payment.Amount,
			&payment.PayMethod, &payment.Status, &payTime, &payment.FailReason,
			&payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if payTime.Valid {
			payment.PayTime = &payTime.Time
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// GetSuccessPaymentByOrderNo 查询订单的成功支付记录
func (r *PaymentRepository) GetSuccessPaymentByOrderNo(orderNo string) (*models.Payment, error) {
	payment := &models.Payment{}
	var payTime sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, order_no, order_id, transaction_id, amount, pay_method, status, pay_time, fail_reason, created_at, updated_at
		 FROM payments WHERE order_no = ? AND status = ? LIMIT 1`,
		orderNo, models.PaymentStatusSuccess,
	).Scan(&payment.ID, &payment.OrderNo, &payment.OrderID, &payment.TransactionID, &payment.Amount,
		&payment.PayMethod, &payment.Status, &payTime, &payment.FailReason,
		&payment.CreatedAt, &payment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if payTime.Valid {
		payment.PayTime = &payTime.Time
	}

	return payment, nil
}

// UpdatePaymentStatus 更新支付状态
func (r *PaymentRepository) UpdatePaymentStatus(transactionID string, status int, payTime *time.Time, failReason string) error {
	var payTimeValue sql.NullTime
	if payTime != nil {
		payTimeValue = sql.NullTime{Time: *payTime, Valid: true}
	}

	_, err := r.db.Exec(
		`UPDATE payments SET status = ?, pay_time = ?, fail_reason = ?, updated_at = NOW()
		 WHERE transaction_id = ?`,
		status, payTimeValue, failReason, transactionID,
	)
	return err
}

// GetDB 获取数据库连接（用于健康检查）
func (r *PaymentRepository) GetDB() *sql.DB {
	return r.db
}
