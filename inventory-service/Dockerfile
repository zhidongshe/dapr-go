FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy inventory-service files
COPY inventory-service/go.mod inventory-service/go.sum ./inventory-service/
COPY inventory-service/ ./inventory-service/

WORKDIR /app/inventory-service

RUN go mod tidy
RUN go build -o inventory-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/inventory-service/inventory-service .
CMD ["./inventory-service"]
