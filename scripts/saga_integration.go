package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	OrderServiceURL     = "http://localhost:8080"
	InventoryServiceURL = "http://localhost:8082"
	PaymentServiceURL   = "http://localhost:8081"
)

type CreateOrderRequest struct {
	UserID int64 `json:"user_id"`
	Items  []struct {
		ProductID   int64  `json:"product_id"`
		ProductName string `json:"product_name"`
		Quantity    int    `json:"quantity"`
		UnitPrice   int64  `json:"unit_price"`
	} `json:"items"`
}

type OrderData struct {
	ID          int64  `json:"id"`
	OrderID     int64  `json:"order_id"`
	OrderNo     string `json:"order_no"`
	Status      int    `json:"status"`
	TotalAmount int64  `json:"total_amount"`
}

type OrderResponse struct {
	Code    int       `json:"code"`
	Data    OrderData `json:"data"`
	Message string    `json:"message"`
}

type InventoryData struct {
	ProductID      int64 `json:"product_id"`
	AvailableStock int   `json:"available_stock"`
	ReservedStock  int   `json:"reserved_stock"`
}

type InventoryResponse struct {
	Code int           `json:"code"`
	Data InventoryData `json:"data"`
}

type PaymentRequest struct {
	OrderNo   string  `json:"order_no"`
	PayMethod string  `json:"pay_method"`
	Amount    float64 `json:"amount"`
}

type PaymentResponse struct {
	Code int `json:"code"`
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
	Message string `json:"message"`
}

var testResults []string

func main() {
	fmt.Println("====================================")
	fmt.Println("  Saga Integration Test Suite")
	fmt.Println("====================================")
	fmt.Println()

	// Wait for services to be ready
	waitForServices()

	// Run test scenarios
	fmt.Println("\n📋 Test Scenarios:")
	fmt.Println()

	testScenario1_NormalFlow()
	testScenario2_InsufficientStock()
	testScenario3_CancelOrder()

	// Summary
	fmt.Println("\n====================================")
	fmt.Println("  Test Summary")
	fmt.Println("====================================")
	passed := 0
	failed := 0
	for _, result := range testResults {
		if strings.HasPrefix(result, "✅") {
			passed++
		} else {
			failed++
		}
		fmt.Println(result)
	}
	fmt.Printf("\nTotal: %d tests, %d passed, %d failed\n", len(testResults), passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func waitForServices() {
	fmt.Println("⏳ Waiting for services to be ready...")
	services := []string{
		OrderServiceURL + "/health",
		InventoryServiceURL + "/api/v1/inventory/1",
		PaymentServiceURL + "/health",
	}

	for _, url := range services {
		for i := 0; i < 30; i++ {
			resp, err := http.Get(url)
			if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 404) {
				resp.Body.Close()
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Println("✅ Services are ready\n")
}

// Scenario 1: Normal Flow
// Create Order -> Reserve Stock -> Pay -> Confirm Stock
func testScenario1_NormalFlow() {
	fmt.Println("📌 Scenario 1: Normal Flow (Create -> Reserve -> Pay -> Confirm)")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Check initial inventory
	fmt.Println("Step 1: Check initial inventory for product 1")
	initialStock := getInventory(1)
	if initialStock == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to get initial inventory")
		return
	}
	fmt.Printf("  Initial stock: Available=%d, Reserved=%d\n", initialStock.Data.AvailableStock, initialStock.Data.ReservedStock)

	// Step 2: Create order
	fmt.Println("Step 2: Create order")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID   int64  `json:"product_id"`
			ProductName string `json:"product_name"`
			Quantity    int    `json:"quantity"`
			UnitPrice   int64  `json:"unit_price"`
		}{
			{ProductID: 1, ProductName: "iPhone 16", Quantity: 1, UnitPrice: 899900},
		},
	}

	order := createOrder(orderReq)
	if order == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to create order")
		return
	}
	fmt.Printf("  Order created: ID=%d, OrderNo=%s, Status=%d\n", getOrderID(order), order.Data.OrderNo, order.Data.Status)

	// Wait for inventory reservation
	time.Sleep(2 * time.Second)

	// Step 3: Check inventory after reservation
	fmt.Println("Step 3: Check inventory after reservation")
	afterReserveStock := getInventory(1)
	if afterReserveStock == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to get inventory after reservation")
		return
	}
	fmt.Printf("  After reserve: Available=%d, Reserved=%d\n", afterReserveStock.Data.AvailableStock, afterReserveStock.Data.ReservedStock)

	// Verify stock is reserved
	if afterReserveStock.Data.ReservedStock != initialStock.Data.ReservedStock+1 {
		testResults = append(testResults, "❌ Scenario 1: Stock not reserved properly")
		return
	}

	// Step 4: Process payment
	fmt.Println("Step 4: Process payment")
	payReq := PaymentRequest{
		OrderNo:   order.Data.OrderNo,
		Amount:    float64(order.Data.TotalAmount) / 100, // Convert cents to yuan
		PayMethod: "credit_card",
	}
	payment := processPayment(payReq)
	if payment == nil {
		testResults = append(testResults, "❌ Scenario 1: Payment request failed")
		return
	}
	fmt.Printf("  Payment status: %s\n", payment.Data.Status)

	// Wait for inventory confirmation
	time.Sleep(5 * time.Second)

	// Step 5: Check inventory after confirmation
	fmt.Println("Step 5: Check inventory after confirmation")
	afterConfirmStock := getInventory(1)
	if afterConfirmStock == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to get inventory after confirmation")
		return
	}
	fmt.Printf("  After confirm: Available=%d, Reserved=%d\n", afterConfirmStock.Data.AvailableStock, afterConfirmStock.Data.ReservedStock)

	// Verify stock is confirmed (reserved back to available, actually deducted)
	if afterConfirmStock.Data.ReservedStock != initialStock.Data.ReservedStock {
		testResults = append(testResults, "❌ Scenario 1: Stock not confirmed properly")
		return
	}

	// Step 6: Check order status
	fmt.Println("Step 6: Check order status")
	orderStatus := getOrder(getOrderID(order))
	if orderStatus == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to get order status")
		return
	}
	// Order status: 0=Pending, 1=Paid, 5=Cancelled
	fmt.Printf("  Order status: %d (1=PAID)\n", orderStatus.Data.Status)

	if orderStatus.Data.Status != 1 { // PAID status
		testResults = append(testResults, "❌ Scenario 1: Order status not PAID")
		return
	}

	testResults = append(testResults, "✅ Scenario 1: Normal Flow - PASSED")
	fmt.Println()
}

// Scenario 2: Insufficient Stock
// Create Order -> Reserve Stock Failed -> Order Cancelled
func testScenario2_InsufficientStock() {
	fmt.Println("📌 Scenario 2: Insufficient Stock (Create -> Reserve Failed -> Cancel)")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Check inventory for product with low stock
	fmt.Println("Step 1: Check inventory for product 999 (low stock)")
	initialStock := getInventory(999)
	if initialStock == nil {
		testResults = append(testResults, "❌ Scenario 2: Failed to get initial inventory")
		return
	}
	fmt.Printf("  Initial stock: Available=%d, Reserved=%d\n", initialStock.Data.AvailableStock, initialStock.Data.ReservedStock)

	// Step 2: Create order requesting more than available
	fmt.Println("Step 2: Create order with quantity exceeding stock")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID   int64  `json:"product_id"`
			ProductName string `json:"product_name"`
			Quantity    int    `json:"quantity"`
			UnitPrice   int64  `json:"unit_price"`
		}{
			{ProductID: 999, ProductName: "Limited Item", Quantity: 1000, UnitPrice: 10000},
		},
	}

	order := createOrder(orderReq)
	if order == nil {
		testResults = append(testResults, "❌ Scenario 2: Failed to create order")
		return
	}
	fmt.Printf("  Order created: ID=%d, OrderNo=%s, Status=%d\n", getOrderID(order), order.Data.OrderNo, order.Data.Status)

	// Wait for inventory reservation failure
	time.Sleep(3 * time.Second)

	// Step 3: Check order status should be CANCELLED
	fmt.Println("Step 3: Check order status (should be CANCELLED)")
	orderStatus := getOrder(getOrderID(order))
	if orderStatus == nil {
		testResults = append(testResults, "❌ Scenario 2: Failed to get order status")
		return
	}
	// Order status: 5=CANCELLED
	fmt.Printf("  Order status: %d (5=CANCELLED)\n", orderStatus.Data.Status)

	// Verify order is cancelled
	if orderStatus.Data.Status != 5 { // CANCELLED status
		testResults = append(testResults, "❌ Scenario 2: Order not cancelled after reserve failure")
		return
	}

	// Step 4: Verify inventory unchanged
	fmt.Println("Step 4: Verify inventory unchanged")
	afterStock := getInventory(999)
	if afterStock == nil {
		testResults = append(testResults, "❌ Scenario 2: Failed to get final inventory")
		return
	}
	fmt.Printf("  Final stock: Available=%d, Reserved=%d\n", afterStock.Data.AvailableStock, afterStock.Data.ReservedStock)

	if afterStock.Data.AvailableStock != initialStock.Data.AvailableStock {
		testResults = append(testResults, "❌ Scenario 2: Inventory was modified despite reserve failure")
		return
	}

	testResults = append(testResults, "✅ Scenario 2: Insufficient Stock - PASSED")
	fmt.Println()
}

// Scenario 3: Cancel Order
// Create Order -> Reserve Stock -> Cancel -> Release Stock
func testScenario3_CancelOrder() {
	fmt.Println("📌 Scenario 3: Cancel Order (Create -> Reserve -> Cancel -> Release)")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Check initial inventory
	fmt.Println("Step 1: Check initial inventory for product 2")
	initialStock := getInventory(2)
	if initialStock == nil {
		testResults = append(testResults, "❌ Scenario 3: Failed to get initial inventory")
		return
	}
	fmt.Printf("  Initial stock: Available=%d, Reserved=%d\n", initialStock.Data.AvailableStock, initialStock.Data.ReservedStock)

	// Step 2: Create order
	fmt.Println("Step 2: Create order")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID   int64  `json:"product_id"`
			ProductName string `json:"product_name"`
			Quantity    int    `json:"quantity"`
			UnitPrice   int64  `json:"unit_price"`
		}{
			{ProductID: 2, ProductName: "MacBook Pro", Quantity: 2, UnitPrice: 1999900},
		},
	}

	order := createOrder(orderReq)
	if order == nil {
		testResults = append(testResults, "❌ Scenario 3: Failed to create order")
		return
	}
	fmt.Printf("  Order created: ID=%d, OrderNo=%s, Status=%d\n", getOrderID(order), order.Data.OrderNo, order.Data.Status)

	// Wait for inventory reservation
	time.Sleep(2 * time.Second)

	// Step 3: Check inventory after reservation
	fmt.Println("Step 3: Check inventory after reservation")
	afterReserveStock := getInventory(2)
	if afterReserveStock == nil {
		testResults = append(testResults, "❌ Scenario 3: Failed to get inventory after reservation")
		return
	}
	fmt.Printf("  After reserve: Available=%d, Reserved=%d\n", afterReserveStock.Data.AvailableStock, afterReserveStock.Data.ReservedStock)

	// Verify stock is reserved
	if afterReserveStock.Data.ReservedStock != initialStock.Data.ReservedStock+2 {
		testResults = append(testResults, "❌ Scenario 3: Stock not reserved properly")
		return
	}

	// Step 4: Cancel order
	fmt.Println("Step 4: Cancel order")
	if !cancelOrder(getOrderID(order)) {
		testResults = append(testResults, "❌ Scenario 3: Failed to cancel order")
		return
	}
	fmt.Println("  Order cancelled successfully")

	// Wait for inventory release
	time.Sleep(2 * time.Second)

	// Step 5: Check inventory after release
	fmt.Println("Step 5: Check inventory after release")
	afterReleaseStock := getInventory(2)
	if afterReleaseStock == nil {
		testResults = append(testResults, "❌ Scenario 3: Failed to get inventory after release")
		return
	}
	fmt.Printf("  After release: Available=%d, Reserved=%d\n", afterReleaseStock.Data.AvailableStock, afterReleaseStock.Data.ReservedStock)

	// Verify stock is released
	if afterReleaseStock.Data.ReservedStock != initialStock.Data.ReservedStock {
		testResults = append(testResults, "❌ Scenario 3: Stock not released properly")
		return
	}
	if afterReleaseStock.Data.AvailableStock != initialStock.Data.AvailableStock {
		testResults = append(testResults, "❌ Scenario 3: Available stock not restored")
		return
	}

	// Step 6: Check order status
	fmt.Println("Step 6: Check order status")
	orderStatus := getOrder(getOrderID(order))
	if orderStatus == nil {
		testResults = append(testResults, "❌ Scenario 3: Failed to get order status")
		return
	}
	// Order status: 5=CANCELLED
	fmt.Printf("  Order status: %d (5=CANCELLED)\n", orderStatus.Data.Status)

	if orderStatus.Data.Status != 5 { // CANCELLED status
		testResults = append(testResults, "❌ Scenario 3: Order status not CANCELLED")
		return
	}

	testResults = append(testResults, "✅ Scenario 3: Cancel Order - PASSED")
	fmt.Println()
}

// Helper function to get order ID from response (handles both id and order_id fields)
func getOrderID(order *OrderResponse) int64 {
	if order.Data.OrderID != 0 {
		return order.Data.OrderID
	}
	return getOrderID(order)
}

// Helper functions

func getInventory(productID int64) *InventoryResponse {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/inventory/%d", InventoryServiceURL, productID))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("  HTTP Error: %d\n", resp.StatusCode)
		return nil
	}

	var result InventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("  Decode Error: %v\n", err)
		return nil
	}

	if result.Code != 0 {
		fmt.Printf("  API Error: code=%d\n", result.Code)
		return nil
	}

	return &result
}

func createOrder(req CreateOrderRequest) *OrderResponse {
	body, _ := json.Marshal(req)
	resp, err := http.Post(OrderServiceURL+"/api/v1/orders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("  Decode Error: %v\n", err)
		return nil
	}

	if result.Code != 0 {
		fmt.Printf("  API Error: code=%d, message=%s\n", result.Code, result.Message)
		return nil
	}

	return &result
}

func getOrder(orderID int64) *OrderResponse {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/orders/%d", OrderServiceURL, orderID))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("  Decode Error: %v\n", err)
		return nil
	}

	if result.Code != 0 {
		fmt.Printf("  API Error: code=%d, message=%s\n", result.Code, result.Message)
		return nil
	}

	return &result
}

func processPayment(req PaymentRequest) *PaymentResponse {
	body, _ := json.Marshal(req)
	resp, err := http.Post(PaymentServiceURL+"/api/v1/payments", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("  Decode Error: %v\n", err)
		return nil
	}

	return &result
}

func cancelOrder(orderID int64) bool {
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/cancel", OrderServiceURL, orderID), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
