APP_NAME=ticket-booking
DOCKER_COMPOSE_FILE=docker-compose.yml
GO_CMD=go
GOCMD_BUILD=$(GO_CMD) build
BINARY=bin/$(APP_NAME)

.PHONY: all build docker-image up logs down restart migrate seed dummy-gen dummy-load dummy test test-race test-cover test-full deps-mockgen mocks vet ci

# ---- Build Go binary ----
build:
	@echo "🔨 Building Go binary..."
	mkdir -p bin
	$(GOCMD_BUILD) -o $(BINARY) ./cmd/server

# ---- Build Docker image ----
docker-image: build
	@echo "🐳 Building Docker image..."
	docker build -t $(APP_NAME) .

# ---- Start services ----
up: docker-image
	@echo "🚀 Starting all services..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

# ---- Tail logs ----
logs:
	@echo "📖 Tailing logs..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# ---- Stop services ----
down:
	@echo "🛑 Stopping services..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

# ---- Restart workflow ----
restart: down docker-image up logs

# ---- DB migrations ----
migrate:
	@echo "🗄️  Running DB migrations..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec app $(BINARY) migrate

# ---- Seed data ----
seed:
	@echo "🌱 Seeding initial data..."
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec app $(BINARY) seed

# ---- Load dummy data ----
dummy-load:
	@echo "📥 Loading dummy data..."
	./load-dummy-data.sh


# ---- Run full workflow: build -> up -> logs ----
run: docker-image up logs

# ---- Dependencies: mockgen ----
deps-mockgen:
	@echo "📦 Installing mockgen..."
	$(GO_CMD) install go.uber.org/mock/mockgen@latest
	@echo "✅ mockgen installed at $$($(GO_CMD) env GOPATH)/bin/mockgen"

# ---- Regenerate mocks ----
mocks: deps-mockgen
	@echo "🧪 Regenerating mocks..."
	$$($(GO_CMD) env GOPATH)/bin/mockgen -source="internal/booking/repository.go" -destination="internal/mocks/mock_booking_repository.go" -package=mocks
	$$($(GO_CMD) env GOPATH)/bin/mockgen -source="internal/event/repository.go"   -destination="internal/mocks/mock_event_repository.go"   -package=mocks
	$$($(GO_CMD) env GOPATH)/bin/mockgen -source="pkg/mq/rabbit.go"               -destination="internal/mocks/mock_rabbit.go"             -package=mocks
	$$($(GO_CMD) env GOPATH)/bin/mockgen -source="pkg/cache/redis.go"             -destination="internal/mocks/mock_redis.go"              -package=mocks
	$$($(GO_CMD) env GOPATH)/bin/mockgen -source="internal/database/database.go"   -destination="internal/mocks/mock_database.go"           -package=mocks
	$$($(GO_CMD) env GOPATH)/bin/mockgen -destination="internal/mocks/event_reserver.go" -package=mocks ticket-booking/internal/booking EventReserver
	@echo "✅ Mocks regenerated"

# ---- Static checks ----
vet:
	@echo "🔎 Running go vet..."
	$(GO_CMD) vet ./...

# ---- Tests ----
test:
	@echo "🧪 Running tests..."
	$(GO_CMD) test ./...

test-race:
	@echo "🧪 Running tests with race detector..."
	$(GO_CMD) test -race ./...

test-cover:
	@echo "🧪 Running tests with coverage..."
	$(GO_CMD) test -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -func=coverage.out

# ---- Full test suite (recommended) ----
test-full:
	@echo "🧪 Running comprehensive test suite..."
	./scripts/run-tests.sh

# ---- CI target (mocks + vet + tests with coverage) ----
ci: mocks vet test-cover
