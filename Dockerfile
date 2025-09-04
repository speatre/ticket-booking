# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV CGO_ENABLED=0

# Chỉ copy module files trước để tận dụng cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -o ticket-booking ./cmd/server/main.go

# Runtime stage
FROM gcr.io/distroless/base:nonroot
WORKDIR /app
COPY --from=builder /app/ticket-booking .
COPY --from=builder /app/configs ./configs

# Non-root user
USER nonroot:nonroot
EXPOSE 8080
CMD ["./ticket-booking"]
