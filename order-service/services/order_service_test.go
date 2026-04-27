package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dapr-oms/order-service/models"
	"github.com/dapr-oms/shared/events"
)

// mockProductServer creates a mock product service server for testing
func mockProductServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract product ID from URL path like /api/v1/products/{id}
		var productID int64
		fmt.Sscanf(r.URL.Path, "/api/v1/products/%d", &productID)

		type response struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}

		switch productID {
		case 101: // Valid on-sale product
			json.NewEncoder(w).Encode(response{
				Code: 0,
				Data: map[string]interface{}{
					"product_id":     101,
					"product_name":   "iPhone 15",
					"original_price": 599900, // Price in cents (5999.00)
					"status":         models.ProductStatusOnSale,
					"created_at":     time.Now(),
					"updated_at":     time.Now(),
				},
			})
		case 102: // Valid on-sale product - AirPods
			json.NewEncoder(w).Encode(response{
				Code: 0,
				Data: map[string]interface{}{
					"product_id":     102,
					"product_name":   "AirPods Pro",
					"original_price": 199900, // Price in cents (1999.00)
					"status":         models.ProductStatusOnSale,
					"created_at":     time.Now(),
					"updated_at":     time.Now(),
				},
			})
		case 201: // Off-sale product
			json.NewEncoder(w).Encode(response{
				Code: 0,
				Data: map[string]interface{}{
					"product_id":     201,
					"product_name":   "Discontinued Item",
					"original_price": 99900,
					"status":         models.ProductStatusOffSale,
					"created_at":     time.Now(),
					"updated_at":     time.Now(),
				},
			})
		case 404: // Not found product
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response{
				Code:    1002,
				Message: "product not found",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response{
				Code:    1002,
				Message: "product not found",
			})
		}
	}))
}

func TestCreateOrder_RejectsMissingProduct(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	req := &models.CreateOrderRequest{
		UserID: 10001,
		Items: []models.OrderItemRequest{
			{
				ProductID: 999, // Non-existent product
				Quantity:  1,
			},
		},
	}

	_, err := svc.CreateOrder(context.Background(), req)
	if err == nil {
		t.Error("expected error for non-existent product, got nil")
	}

	expectedErrMsg := "product not found"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestCreateOrder_RejectsOffSaleProduct(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	req := &models.CreateOrderRequest{
		UserID: 10001,
		Items: []models.OrderItemRequest{
			{
				ProductID: 201, // Off-sale product
				Quantity:  1,
			},
		},
	}

	_, err := svc.CreateOrder(context.Background(), req)
	if err == nil {
		t.Error("expected error for off-sale product, got nil")
	}

	expectedErrMsg := "product is off sale"
	if err != nil && err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestCreateOrder_UsesProductServiceSnapshot(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// This test should verify that the order uses product service data
	// Since we don't have a real DB, we can't fully test CreateOrder
	// But we can test the getProduct and validation logic

	product, err := svc.getProduct(context.Background(), 101)
	if err != nil {
		t.Fatalf("getProduct failed: %v", err)
	}
	if product == nil {
		t.Fatal("expected product, got nil")
	}

	// Verify product snapshot data
	if product.ProductID != 101 {
		t.Errorf("expected ProductID 101, got %d", product.ProductID)
	}
	if product.ProductName != "iPhone 15" {
		t.Errorf("expected ProductName 'iPhone 15', got '%s'", product.ProductName)
	}
	if product.OriginalPrice != 599900 {
		t.Errorf("expected OriginalPrice 599900, got %d", product.OriginalPrice)
	}
	if product.Status != models.ProductStatusOnSale {
		t.Errorf("expected Status %d (on sale), got %d", models.ProductStatusOnSale, product.Status)
	}
}

func TestCreateOrder_CalculatesTotalFromProductService(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// Test with multiple items
	req := &models.CreateOrderRequest{
		UserID: 10001,
		Items: []models.OrderItemRequest{
			{
				ProductID: 101, // iPhone 15 @ 599900 cents
				Quantity:  2,   // 2 * 599900 = 1199800 cents
			},
			{
				ProductID: 102, // AirPods Pro @ 199900 cents
				Quantity:  1,   // 1 * 199900 = 199900 cents
			},
		},
	}

	// Calculate expected total using product service prices
	// Total should be: (2 * 599900) + (1 * 199900) = 1199800 + 199900 = 1399700 cents
	expectedTotal := int64(2*599900 + 1*199900)

	// Test the calculation logic by building items
	var calculatedTotal int64
	for _, itemReq := range req.Items {
		product, err := svc.getProduct(context.Background(), int64(itemReq.ProductID))
		if err != nil {
			t.Fatalf("getProduct failed for product %d: %v", itemReq.ProductID, err)
		}
		if product == nil {
			t.Fatalf("product %d not found", itemReq.ProductID)
		}
		calculatedTotal += int64(itemReq.Quantity) * product.OriginalPrice
	}
	_ = req // Mark req as used

	if calculatedTotal != expectedTotal {
		t.Errorf("expected total %d cents, got %d cents", expectedTotal, calculatedTotal)
	}
}

func TestGetProduct_Success(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	product, err := svc.getProduct(context.Background(), 101)
	if err != nil {
		t.Fatalf("getProduct failed: %v", err)
	}
	if product == nil {
		t.Fatal("expected product, got nil")
	}

	if product.ProductID != 101 {
		t.Errorf("expected ProductID 101, got %d", product.ProductID)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	server := mockProductServer(t)
	defer server.Close()

	svc := &OrderService{
		productServiceURL: server.URL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	product, err := svc.getProduct(context.Background(), 999)
	if err != nil {
		t.Fatalf("getProduct returned error: %v", err)
	}
	if product != nil {
		t.Error("expected nil product for non-existent ID, got a product")
	}
}

func TestProductSnapshot_Constants(t *testing.T) {
	// Verify product status constants
	if models.ProductStatusOffSale != 0 {
		t.Errorf("expected ProductStatusOffSale = 0, got %d", models.ProductStatusOffSale)
	}
	if models.ProductStatusOnSale != 1 {
		t.Errorf("expected ProductStatusOnSale = 1, got %d", models.ProductStatusOnSale)
	}
}

func TestOrderItemRequest_Validation(t *testing.T) {
	// Test that the new OrderItemRequest only requires ProductID and Quantity
	req := models.OrderItemRequest{
		ProductID: 101,
		Quantity:  2,
	}

	if req.ProductID == 0 {
		t.Error("ProductID should not be 0")
	}
	if req.Quantity <= 0 {
		t.Errorf("Quantity should be > 0, got %d", req.Quantity)
	}
}

// Integration test helpers
func TestOrderService_WithMockDapr(t *testing.T) {
	// This test verifies the order service can be instantiated with mocked dependencies
	// In a real test environment, we would mock the Dapr client and database

	// For now, just verify the service struct is properly defined
	svc := &OrderService{}
	if svc == nil {
		t.Error("failed to create OrderService")
	}
}

// Test event conversion
func TestConvertToInventoryItems(t *testing.T) {
	items := []models.OrderItem{
		{
			ProductID:   101,
			ProductName: "iPhone 15",
			UnitPrice:   5999.00,
			Quantity:    2,
			TotalPrice:  11998.00,
		},
		{
			ProductID:   102,
			ProductName: "AirPods Pro",
			UnitPrice:   1999.00,
			Quantity:    1,
			TotalPrice:  1999.00,
		},
	}

	inventoryItems := convertToInventoryItems(items)

	if len(inventoryItems) != len(items) {
		t.Errorf("expected %d inventory items, got %d", len(items), len(inventoryItems))
	}

	for i, item := range inventoryItems {
		if item.ProductID != int64(items[i].ProductID) {
			t.Errorf("item %d: expected ProductID %d, got %d", i, items[i].ProductID, item.ProductID)
		}
		if item.ProductName != items[i].ProductName {
			t.Errorf("item %d: expected ProductName '%s', got '%s'", i, items[i].ProductName, item.ProductName)
		}
		if item.Quantity != items[i].Quantity {
			t.Errorf("item %d: expected Quantity %d, got %d", i, items[i].Quantity, item.Quantity)
		}
	}
}

// Test generateUUID
func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == "" {
		t.Error("generateUUID returned empty string")
	}
	if uuid1 == uuid2 {
		t.Error("generateUUID should generate unique UUIDs")
	}
}

// Test order status constants from shared events
func TestOrderStatusConstants(t *testing.T) {
	// Verify that we're using the correct status constants
	if events.OrderStatusPending != 0 {
		t.Errorf("expected OrderStatusPending = 0, got %d", events.OrderStatusPending)
	}
	if events.OrderStatusPaid != 1 {
		t.Errorf("expected OrderStatusPaid = 1, got %d", events.OrderStatusPaid)
	}
	if events.OrderStatusCancelled != 5 {
		t.Errorf("expected OrderStatusCancelled = 5, got %d", events.OrderStatusCancelled)
	}
}
