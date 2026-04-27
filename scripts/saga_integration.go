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
	ProductServiceURL   = "http://localhost:8083"
)

// CreateOrderRequest - updated to only send product_id and quantity
// Product name and price are fetched from product service
type CreateOrderRequest struct {
	UserID int64 `json:"user_id"`
	Items  []struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	} `json:"items"`
}

type OrderData struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	OrderNo     string  `json:"order_no"`
	Status      int     `json:"status"`
	TotalAmount int64   `json:"total_amount"`
	TotalAmountFloat float64 `json:"total_amount_float"`
}

type OrderResponse struct {
	Code    int       `json:"code"`
	Data    OrderData `json:"data"`
	Message string    `json:"message"`
}

type OrderDetailResponse struct {
	Code int `json:"code"`
	Data struct {
		ID          int64   `json:"id"`
		OrderID     int64   `json:"order_id"`
		OrderNo     string  `json:"order_no"`
		Status      int     `json:"status"`
		TotalAmount float64 `json:"total_amount"`
		Items       []struct {
			ProductID   int64   `json:"product_id"`
			ProductName string  `json:"product_name"`
			UnitPrice   float64 `json:"unit_price"`
			Quantity    int     `json:"quantity"`
			TotalPrice  float64 `json:"total_price"`
		} `json:"items"`
	} `json:"data"`
	Message string `json:"message"`
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

// Product types for product service integration
type Product struct {
	ProductID     int64  `json:"product_id"`
	ProductName   string `json:"product_name"`
	OriginalPrice int64  `json:"original_price"`
	Status        int    `json:"status"`
}

type ProductResponse struct {
	Code    int     `json:"code"`
	Data    Product `json:"data"`
	Message string  `json:"message"`
}

type UpdatePriceRequest struct {
	OriginalPrice int64 `json:"original_price"`
}

type UpdateStatusRequest struct {
	Status int `json:"status"`
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
	testScenario4_PriceChangeAffectsNewOrdersOnly()
	testScenario5_OffSaleProductBlocksNewOrders()

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
		ProductServiceURL + "/health",
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

	// Step 2: Create order (only send product_id and quantity)
	fmt.Println("Step 2: Create order")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID int64 `json:"product_id"`
			Quantity  int   `json:"quantity"`
		}{
			{ProductID: 1, Quantity: 1},
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
	// Get order details to find the actual total amount
	orderDetail := getOrderDetail(getOrderID(order))
	if orderDetail == nil {
		testResults = append(testResults, "❌ Scenario 1: Failed to get order details")
		return
	}
	payReq := PaymentRequest{
		OrderNo:   order.Data.OrderNo,
		Amount:    orderDetail.Data.TotalAmount,
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
			ProductID int64 `json:"product_id"`
			Quantity  int   `json:"quantity"`
		}{
			{ProductID: 999, Quantity: 1000},
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
			ProductID int64 `json:"product_id"`
			Quantity  int   `json:"quantity"`
		}{
			{ProductID: 2, Quantity: 2},
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

// Scenario 4: Price Change Affects New Orders Only
// Create Order -> Change Product Price -> Create New Order -> Verify Old Order Keeps Historical Price
func testScenario4_PriceChangeAffectsNewOrdersOnly() {
	fmt.Println("📌 Scenario 4: Price Change Affects New Orders Only")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Get product 1 initial price
	fmt.Println("Step 1: Get product 1 initial price")
	product := getProduct(1)
	if product == nil {
		testResults = append(testResults, "❌ Scenario 4: Failed to get product")
		return
	}
	originalPrice := product.Data.OriginalPrice
	fmt.Printf("  Product 1 original price: %d cents\n", originalPrice)

	// Step 2: Create first order
	fmt.Println("Step 2: Create first order")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID int64 `json:"product_id"`
			Quantity  int   `json:"quantity"`
		}{
			{ProductID: 1, Quantity: 1},
		},
	}

	order1 := createOrder(orderReq)
	if order1 == nil {
		testResults = append(testResults, "❌ Scenario 4: Failed to create first order")
		return
	}
	fmt.Printf("  Order 1 created: ID=%d, OrderNo=%s\n", getOrderID(order1), order1.Data.OrderNo)

	// Get order 1 details to capture the unit price
	order1Detail := getOrderDetail(getOrderID(order1))
	if order1Detail == nil {
		testResults = append(testResults, "❌ Scenario 4: Failed to get order 1 details")
		return
	}
	order1UnitPrice := order1Detail.Data.Items[0].UnitPrice
	order1Total := order1Detail.Data.TotalAmount
	fmt.Printf("  Order 1 unit price: %.2f, total: %.2f\n", order1UnitPrice, order1Total)

	// Step 3: Update product price (increase by 10000 cents = 100 yuan)
	fmt.Println("Step 3: Update product price")
	newPrice := originalPrice + 10000
	if err := updateProductPrice(1, newPrice); err != nil {
		testResults = append(testResults, "❌ Scenario 4: Failed to update product price")
		return
	}
	fmt.Printf("  Product price updated from %d to %d cents\n", originalPrice, newPrice)

	// Step 4: Create second order (should use new price)
	fmt.Println("Step 4: Create second order (should use new price)")
	order2 := createOrder(orderReq)
	if order2 == nil {
		// Restore original price before failing
		updateProductPrice(1, originalPrice)
		testResults = append(testResults, "❌ Scenario 4: Failed to create second order")
		return
	}
	fmt.Printf("  Order 2 created: ID=%d, OrderNo=%s\n", getOrderID(order2), order2.Data.OrderNo)

	// Get order 2 details
	order2Detail := getOrderDetail(getOrderID(order2))
	if order2Detail == nil {
		updateProductPrice(1, originalPrice)
		testResults = append(testResults, "❌ Scenario 4: Failed to get order 2 details")
		return
	}
	order2UnitPrice := order2Detail.Data.Items[0].UnitPrice
	order2Total := order2Detail.Data.TotalAmount
	fmt.Printf("  Order 2 unit price: %.2f, total: %.2f\n", order2UnitPrice, order2Total)

	// Step 5: Verify old order keeps historical price
	fmt.Println("Step 5: Verify old order keeps historical price")

	// Check order 1 still has original price
	order1Check := getOrderDetail(getOrderID(order1))
	if order1Check == nil {
		updateProductPrice(1, originalPrice)
		testResults = append(testResults, "❌ Scenario 4: Failed to verify order 1 price")
		return
	}
	if order1Check.Data.Items[0].UnitPrice != order1UnitPrice {
		updateProductPrice(1, originalPrice)
		testResults = append(testResults, "❌ Scenario 4: Order 1 price changed after product update")
		return
	}
	fmt.Printf("  Order 1 still has original unit price: %.2f\n", order1Check.Data.Items[0].UnitPrice)

	// Verify order 2 has new price
	if order2UnitPrice <= order1UnitPrice {
		updateProductPrice(1, originalPrice)
		testResults = append(testResults, "❌ Scenario 4: Order 2 does not have increased price")
		return
	}
	fmt.Printf("  Order 2 has new higher price: %.2f\n", order2UnitPrice)

	// Step 6: Restore original price
	fmt.Println("Step 6: Restore original price")
	if err := updateProductPrice(1, originalPrice); err != nil {
		testResults = append(testResults, "❌ Scenario 4: Failed to restore product price")
		return
	}
	fmt.Println("  Original price restored")

	// Cancel both orders to clean up
	cancelOrder(getOrderID(order1))
	cancelOrder(getOrderID(order2))

	testResults = append(testResults, "✅ Scenario 4: Price Change Affects New Orders Only - PASSED")
	fmt.Println()
}

// Scenario 5: Off-Sale Product Blocks New Orders
// Create Order with On-Sale Product -> Set Product Off-Sale -> Try Create Order -> Should Fail
func testScenario5_OffSaleProductBlocksNewOrders() {
	fmt.Println("📌 Scenario 5: Off-Sale Product Blocks New Orders")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Ensure product 3 is on sale
	fmt.Println("Step 1: Ensure product 3 is on sale")
	product := getProduct(3)
	if product == nil {
		testResults = append(testResults, "❌ Scenario 5: Failed to get product")
		return
	}
	if product.Data.Status != 1 {
		if err := updateProductStatus(3, 1); err != nil {
			testResults = append(testResults, "❌ Scenario 5: Failed to set product on sale")
			return
		}
		fmt.Println("  Product 3 set to on-sale")
	} else {
		fmt.Println("  Product 3 is already on-sale")
	}

	// Step 2: Create order with on-sale product (should succeed)
	fmt.Println("Step 2: Create order with on-sale product")
	orderReq := CreateOrderRequest{
		UserID: 1001,
		Items: []struct {
			ProductID int64 `json:"product_id"`
			Quantity  int   `json:"quantity"`
		}{
			{ProductID: 3, Quantity: 1},
		},
	}

	order1 := createOrder(orderReq)
	if order1 == nil {
		testResults = append(testResults, "❌ Scenario 5: Failed to create order with on-sale product")
		return
	}
	fmt.Printf("  Order created successfully: ID=%d, OrderNo=%s\n", getOrderID(order1), order1.Data.OrderNo)

	// Step 3: Set product off-sale
	fmt.Println("Step 3: Set product 3 off-sale")
	if err := updateProductStatus(3, 0); err != nil {
		testResults = append(testResults, "❌ Scenario 5: Failed to set product off-sale")
		return
	}
	fmt.Println("  Product 3 is now off-sale")

	// Step 4: Try to create order with off-sale product (should fail)
	fmt.Println("Step 4: Try to create order with off-sale product")
	order2 := createOrder(orderReq)
	if order2 != nil {
		// Restore product status before failing
		updateProductStatus(3, 1)
		cancelOrder(getOrderID(order2))
		testResults = append(testResults, "❌ Scenario 5: Order with off-sale product should have failed")
		return
	}
	fmt.Println("  Order creation correctly rejected for off-sale product")

	// Step 5: Restore product on-sale status
	fmt.Println("Step 5: Restore product on-sale status")
	if err := updateProductStatus(3, 1); err != nil {
		testResults = append(testResults, "❌ Scenario 5: Failed to restore product status")
		return
	}
	fmt.Println("  Product 3 restored to on-sale")

	// Cancel first order to clean up
	cancelOrder(getOrderID(order1))

	testResults = append(testResults, "✅ Scenario 5: Off-Sale Product Blocks New Orders - PASSED")
	fmt.Println()
}

// Helper function to get order ID from response (handles both id and order_id fields)
func getOrderID(order *OrderResponse) int64 {
	if order.Data.OrderID != 0 {
		return order.Data.OrderID
	}
	return order.Data.ID
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

func getOrderDetail(orderID int64) *OrderDetailResponse {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/orders/%d", OrderServiceURL, orderID))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result OrderDetailResponse
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

// Product service helper functions

func getProduct(productID int64) *ProductResponse {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/products/%d", ProductServiceURL, productID))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("  HTTP Error: %d\n", resp.StatusCode)
		return nil
	}

	var result ProductResponse
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

func updateProductPrice(productID int64, originalPrice int64) error {
	reqBody := UpdatePriceRequest{OriginalPrice: originalPrice}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/products/%d/price", ProductServiceURL, productID), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("API error: %s", result.Message)
	}

	return nil
}

func updateProductStatus(productID int64, status int) error {
	reqBody := UpdateStatusRequest{Status: status}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/products/%d/status", ProductServiceURL, productID), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("API error: %s", result.Message)
	}

	return nil
}
