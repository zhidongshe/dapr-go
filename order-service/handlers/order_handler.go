package handlers

import (
    "net/http"
    "strconv"

    "github.com/dapr-oms/order-service/models"
    "github.com/dapr-oms/order-service/services"
    "github.com/dapr-oms/shared/dto"
    "github.com/gin-gonic/gin"
)

type OrderHandler struct {
    service *services.OrderService
}

func NewOrderHandler() *OrderHandler {
    return &OrderHandler{
        service: services.NewOrderService(),
    }
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req models.CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
        return
    }

    order, err := h.service.CreateOrder(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    resp := models.OrderResponse{
        OrderID:     order.ID,
        OrderNo:     order.OrderNo,
        TotalAmount: order.TotalAmount,
        Status:      order.Status,
        CreatedAt:   order.CreatedAt,
    }
    c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
    // 支持通过 order_no 查询 (?order_no=xxx)
    if orderNo := c.Query("order_no"); orderNo != "" {
        order, err := h.service.GetOrderByNo(c.Request.Context(), orderNo)
        if err != nil {
            c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
            return
        }
        if order == nil {
            c.JSON(http.StatusNotFound, dto.Error(1002, "order not found"))
            return
        }
        c.JSON(http.StatusOK, dto.Success(order))
        return
    }

    // 通过 ID 查询 (/orders/:id)
    idStr := c.Param("id")
    orderID, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid order id"))
        return
    }

    order, err := h.service.GetOrder(c.Request.Context(), orderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }
    if order == nil {
        c.JSON(http.StatusNotFound, dto.Error(1002, "order not found"))
        return
    }

    c.JSON(http.StatusOK, dto.Success(order))
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
    userIDStr := c.Query("user_id")
    userID, err := strconv.ParseUint(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid user_id"))
        return
    }

    limitStr := c.DefaultQuery("limit", "10")
    offsetStr := c.DefaultQuery("offset", "0")
    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    orders, err := h.service.ListOrders(c.Request.Context(), userID, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.Success(orders))
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
    idStr := c.Param("id")
    orderID, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid order id"))
        return
    }

    var req struct {
        Reason string `json:"reason"`
    }
    c.ShouldBindJSON(&req)

    if err := h.service.CancelOrder(c.Request.Context(), orderID, req.Reason); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.Success(nil))
}
