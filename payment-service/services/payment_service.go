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
	"github.com/dapr-oms/shared/events"
)

type PaymentService struct {
	daprClient      dapr.Client
	orderServiceURL string
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

	return &PaymentService{
		daprClient:      client,
		orderServiceURL: orderURL,
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

	// Simulate payment processing
	transactionID := models.GenerateTransactionID()

	// For demo, we simulate successful payment
	paySuccess := true

	if paySuccess {
		// Publish order paid event
		event := events.OrderPaidEvent{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			UserID:    order.UserID,
			OldStatus: events.OrderStatusPending,
			NewStatus: events.OrderStatusPaid,
			PayTime:   time.Now(),
			PayMethod: req.PayMethod,
		}

		eventData, _ := json.Marshal(event)
		if err := s.daprClient.PublishEvent(ctx, "order-pubsub", events.TopicOrderPaid, eventData); err != nil {
			fmt.Printf("failed to publish order paid event: %v\n", err)
		}

		return &models.PaymentResponse{
			TransactionID: transactionID,
			Status:        "success",
			Message:       "payment processed successfully",
		}, nil
	}

	return &models.PaymentResponse{
		TransactionID: transactionID,
		Status:        "failed",
		Message:       "payment failed",
	}, nil
}

func (s *PaymentService) HandleCallback(ctx context.Context, req *models.PaymentCallbackRequest) error {
	// Handle async payment callback
	if req.Status == "success" {
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
			PayTime:   time.Now(),
			PayMethod: "callback",
		}

		eventData, _ := json.Marshal(event)
		return s.daprClient.PublishEvent(ctx, "order-pubsub", events.TopicOrderPaid, eventData)
	}

	return nil
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
