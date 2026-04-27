package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dapr-oms/product-service/models"
	"github.com/dapr-oms/product-service/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService() *ProductService {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
	}

	repo, err := repository.NewProductRepository(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	// Validate product name
	name := strings.TrimSpace(req.ProductName)
	if name == "" {
		return nil, errors.New("product_name is required")
	}

	// Validate price
	if req.OriginalPrice <= 0 {
		return nil, errors.New("original_price must be greater than 0")
	}

	// Validate and set status
	status := models.ProductStatusOnSale
	if req.Status != nil {
		status = *req.Status
	}
	if status != models.ProductStatusOnSale && status != models.ProductStatusOffSale {
		return nil, errors.New("invalid status")
	}

	product := &models.Product{
		ProductName:   name,
		OriginalPrice: req.OriginalPrice,
		Status:        status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateProduct(product); err != nil {
		return nil, fmt.Errorf("create product failed: %w", err)
	}

	return product, nil
}

func (s *ProductService) GetProduct(ctx context.Context, productID int64) (*models.Product, error) {
	return s.repo.GetProductByID(productID)
}

func (s *ProductService) UpdatePrice(ctx context.Context, productID int64, req *models.UpdatePriceRequest) error {
	product, err := s.repo.GetProductByID(productID)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	return s.repo.UpdatePrice(productID, req.OriginalPrice)
}

func (s *ProductService) UpdateStatus(ctx context.Context, productID int64, req *models.UpdateStatusRequest) error {
	// Validate status
	if req.Status != models.ProductStatusOnSale && req.Status != models.ProductStatusOffSale {
		return errors.New("invalid status")
	}

	product, err := s.repo.GetProductByID(productID)
	if err != nil {
		return err
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	return s.repo.UpdateStatus(productID, req.Status)
}

func (s *ProductService) ListProducts(ctx context.Context, status *int, keyword string, limit, offset int) ([]models.Product, int64, error) {
	return s.repo.ListProducts(status, keyword, limit, offset)
}
