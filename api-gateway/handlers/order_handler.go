package handlers

import (
	"bytes"
	"encoding/json"
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

// CreateOrder 创建订单
func CreateOrder(c *gin.Context) {
	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	// 读取原始请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.Error(c, utils.CodeBadRequest, "failed to read request body")
		return
	}
	// 恢复请求体以便后续使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 解析为 map 进行验证（可选）
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		utils.Error(c, utils.CodeBadRequest, "invalid JSON")
		return
	}

	resp, err := client.ForwardPOST("order", "/api/v1/orders", bodyMap, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	// 复制响应
	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// RegisterOrderRoutes 注册订单路由
func RegisterOrderRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	orders.Use(middleware.Auth())
	{
		orders.GET("", ListOrders)
		orders.GET("/:id", GetOrder)
		orders.POST("", CreateOrder)
	}
}
