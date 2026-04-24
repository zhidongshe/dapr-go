FROM golang:1.26-alpine AS builder

# Enable Go toolchain auto-upgrade
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Copy shared module first (from root context)
COPY shared/ /shared/

# Copy go mod files
COPY order-service/go.mod order-service/go.sum ./

# Update go.mod to use correct path for shared module
RUN sed -i 's|=> ../shared|=> /shared|g' go.mod

# Download dependencies (toolchain will auto-upgrade if needed)
RUN go mod download

# Copy source code
COPY order-service/ .

# Tidy and build
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o order-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/order-service .
EXPOSE 8080
CMD ["./order-service"]
