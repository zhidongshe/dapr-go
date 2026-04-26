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

// GetPaymentByTransactionID 根据交易号查询支付记录
func (h *PaymentHandler) GetPaymentByTransactionID(c *gin.Context) {
	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "transaction_id is required"))
		return
	}

	payment, err := h.service.GetPaymentByTransactionID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	if payment == nil {
		c.JSON(http.StatusNotFound, dto.Error(1002, "payment not found"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(payment))
}

func (h *PaymentHandler) GetPaymentStats(c *gin.Context) {
	stats, err := h.service.GetPaymentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(stats))
}

// GetPaymentsByOrderNo 根据订单号查询支付记录
func (h *PaymentHandler) GetPaymentsByOrderNo(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "order_no is required"))
		return
	}

	payments, err := h.service.GetPaymentsByOrderNo(orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(payments))
}
