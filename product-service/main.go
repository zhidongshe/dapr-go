package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dapr-oms/product-service/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8083"
	}

	r := gin.Default()

	productHandler := handlers.NewProductHandler()

	api := r.Group("/api/v1")
	{
		api.POST("/products", productHandler.CreateProduct)
		api.GET("/products", productHandler.ListProducts)
		api.GET("/products/:id", productHandler.GetProduct)
		api.PUT("/products/:id/price", productHandler.UpdatePrice)
		api.PUT("/products/:id/status", productHandler.UpdateStatus)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Product Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
