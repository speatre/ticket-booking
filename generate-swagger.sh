#!/usr/bin/env bash
set -euo pipefail

echo "📚 Generating Swagger documentation..."

echo "Installing swag tool..."
go install github.com/swaggo/swag/cmd/swag@latest

echo "Generating docs..."
$(go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

echo "✅ Swagger docs generated!"
echo ""
echo "🚀 To view Swagger UI:"
echo "1. Start the server: go run cmd/server/main.go"
echo "2. Open: http://localhost:8080/swagger/index.html"
echo ""

