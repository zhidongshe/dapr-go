package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		DisableKeepAlives: false,
	},
}

func getServiceURL(serviceName string) string {
	switch serviceName {
	case "order":
		if url := os.Getenv("ORDER_SERVICE_URL"); url != "" {
			return url
		}
		return "http://localhost:8080"
	case "payment":
		if url := os.Getenv("PAYMENT_SERVICE_URL"); url != "" {
			return url
		}
		return "http://localhost:8081"
	case "inventory":
		if url := os.Getenv("INVENTORY_SERVICE_URL"); url != "" {
			return url
		}
		return "http://localhost:8082"
	case "product":
		if url := os.Getenv("PRODUCT_SERVICE_URL"); url != "" {
			return url
		}
		return "http://localhost:8083"
	default:
		return ""
	}
}

// ForwardRequest 转发请求到后端服务
func ForwardRequest(serviceName, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	baseURL := getServiceURL(serviceName)
	if baseURL == "" {
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}

	url := baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// 复制 headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return httpClient.Do(req)
}

// ForwardGET 转发 GET 请求
func ForwardGET(serviceName, path string, headers map[string]string) (*http.Response, error) {
	return ForwardRequest(serviceName, "GET", path, nil, headers)
}

// ForwardPOST 转发 POST 请求
func ForwardPOST(serviceName, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return ForwardRequest(serviceName, "POST", path, body, headers)
}

// ForwardPUT 转发 PUT 请求
func ForwardPUT(serviceName, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return ForwardRequest(serviceName, "PUT", path, body, headers)
}
