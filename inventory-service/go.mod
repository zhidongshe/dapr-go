module github.com/dapr-oms/inventory-service

go 1.21

require (
	github.com/dapr-oms/shared v0.0.0
	github.com/dapr/go-sdk v1.9.1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-sql-driver/mysql v1.7.1
)

replace github.com/dapr-oms/shared => ../shared
