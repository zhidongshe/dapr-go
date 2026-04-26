package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dapr-oms/payment-service/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	r := gin.Default()

	paymentHandler := handlers.NewPaymentHandler()

	api := r.Group("/api/v1")
	{
		api.POST("/payments", paymentHandler.CreatePayment)
		api.POST("/payments/callback", paymentHandler.PaymentCallback)
		api.GET("/payments/:transaction_id", paymentHandler.GetPaymentByTransactionID)
		api.GET("/payments", paymentHandler.GetPaymentsByOrderNo)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Payment Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
