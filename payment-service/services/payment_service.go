package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr-oms/payment-service/models"
	"github.com/dapr-oms/payment-service/repository"
	"github.com/dapr-oms/shared/events"
)

type PaymentService struct {
	daprClient      dapr.Client
	orderServiceURL string
	repo            *repository.PaymentRepository
}

func NewPaymentService() *PaymentService {
	client, err := dapr.NewClient()
	if err != nil {
		panic(fmt.Sprintf("failed to create dapr client: %v", err))
	}

	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://order-service:8080"
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
	}

	repo, err := repository.NewPaymentRepository(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to create payment repository: %v", err))
	}

	return &PaymentService{
		daprClient:      client,
		orderServiceURL: orderURL,
		repo:            repo,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// Get order info from order service
	order, err := s.getOrderByNo(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("get order failed: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", req.OrderNo)
	}

	if order.Status != events.OrderStatusPending {
		return nil, fmt.Errorf("order cannot be paid, current status: %d", order.Status)
	}

	// Check if there's already a successful payment for this order
	existingPayment, err := s.repo.GetSuccessPaymentByOrderNo(req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("check existing payment failed: %w", err)
	}
	if existingPayment != nil {
		return &models.PaymentResponse{
			TransactionID: existingPayment.TransactionID,
			Status:        models.GetStatusText(existingPayment.Status),
			Message:       "order already paid",
		}, nil
	}

	// Generate transaction ID
	transactionID := models.GenerateTransactionID()

	// Create payment record
	payment := &models.Payment{
		OrderNo:       req.OrderNo,
		OrderID:       order.ID,
		TransactionID: transactionID,
		Amount:        req.Amount,
		PayMethod:     req.PayMethod,
		Status:        models.PaymentStatusPending,
	}

	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("create payment record failed: %w", err)
	}

	fmt.Printf("payment record created: transaction_id=%s, order_no=%s\n", transactionID, req.OrderNo)

	// Simulate payment processing
	// For demo, we simulate successful payment
	paySuccess := true

	if paySuccess {
		payTime := time.Now()

		// Update payment status to success
		if err := s.repo.UpdatePaymentStatus(transactionID, models.PaymentStatusSuccess, &payTime, ""); err != nil {
			return nil, fmt.Errorf("update payment status failed: %w", err)
		}

		// Publish order paid event
		event := events.OrderPaidEvent{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			OldStatus: events.OrderStatusPending,
			NewStatus: events.OrderStatusPaid,
			PayTime:   payTime,
			PayMethod: req.PayMethod,
		}

		eventData, _ := json.Marshal(event)
		if err := s.daprClient.PublishEvent(ctx, "order-pubsub", events.TopicOrderPaid, eventData); err != nil {
			fmt.Printf("failed to publish order paid event: %v\n", err)
		}

		fmt.Printf("payment success: transaction_id=%s, order_no=%s\n", transactionID, req.OrderNo)

		return &models.PaymentResponse{
			TransactionID: transactionID,
			Status:        "success",
			Message:       "payment processed successfully",
		}, nil
	}

	// Payment failed
	failReason := "payment processing failed"
	if err := s.repo.UpdatePaymentStatus(transactionID, models.PaymentStatusFailed, nil, failReason); err != nil {
		return nil, fmt.Errorf("update payment status failed: %w", err)
	}

	return &models.PaymentResponse{
		TransactionID: transactionID,
		Status:        "failed",
		Message:       failReason,
	}, nil
}

func (s *PaymentService) HandleCallback(ctx context.Context, req *models.PaymentCallbackRequest) error {
	// Handle async payment callback
	payment, err := s.repo.GetPaymentByTransactionID(req.TransactionID)
	if err != nil {
		return fmt.Errorf("get payment failed: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found: %s", req.TransactionID)
	}

	if req.Status == "success" {
		// Check if already processed
		if payment.Status == models.PaymentStatusSuccess {
			return nil
		}

		payTime := time.Now()
		if err := s.repo.UpdatePaymentStatus(req.TransactionID, models.PaymentStatusSuccess, &payTime, ""); err != nil {
			return fmt.Errorf("update payment status failed: %w", err)
		}

		order, err := s.getOrderByNo(ctx, req.OrderNo)
		if err != nil {
			return err
		}

		event := events.OrderPaidEvent{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			OldStatus: events.OrderStatusPending,
			NewStatus: events.OrderStatusPaid,
			PayTime:   payTime,
			PayMethod: "callback",
		}

		eventData, _ := json.Marshal(event)
		return s.daprClient.PublishEvent(ctx, "order-pubsub", events.TopicOrderPaid, eventData)
	}

	// Payment failed
	failReason := "callback reported failure"
	if err := s.repo.UpdatePaymentStatus(req.TransactionID, models.PaymentStatusFailed, nil, failReason); err != nil {
		return fmt.Errorf("update payment status failed: %w", err)
	}

	return nil
}

func (s *PaymentService) GetPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	return s.repo.GetPaymentByTransactionID(transactionID)
}

func (s *PaymentService) GetPaymentsByOrderNo(orderNo string) ([]*models.Payment, error) {
	return s.repo.GetPaymentsByOrderNo(orderNo)
}

func (s *PaymentService) GetPaymentStats() (*repository.PaymentStats, error) {
	return s.repo.GetPaymentStats()
}

func (s *PaymentService) getOrderByNo(ctx context.Context, orderNo string) (*models.OrderInfo, error) {
	// Use HTTP call to order service
	url := fmt.Sprintf("%s/api/v1/orders?order_no=%s", s.orderServiceURL, orderNo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    models.OrderInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("order service error: %s", result.Message)
	}

	return &result.Data, nil
}
