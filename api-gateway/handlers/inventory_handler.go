package handlers

import (
	"io"

	"github.com/dapr-oms/api-gateway/client"
	"github.com/dapr-oms/api-gateway/middleware"
	"github.com/dapr-oms/api-gateway/utils"
	"github.com/gin-gonic/gin"
)

func ListInventory(c *gin.Context) {
	headers := map[string]string{"Content-Type": c.GetHeader("Content-Type")}
	resp, err := client.ForwardGET("inventory", "/api/v1/inventory", headers)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, "failed to forward request")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func RegisterInventoryRoutes(r *gin.RouterGroup) {
	inv := r.Group("/inventory")
	inv.Use(middleware.Auth())
	{
		inv.GET("", ListInventory)
	}
}
