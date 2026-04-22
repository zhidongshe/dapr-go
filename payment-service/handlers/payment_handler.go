package handlers

import (
	"net/http"

	"github.com/dapr-oms/payment-service/models"
	"github.com/dapr-oms/payment-service/services"
	"github.com/dapr-oms/shared/dto"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service *services.PaymentService
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		service: services.NewPaymentService(),
	}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	resp, err := h.service.ProcessPayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(1004, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *PaymentHandler) PaymentCallback(c *gin.Context) {
	var req models.PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	if err := h.service.HandleCallback(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}
