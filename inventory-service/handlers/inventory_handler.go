package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dapr-oms/inventory-service/repository"
	"github.com/dapr-oms/inventory-service/services"
	"github.com/dapr-oms/shared/dto"
	"github.com/dapr-oms/shared/events"
	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	service     *services.InventoryService
	messageRepo *repository.MessageRepository
}

type daprSubscription struct {
	PubsubName string `json:"pubsubname"`
	Topic      string `json:"topic"`
	Route      string `json:"route"`
}

type daprPubsubMessage struct {
	Data json.RawMessage `json:"data"`
}

func NewInventoryHandler() *InventoryHandler {
	service := services.NewInventoryService()

	// Initialize message repository for idempotency
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(mysql:3306)/oms_db?charset=utf8mb4&parseTime=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database for message repository: %v", err))
	}
	messageRepo := repository.NewMessageRepository(db)

	return &InventoryHandler{
		service:     service,
		messageRepo: messageRepo,
	}
}

// GetInventory handles GET /api/v1/inventory/:product_id
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, "invalid product id"))
		return
	}

	inv, err := h.service.GetInventory(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}

	if inv == nil {
		c.JSON(http.StatusNotFound, dto.Error(1002, "inventory not found"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(inv))
}

func (h *InventoryHandler) ListAllInventory(c *gin.Context) {
	list, err := h.service.ListAllInventory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(5000, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(list))
}

// DaprSubscribe returns Dapr subscription configuration
func (h *InventoryHandler) DaprSubscribe(c *gin.Context) {
	subscriptions := []daprSubscription{
		{
			PubsubName: "order-pubsub",
			Topic:      events.TopicInventoryReserve,
			Route:      "/events/inventory-reserve",
		},
		{
			PubsubName: "order-pubsub",
			Topic:      events.TopicInventoryConfirm,
			Route:      "/events/inventory-confirm",
		},
		{
			PubsubName: "order-pubsub",
			Topic:      events.TopicInventoryRelease,
			Route:      "/events/inventory-release",
		},
	}
	c.JSON(http.StatusOK, subscriptions)
}

// HandleReserve processes inventory reserve events
func (h *InventoryHandler) HandleReserve(c *gin.Context) {
	var msg daprPubsubMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	event, err := decodeReserveEvent(msg.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	// Check if message already processed
	isProcessed, err := h.messageRepo.IsProcessed(event.MessageID)
	if err != nil {
		fmt.Printf("check message processed failed: %v\n", err)
	}
	if isProcessed {
		fmt.Printf("message %s already processed, skip\n", event.MessageID)
		c.JSON(http.StatusOK, dto.Success(nil))
		return
	}

	fmt.Printf("received inventory-reserve event: order_no=%s\n", event.OrderNo)

	if err := h.service.ReserveStock(c.Request.Context(), event); err != nil {
		fmt.Printf("reserve stock failed: %v\n", err)
		// Publish failure event
		h.service.HandleReserveFailed(c.Request.Context(), event, err.Error())
		// Return 200 to acknowledge the message
		c.JSON(http.StatusOK, dto.Success(nil))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// HandleConfirm processes inventory confirm events
func (h *InventoryHandler) HandleConfirm(c *gin.Context) {
	var msg daprPubsubMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	event, err := decodeConfirmEvent(msg.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	fmt.Printf("received inventory-confirm event: order_no=%s\n", event.OrderNo)

	if err := h.service.ConfirmStock(c.Request.Context(), event); err != nil {
		fmt.Printf("confirm stock failed: %v\n", err)
		// Return 200 to avoid retry, but log the error
		c.JSON(http.StatusOK, dto.Success(nil))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// HandleRelease processes inventory release events
func (h *InventoryHandler) HandleRelease(c *gin.Context) {
	var msg daprPubsubMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	event, err := decodeReleaseEvent(msg.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(1001, err.Error()))
		return
	}

	fmt.Printf("received inventory-release event: order_no=%s\n", event.OrderNo)

	if err := h.service.ReleaseStock(c.Request.Context(), event); err != nil {
		fmt.Printf("release stock failed: %v\n", err)
		c.JSON(http.StatusOK, dto.Success(nil))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// decode functions
func decodeReserveEvent(data json.RawMessage) (*events.InventoryReserveEvent, error) {
	var event events.InventoryReserveEvent
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

func decodeConfirmEvent(data json.RawMessage) (*events.InventoryConfirmEvent, error) {
	var event events.InventoryConfirmEvent
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

func decodeReleaseEvent(data json.RawMessage) (*events.InventoryReleaseEvent, error) {
	var event events.InventoryReleaseEvent
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
