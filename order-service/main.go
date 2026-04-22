package main

import (
    "log"
    "net/http"
    "os"

    "github.com/dapr-oms/order-service/handlers"
    "github.com/gin-gonic/gin"
)

func main() {
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = "8080"
    }

    r := gin.Default()

    orderHandler := handlers.NewOrderHandler()

    api := r.Group("/api/v1")
    {
        api.POST("/orders", orderHandler.CreateOrder)
        api.GET("/orders/:id", orderHandler.GetOrder)
        api.GET("/orders", orderHandler.ListOrders)
        api.POST("/orders/:id/cancel", orderHandler.CancelOrder)
    }

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    log.Printf("Order Service starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
