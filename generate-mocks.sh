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
if [ -f "internal/event/repository.go" ]; then
    mockgen -source=internal/event/repository.go -destination=internal/mocks/mock_event_repository.go -package=mocks
else
    echo "WARNING: internal/event/repository.go not found, skipping..."
fi

echo "Generating rabbit MQ mock..."
if [ -f "pkg/mq/rabbit.go" ]; then
    mockgen -source=pkg/mq/rabbit.go -destination=internal/mocks/mock_rabbit.go -package=mocks
else
    echo "WARNING: pkg/mq/rabbit.go not found, skipping..."
fi

echo "Generating redis cache mock..."
if [ -f "pkg/cache/redis.go" ]; then
    mockgen -source=pkg/cache/redis.go -destination=internal/mocks/mock_redis.go -package=mocks
else
    echo "WARNING: pkg/cache/redis.go not found, skipping..."
fi

echo "Generating database mock..."
mockgen -source=internal/database/database.go -destination=internal/mocks/mock_database.go -package=mocks

echo "Generating event reserver mock..."
mockgen -destination=internal/mocks/event_reserver.go -package=mocks ticket-booking/internal/booking EventReserver

echo "✅ Mocks regenerated successfully!"
echo ""
echo "🧪 Running tests to verify..."
go test ./internal/booking/
echo ""
echo "✅ All done!"
