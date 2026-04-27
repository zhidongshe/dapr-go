package handlers

import (
	"net/http"
	"strconv"

	"github.com/dapr-oms/product-service/models"
	"github.com/dapr-oms/product-service/services"
	"github.com/dapr-oms/shared/dto"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		service: services.NewProductService(),
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(product))
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid product id"))
		return
	}

	product, err := h.service.GetProduct(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, dto.Error(1002, "product not found"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(product))
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	var status *int
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := strconv.Atoi(statusStr)
		if err == nil {
			status = &s
		}
	}

	keyword := c.Query("keyword")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	products, total, err := h.service.ListProducts(c.Request.Context(), status, keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	c.JSON(http.StatusOK, dto.Success(gin.H{
		"list":     products,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (h *ProductHandler) UpdatePrice(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid product id"))
		return
	}

	var req models.UpdatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	if err := h.service.UpdatePrice(c.Request.Context(), productID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid product id"))
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), productID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}
