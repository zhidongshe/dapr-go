package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dapr-oms/inventory-service/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8082"
	}

	r := gin.Default()

	handler := handlers.NewInventoryHandler()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Dapr subscription endpoint
	r.GET("/dapr/subscribe", handler.DaprSubscribe)

	log.Printf("Inventory Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
