package handlers

import (
	"io"

	"github.com/dapr-oms/api-gateway/client"
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

// ListOrders 获取订单列表
func ListOrders(c *gin.Context) {
	// 构建查询参数
	query := c.Request.URL.RawQuery
	path := "/api/v1/orders"
	if query != "" {
		path = path + "?" + query
	}

	// 转发请求
	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardGET("order", path, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	// 复制响应
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// GetOrder 获取订单详情
func GetOrder(c *gin.Context) {
	id := c.Param("id")
	path := "/api/v1/orders/" + id

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardGET("order", path, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// RegisterOrderRoutes 注册订单路由
func RegisterOrderRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	orders.Use(middleware.Auth())
	{
		orders.GET("", ListOrders)
		orders.GET("/:id", GetOrder)
	}
}
