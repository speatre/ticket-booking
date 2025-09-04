#!/bin/bash
set -euo pipefail

APP_NAME="ticket-booking"

echo "🔨 Step 1: Build Docker image..."
docker build -t $APP_NAME .

echo "🚀 Step 2: Start all services..."
docker-compose up -d

echo "📖 Step 3: Tail app logs (Ctrl+C to exit)..."
docker-compose logs -f app

echo "🛑 Step 4: Shutdown services..."
docker-compose down
