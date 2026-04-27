package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dapr-oms/product-service/models"
	_ "github.com/go-sql-driver/mysql"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(dsn string) (*ProductRepository, error) {
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

	return &ProductRepository{db: db}, nil
}

func (r *ProductRepository) Close() error {
	return r.db.Close()
}

func (r *ProductRepository) CreateProduct(product *models.Product) error {
	result, err := r.db.Exec(
		`INSERT INTO products (product_name, original_price, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		product.ProductName, product.OriginalPrice, product.Status, product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert product failed: %w", err)
	}

	productID, _ := result.LastInsertId()
	product.ProductID = productID

	return nil
}

func (r *ProductRepository) GetProductByID(productID int64) (*models.Product, error) {
	product := &models.Product{}
	err := r.db.QueryRow(
		`SELECT product_id, product_name, original_price, status, created_at, updated_at
		 FROM products WHERE product_id = ?`, productID,
	).Scan(&product.ProductID, &product.ProductName, &product.OriginalPrice,
		&product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) UpdatePrice(productID int64, price int64) error {
	_, err := r.db.Exec(
		`UPDATE products SET original_price = ? WHERE product_id = ?`,
		price, productID,
	)
	return err
}

func (r *ProductRepository) UpdateStatus(productID int64, status int) error {
	_, err := r.db.Exec(
		`UPDATE products SET status = ? WHERE product_id = ?`,
		status, productID,
	)
	return err
}

func (r *ProductRepository) ListProducts(status *int, keyword string, limit, offset int) ([]models.Product, int64, error) {
	where := "1=1"
	args := []interface{}{}

	if status != nil {
		where += " AND status = ?"
		args = append(args, *status)
	}

	if keyword != "" {
		where += " AND product_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := r.db.QueryRow("SELECT COUNT(*) FROM products WHERE "+where, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(
		"SELECT product_id, product_name, original_price, status, created_at, updated_at FROM products WHERE "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ProductID, &product.ProductName, &product.OriginalPrice,
			&product.Status, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}

	return products, total, nil
}
