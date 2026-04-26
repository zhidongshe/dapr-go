package handlers

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/dapr-oms/api-gateway/client"
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	Orders    OrderStats     `json:"orders"`
	Payments  PaymentStats   `json:"payments"`
	Inventory InventoryStats `json:"inventory"`
}

type OrderStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	Paid       int64 `json:"paid"`
	Processing int64 `json:"processing"`
	Shipped    int64 `json:"shipped"`
	Completed  int64 `json:"completed"`
	Cancelled  int64 `json:"cancelled"`
}

type PaymentStats struct {
	TodayAmount float64 `json:"todayAmount"`
	TodayCount  int64   `json:"todayCount"`
	WeekAmount  float64 `json:"weekAmount"`
	MonthAmount float64 `json:"monthAmount"`
}

type InventoryStats struct {
	TotalProducts int `json:"totalProducts"`
	WarningCount  int `json:"warningCount"`
}

type apiResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

type orderStatusCount struct {
	Status int     `json:"status"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

type inventoryItem struct {
	AvailableStock int `json:"available_stock"`
}

func GetDashboardStats(c *gin.Context) {
	var stats DashboardStats
	headers := map[string]string{"Content-Type": "application/json"}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		resp, err := client.ForwardGET("order", "/api/v1/orders/stats", headers)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var apiResp apiResponse
		if json.Unmarshal(body, &apiResp) != nil || apiResp.Code != 0 {
			return
		}
		var counts []orderStatusCount
		if json.Unmarshal(apiResp.Data, &counts) != nil {
			return
		}
		for _, sc := range counts {
			stats.Orders.Total += sc.Count
			switch sc.Status {
			case 0:
				stats.Orders.Pending = sc.Count
			case 1:
				stats.Orders.Paid = sc.Count
			case 2:
				stats.Orders.Processing = sc.Count
			case 3:
				stats.Orders.Shipped = sc.Count
			case 4:
				stats.Orders.Completed = sc.Count
			case 5:
				stats.Orders.Cancelled = sc.Count
			}
		}
	}()

	go func() {
		defer wg.Done()
		resp, err := client.ForwardGET("payment", "/api/v1/payments/stats", headers)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var apiResp apiResponse
		if json.Unmarshal(body, &apiResp) != nil || apiResp.Code != 0 {
			return
		}
		json.Unmarshal(apiResp.Data, &stats.Payments)
	}()

	go func() {
		defer wg.Done()
		resp, err := client.ForwardGET("inventory", "/api/v1/inventory", headers)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var apiResp apiResponse
		if json.Unmarshal(body, &apiResp) != nil || apiResp.Code != 0 {
			return
		}
		var items []inventoryItem
		if json.Unmarshal(apiResp.Data, &items) != nil {
			return
		}
		stats.Inventory.TotalProducts = len(items)
		for _, item := range items {
			if item.AvailableStock < 20 {
				stats.Inventory.WarningCount++
			}
		}
	}()

	wg.Wait()
	utils.Success(c, stats)
}

func RegisterDashboardRoutes(r *gin.RouterGroup) {
	dashboard := r.Group("/dashboard")
	dashboard.Use(middleware.Auth())
	{
		dashboard.GET("/stats", GetDashboardStats)
	}
}
