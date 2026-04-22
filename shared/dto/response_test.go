package dto

import (
	"testing"
)

func TestSuccess(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected Response
	}{
		{
			name: "with string data",
			data: "test message",
			expected: Response{
				Code:    0,
				Message: "success",
				Data:    "test message",
			},
		},
		{
			name: "with int data",
			data: 42,
			expected: Response{
				Code:    0,
				Message: "success",
				Data:    42,
			},
		},
		{
			name: "with nil data",
			data: nil,
			expected: Response{
				Code:    0,
				Message: "success",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Success(tt.data)
			if result.Code != tt.expected.Code {
				t.Errorf("Success() Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Success() Message = %v, want %v", result.Message, tt.expected.Message)
			}
			if result.Data != tt.expected.Data {
				t.Errorf("Success() Data = %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		expected Response
	}{
		{
			name:    "parameter error",
			code:    1001,
			message: "invalid parameter",
			expected: Response{
				Code:    1001,
				Message: "invalid parameter",
				Data:    nil,
			},
		},
		{
			name:    "not found error",
			code:    1002,
			message: "order not found",
			expected: Response{
				Code:    1002,
				Message: "order not found",
				Data:    nil,
			},
		},
		{
			name:    "system error",
			code:    5000,
			message: "internal server error",
			expected: Response{
				Code:    5000,
				Message: "internal server error",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Error(tt.code, tt.message)
			if result.Code != tt.expected.Code {
				t.Errorf("Error() Code = %v, want %v", result.Code, tt.expected.Code)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("Error() Message = %v, want %v", result.Message, tt.expected.Message)
			}
		})
	}
}
