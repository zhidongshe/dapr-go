.PHONY: build up down logs test test-unit test-api clean

# Build all services
build:
	docker-compose build

# Run unit tests
test-unit:
	cd shared && go test ./... -v
	cd order-service && go test ./models/... -v
	cd payment-service && go test ./models/... -v

# Run all tests
test: test-unit

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Test order creation
test-create-order:
	curl -X POST http://localhost:8080/api/v1/orders \
		-H "Content-Type: application/json" \
		-d '{"user_id":10001,"items":[{"product_id":101,"product_name":"iPhone 15","unit_price":5999,"quantity":1}]}'

# Test get order
test-get-order:
	curl http://localhost:8080/api/v1/orders/1

# Test list orders
test-list-orders:
	curl "http://localhost:8080/api/v1/orders?user_id=10001"

# Test payment
test-payment:
	curl -X POST http://localhost:8081/api/v1/payments \
		-H "Content-Type: application/json" \
		-d '{"order_no":"ORD202504220001","pay_method":"alipay"}'

# Test cancel order
test-cancel-order:
	curl -X POST http://localhost:8080/api/v1/orders/1/cancel \
		-H "Content-Type: application/json" \
		-d '{"reason":"test cancellation"}'

# Clean up
clean:
	docker-compose down -v
	docker system prune -f
