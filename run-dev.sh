#!/usr/bin/env bash
set -euo pipefail

echo "🚀 Starting Ticket Booking Server in Development Mode..."
echo ""
echo "📚 Swagger UI will be available at: http://localhost:8080/swagger/index.html"
echo "🔍 API Base URL: http://localhost:8080/api/v1"
echo ""
echo "⚠️  Note: Make sure PostgreSQL, Redis, and RabbitMQ are running"
echo "    Or use Docker: docker-compose up -d postgres redis rabbitmq"
echo ""
echo "Starting server..."
go run cmd/server/main.go

