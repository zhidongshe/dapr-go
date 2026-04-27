package services

import (
	"context"
	"fmt"
	"os"
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
	status := models.ProductStatusOnSale
	if req.Status != nil {
		status = *req.Status
	}

	product := &models.Product{
		ProductName:   req.ProductName,
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
