package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterProductRoutes_UsesAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register product routes
	RegisterProductRoutes(r.Group("/api"))

	// Test that routes require auth (should return 401 without auth header)
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/products"},
		{"GET", "/api/products/123"},
		{"POST", "/api/products"},
		{"PUT", "/api/products/123/price"},
		{"PUT", "/api/products/123/status"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		r.ServeHTTP(w, req)
		// Should get 401 Unauthorized (middleware blocks request)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Route %s %s should require auth", tt.method, tt.path)
	}
}

func TestUpdateProductPrice_ForwardsPutRequest(t *testing.T) {
	// This test verifies the PUT forwarding capability exists
	// The actual forwarding is tested via integration tests
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register product routes
	RegisterProductRoutes(r.Group("/api"))

	// Verify the route is registered by checking it returns 401 (auth required)
	// rather than 404 (route not found)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/products/123/price", nil)
	r.ServeHTTP(w, req)

	// Should get 401 (auth middleware blocks) not 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code, "PUT /api/products/:id/price route should be registered")
}
