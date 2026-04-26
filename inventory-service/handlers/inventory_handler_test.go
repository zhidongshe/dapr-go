package handlers

import (
	"encoding/json"
	"testing"

	"github.com/dapr-oms/shared/events"
)

func TestDecodeReserveEvent(t *testing.T) {
	// Test with direct JSON object
	eventData := events.InventoryReserveEvent{
		MessageID: "msg-123",
		OrderID:   1,
		OrderNo:   "ORD20250101120000",
		UserID:    1001,
		Items: []events.InventoryItem{
			{ProductID: 1, ProductName: "iPhone 16", Quantity: 2},
		},
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	decoded, err := decodeReserveEvent(jsonBytes)
	if err != nil {
		t.Errorf("decodeReserveEvent failed: %v", err)
	}

	if decoded.MessageID != eventData.MessageID {
		t.Error("MessageID mismatch")
	}

	if decoded.OrderNo != eventData.OrderNo {
		t.Error("OrderNo mismatch")
	}

	if len(decoded.Items) != len(eventData.Items) {
		t.Error("Items length mismatch")
	}
}

func TestDecodeReserveEventWithStringPayload(t *testing.T) {
	// Test with string-encoded JSON (Dapr sometimes sends this format)
	eventData := events.InventoryReserveEvent{
		MessageID: "msg-456",
		OrderID:   2,
		OrderNo:   "ORD20250101120001",
		UserID:    1002,
		Items: []events.InventoryItem{
			{ProductID: 2, ProductName: "MacBook", Quantity: 1},
		},
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Wrap in a string
	stringWrapped, err := json.Marshal(string(jsonBytes))
	if err != nil {
		t.Fatalf("Failed to wrap in string: %v", err)
	}

	decoded, err := decodeReserveEvent(stringWrapped)
	if err != nil {
		t.Errorf("decodeReserveEvent with string payload failed: %v", err)
	}

	if decoded.MessageID != eventData.MessageID {
		t.Error("MessageID mismatch in string payload test")
	}
}

func TestDecodeConfirmEvent(t *testing.T) {
	eventData := events.InventoryConfirmEvent{
		MessageID: "msg-789",
		OrderID:   1,
		OrderNo:   "ORD20250101120000",
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	decoded, err := decodeConfirmEvent(jsonBytes)
	if err != nil {
		t.Errorf("decodeConfirmEvent failed: %v", err)
	}

	if decoded.MessageID != eventData.MessageID {
		t.Error("MessageID mismatch")
	}

	if decoded.OrderNo != eventData.OrderNo {
		t.Error("OrderNo mismatch")
	}
}

func TestDecodeReleaseEvent(t *testing.T) {
	eventData := events.InventoryReleaseEvent{
		MessageID: "msg-abc",
		OrderID:   1,
		OrderNo:   "ORD20250101120000",
		Reason:    "order cancelled",
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	decoded, err := decodeReleaseEvent(jsonBytes)
	if err != nil {
		t.Errorf("decodeReleaseEvent failed: %v", err)
	}

	if decoded.MessageID != eventData.MessageID {
		t.Error("MessageID mismatch")
	}

	if decoded.Reason != eventData.Reason {
		t.Error("Reason mismatch")
	}
}

func TestDaprSubscriptionStructure(t *testing.T) {
	// Verify subscription structure is properly defined
	sub := daprSubscription{
		PubsubName: "order-pubsub",
		Topic:      events.TopicInventoryReserve,
		Route:      "/events/inventory-reserve",
	}

	if sub.PubsubName != "order-pubsub" {
		t.Error("PubsubName mismatch")
	}

	if sub.Topic != events.TopicInventoryReserve {
		t.Errorf("Topic should be %s", events.TopicInventoryReserve)
	}
}

func TestDaprPubsubMessageStructure(t *testing.T) {
	// Verify message structure
	msg := daprPubsubMessage{
		Data: json.RawMessage(`{"order_id": 1}`),
	}

	if len(msg.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestDecodeEventWithInvalidData(t *testing.T) {
	// Test with invalid JSON
	invalidData := json.RawMessage(`{invalid json}`)

	_, err := decodeReserveEvent(invalidData)
	if err == nil {
		t.Error("decodeReserveEvent should return error for invalid JSON")
	}

	_, err = decodeConfirmEvent(invalidData)
	if err == nil {
		t.Error("decodeConfirmEvent should return error for invalid JSON")
	}

	_, err = decodeReleaseEvent(invalidData)
	if err == nil {
		t.Error("decodeReleaseEvent should return error for invalid JSON")
	}
}

func TestDecodeReserveEventEmptyItems(t *testing.T) {
	// Test with empty items array
	eventData := events.InventoryReserveEvent{
		MessageID: "msg-empty",
		OrderID:   1,
		OrderNo:   "ORD20250101120000",
		UserID:    1001,
		Items:     []events.InventoryItem{},
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	decoded, err := decodeReserveEvent(jsonBytes)
	if err != nil {
		t.Errorf("decodeReserveEvent failed: %v", err)
	}

	if len(decoded.Items) != 0 {
		t.Error("Items should be empty")
	}
}

func TestDecodeReserveEventMultipleItems(t *testing.T) {
	// Test with multiple items
	eventData := events.InventoryReserveEvent{
		MessageID: "msg-multi",
		OrderID:   1,
		OrderNo:   "ORD20250101120000",
		UserID:    1001,
		Items: []events.InventoryItem{
			{ProductID: 1, ProductName: "iPhone 16", Quantity: 2},
			{ProductID: 2, ProductName: "MacBook Pro", Quantity: 1},
			{ProductID: 3, ProductName: "AirPods", Quantity: 3},
		},
	}

	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	decoded, err := decodeReserveEvent(jsonBytes)
	if err != nil {
		t.Errorf("decodeReserveEvent failed: %v", err)
	}

	if len(decoded.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(decoded.Items))
	}

	// Verify all items
	for i, item := range decoded.Items {
		if item.ProductID != eventData.Items[i].ProductID {
			t.Errorf("Item %d ProductID mismatch", i)
		}
		if item.Quantity != eventData.Items[i].Quantity {
			t.Errorf("Item %d Quantity mismatch", i)
		}
	}
}
