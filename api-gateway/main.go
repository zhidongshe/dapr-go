package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dapr-oms/api-gateway/handlers"
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8090"
	}

	r := gin.Default()

	// CORS
	r.Use(middleware.CORS())

	// 公开接口
	r.POST("/api/auth/login", handlers.Login)
	r.POST("/api/auth/logout", handlers.Logout)

	// 需要认证的接口
	api := r.Group("/api")
	{
		handlers.RegisterOrderRoutes(api)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("API Gateway starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
