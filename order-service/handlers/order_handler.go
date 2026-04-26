package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/dapr-oms/order-service/models"
    "github.com/dapr-oms/order-service/services"
    "github.com/dapr-oms/shared/dto"
    "github.com/dapr-oms/shared/events"
    "github.com/gin-gonic/gin"
)

type OrderHandler struct {
    service *services.OrderService
}

type daprSubscription struct {
    PubsubName string `json:"pubsubname"`
    Topic      string `json:"topic"`
    Route      string `json:"route"`
}

type daprPubsubEvent struct {
    Data json.RawMessage `json:"data"`
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

    var userID uint64
    if userIDStr := c.Query("user_id"); userIDStr != "" {
        var err error
        userID, err = strconv.ParseUint(userIDStr, 10, 64)
        if err != nil {
            c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid user_id"))
            return
        }
    }

    var status *int
    if statusStr := c.Query("status"); statusStr != "" {
        s, err := strconv.Atoi(statusStr)
        if err == nil {
            status = &s
        }
    }

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

    orders, total, err := h.service.ListOrders(c.Request.Context(), userID, status, limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    if orders == nil {
        orders = []models.Order{}
    }

    c.JSON(http.StatusOK, dto.Success(gin.H{
        "list":  orders,
        "total": total,
    }))
}

func (h *OrderHandler) GetOrderStats(c *gin.Context) {
    stats, err := h.service.GetOrderStats(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }
    c.JSON(http.StatusOK, dto.Success(stats))
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

func (h *OrderHandler) DaprSubscribe(c *gin.Context) {
    c.JSON(http.StatusOK, []daprSubscription{
        {
            PubsubName: "order-pubsub",
            Topic:      events.TopicOrderPaid,
            Route:      "/events/order-paid",
        },
        {
            PubsubName: "order-pubsub",
            Topic:      events.TopicInventoryReserveFailed,
            Route:      "/events/inventory-reserve-failed",
        },
    })
}

func (h *OrderHandler) HandleOrderPaid(c *gin.Context) {
    var message daprPubsubEvent
    if err := c.ShouldBindJSON(&message); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
        return
    }

    event, err := decodeOrderPaidEvent(message.Data)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
        return
    }

    fmt.Printf("received order-paid event: order_id=%d order_no=%s pay_method=%s\n", event.OrderID, event.OrderNo, event.PayMethod)

    if err := h.service.HandleOrderPaid(c.Request.Context(), event); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.Success(nil))
}

func (h *OrderHandler) HandleInventoryReserveFailed(c *gin.Context) {
    var message daprPubsubEvent
    if err := c.ShouldBindJSON(&message); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
        return
    }

    event, err := decodeInventoryReserveFailedEvent(message.Data)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
        return
    }

    fmt.Printf("received inventory-reserve-failed event: order_no=%s reason=%s\n", event.OrderNo, event.Reason)

    if err := h.service.HandleInventoryReserveFailed(c.Request.Context(), event); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
        return
    }

    c.JSON(http.StatusOK, dto.Success(nil))
}

func decodeOrderPaidEvent(data json.RawMessage) (*events.OrderPaidEvent, error) {
    var event events.OrderPaidEvent
    if err := json.Unmarshal(data, &event); err == nil {
        return &event, nil
    }

    var payload string
    if err := json.Unmarshal(data, &payload); err != nil {
        return nil, err
    }

    if err := json.Unmarshal([]byte(payload), &event); err != nil {
        return nil, err
    }

    return &event, nil
}

func decodeInventoryReserveFailedEvent(data json.RawMessage) (*events.InventoryReserveFailedEvent, error) {
    var event events.InventoryReserveFailedEvent
    if err := json.Unmarshal(data, &event); err == nil {
        return &event, nil
    }

    var payload string
    if err := json.Unmarshal(data, &payload); err != nil {
        return nil, err
    }

    if err := json.Unmarshal([]byte(payload), &event); err != nil {
        return nil, err
    }

    return &event, nil
}
