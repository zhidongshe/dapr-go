package handlers

import (
	"io"

	"github.com/dapr-oms/api-gateway/client"
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

// ListProducts 获取商品列表
func ListProducts(c *gin.Context) {
	query := c.Request.URL.RawQuery
	path := "/api/v1/products"
	if query != "" {
		path = path + "?" + query
	}

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardGET("product", path, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// GetProduct 获取商品详情
func GetProduct(c *gin.Context) {
	id := c.Param("id")
	path := "/api/v1/products/" + id

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardGET("product", path, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// CreateProduct 创建商品
func CreateProduct(c *gin.Context) {
	path := "/api/v1/products"

	// 解析为 map 以传递给转发函数
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, utils.CodeBadRequest, "invalid request body")
		return
	}

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardPOST("product", path, body, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// UpdateProductPrice 更新商品价格
func UpdateProductPrice(c *gin.Context) {
	id := c.Param("id")
	path := "/api/v1/products/" + id + "/price"

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, utils.CodeBadRequest, "invalid request body")
		return
	}

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardPUT("product", path, body, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// UpdateProductStatus 更新商品状态
func UpdateProductStatus(c *gin.Context) {
	id := c.Param("id")
	path := "/api/v1/products/" + id + "/status"

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, utils.CodeBadRequest, "invalid request body")
		return
	}

	headers := map[string]string{
		"Content-Type": c.GetHeader("Content-Type"),
	}

	resp, err := client.ForwardPUT("product", path, body, headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// RegisterProductRoutes 注册商品路由
func RegisterProductRoutes(r *gin.RouterGroup) {
	products := r.Group("/products")
	products.Use(middleware.Auth())
	{
		products.GET("", ListProducts)
		products.GET("/:id", GetProduct)
		products.POST("", CreateProduct)
		products.PUT("/:id/price", UpdateProductPrice)
		products.PUT("/:id/status", UpdateProductStatus)
	}
}
