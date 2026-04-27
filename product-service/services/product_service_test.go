package services

import (
	"context"
	"testing"

	"github.com/dapr-oms/product-service/models"
)

func TestCreateProduct_RejectsNonPositivePrice(t *testing.T) {
	// Create a mock repository or use a test database
	// For unit tests, we'll create a service with a mock
	svc := &ProductService{
		repo: nil, // We'll need to use a mock or test DB
	}

	ctx := context.Background()

	// Test zero price
	status := models.ProductStatusOnSale
	req := &models.CreateProductRequest{
		ProductName:   "Test Product",
		OriginalPrice: 0,
		Status:        &status,
	}
	_, err := svc.CreateProduct(ctx, req)
	if err == nil {
		t.Error("expected error for zero price, got nil")
	}
	if err != nil && err.Error() != "original_price must be greater than 0" {
		t.Errorf("expected 'original_price must be greater than 0', got '%s'", err.Error())
	}

	// Test negative price
	req.OriginalPrice = -100
	_, err = svc.CreateProduct(ctx, req)
	if err == nil {
		t.Error("expected error for negative price, got nil")
	}
	if err != nil && err.Error() != "original_price must be greater than 0" {
		t.Errorf("expected 'original_price must be greater than 0', got '%s'", err.Error())
	}
}

func TestCreateProduct_RejectsEmptyName(t *testing.T) {
	svc := &ProductService{}
	ctx := context.Background()

	status := models.ProductStatusOnSale
	req := &models.CreateProductRequest{
		ProductName:   "",
		OriginalPrice: 100,
		Status:        &status,
	}
	_, err := svc.CreateProduct(ctx, req)
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
	if err != nil && err.Error() != "product_name is required" {
		t.Errorf("expected 'product_name is required', got '%s'", err.Error())
	}

	// Test whitespace-only name
	req.ProductName = "   "
	_, err = svc.CreateProduct(ctx, req)
	if err == nil {
		t.Error("expected error for whitespace-only name, got nil")
	}
	if err != nil && err.Error() != "product_name is required" {
		t.Errorf("expected 'product_name is required', got '%s'", err.Error())
	}
}

func TestCreateProduct_RejectsInvalidStatus(t *testing.T) {
	svc := &ProductService{}
	ctx := context.Background()

	invalidStatus := 999
	req := &models.CreateProductRequest{
		ProductName:   "Test Product",
		OriginalPrice: 100,
		Status:        &invalidStatus,
	}
	_, err := svc.CreateProduct(ctx, req)
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
	if err != nil && err.Error() != "invalid status" {
		t.Errorf("expected 'invalid status', got '%s'", err.Error())
	}
}

func TestUpdateStatus_RejectsInvalidStatus(t *testing.T) {
	svc := &ProductService{}
	ctx := context.Background()

	// Create a mock that returns a valid product
	// For this test, we need to check validation before repository call

	req := &models.UpdateStatusRequest{
		Status: 999,
	}
	err := svc.UpdateStatus(ctx, 1, req)
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
	if err != nil && err.Error() != "invalid status" {
		t.Errorf("expected 'invalid status', got '%s'", err.Error())
	}
}

func TestListProducts_FiltersByStatusAndName(t *testing.T) {
	// This test verifies the service correctly passes filter parameters to repository
	// We'll need to verify the repository receives correct parameters

	// For integration testing, we would:
	// 1. Create products with different statuses and names
	// 2. Filter by status and verify only matching products returned
	// 3. Filter by keyword and verify only matching products returned
	// 4. Filter by both and verify intersection

	// For now, this is a placeholder that will be implemented with proper test setup
	t.Skip("requires database setup or mock repository")
}
