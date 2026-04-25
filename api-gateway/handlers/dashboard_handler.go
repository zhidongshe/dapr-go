package handlers

import (
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

// DashboardStats 看板统计
type DashboardStats struct {
	Orders    OrderStats     `json:"orders"`
	Payments  PaymentStats   `json:"payments"`
	Inventory InventoryStats `json:"inventory"`
}

type OrderStats struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Paid       int `json:"paid"`
	Processing int `json:"processing"`
	Shipped    int `json:"shipped"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
}

type PaymentStats struct {
	TodayAmount float64 `json:"todayAmount"`
	TodayCount  int     `json:"todayCount"`
	WeekAmount  float64 `json:"weekAmount"`
	MonthAmount float64 `json:"monthAmount"`
}

type InventoryStats struct {
	TotalProducts int `json:"totalProducts"`
	WarningCount  int `json:"warningCount"`
}

// GetDashboardStats 获取看板统计数据
func GetDashboardStats(c *gin.Context) {
	stats := DashboardStats{
		Orders: OrderStats{
			Total:      1000,
			Pending:    50,
			Paid:       200,
			Processing: 30,
			Shipped:    100,
			Completed:  590,
			Cancelled:  30,
		},
		Payments: PaymentStats{
			TodayAmount: 15000.00,
			TodayCount:  45,
			WeekAmount:  98000.00,
			MonthAmount: 450000.00,
		},
		Inventory: InventoryStats{
			TotalProducts: 500,
			WarningCount:  10,
		},
	}

	utils.Success(c, stats)
}

// RegisterDashboardRoutes 注册看板路由
func RegisterDashboardRoutes(r *gin.RouterGroup) {
	dashboard := r.Group("/dashboard")
	dashboard.Use(middleware.Auth())
	{
		dashboard.GET("/stats", GetDashboardStats)
	}
}
