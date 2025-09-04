#!/usr/bin/env bash
set -euo pipefail

echo "📦 Ensuring mockgen..."
echo "Installing latest mockgen version..."
go install go.uber.org/mock/mockgen@latest
export PATH="$(go env GOPATH)/bin:$PATH"

echo "🧪 Regenerating mocks..."
echo "Generating booking repository mock..."
mockgen -source=internal/booking/repository.go -destination=internal/mocks/mock_booking_repository.go -package=mocks

echo "Generating event repository mock..."
mockgen -source=internal/event/repository.go -destination=internal/mocks/mock_event_repository.go -package=mocks

echo "Generating rabbit MQ mock..."
mockgen -source=pkg/mq/rabbit.go -destination=internal/mocks/mock_rabbit.go -package=mocks

echo "Generating redis cache mock..."
mockgen -source=pkg/cache/redis.go -destination=internal/mocks/mock_redis.go -package=mocks

echo "Generating database mock..."
mockgen -source=internal/database/database.go -destination=internal/mocks/mock_database.go -package=mocks

echo "Generating event reserver mock..."
mockgen -destination=internal/mocks/event_reserver.go -package=mocks ticket-booking/internal/booking EventReserver

echo "🔎 go vet..."
go vet ./...

echo "🧪 go test with coverage..."
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

echo "✅ Done"

