#!/bin/bash

# ===========================================
# TICKET BOOKING - DUMMY DATA LOADER
# ===========================================
# This script loads dummy data into the PostgreSQL database
# for development and testing purposes
#
# Usage:
#   ./load-dummy-data.sh          # Load all dummy data
#   ./load-dummy-data.sh --fresh  # Drop and recreate database first
# ===========================================

set -e

# Configuration
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"ticket_booking"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"postgres"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running and get container name
get_postgres_container() {
    if docker ps | grep -q postgres; then
        echo "$(docker ps | grep postgres | head -1 | awk '{print $NF}')"
    fi
}

# Function to execute SQL commands
execute_sql() {
    local sql="$1"
    local description="$2"

    log_info "$description..."

    if [ -n "$POSTGRES_CONTAINER" ]; then
        # Using Docker container
        echo "$sql" | docker exec -i "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" 2>/dev/null
    else
        # Using local PostgreSQL
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$sql" 2>/dev/null
    fi
}

# Function to execute SQL file
execute_sql_file() {
    local file="$1"
    local description="$2"

    log_info "$description..."

    if [ -n "$POSTGRES_CONTAINER" ]; then
        # Using Docker container
        docker exec -i "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < "$file" 2>/dev/null
    else
        # Using local PostgreSQL
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" < "$file" 2>/dev/null
    fi
}

# Check database connection
check_db_connection() {
    log_info "Checking database connection..."

    if ! execute_sql "SELECT 1;" "Testing database connection" >/dev/null 2>&1; then
        log_error "Cannot connect to database. Please check:"
        echo "  - Is PostgreSQL running?"
        echo "  - Are connection parameters correct?"
        echo "  - Does the database '$DB_NAME' exist?"
        echo ""
        echo "Connection details:"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  User: $DB_USER"
        exit 1
    fi

    log_success "Database connection successful"
}

# Fresh database setup
fresh_setup() {
    log_warning "Performing fresh database setup..."

    # Drop and recreate database
    execute_sql "DROP DATABASE IF EXISTS $DB_NAME;" "Dropping existing database"
    execute_sql "CREATE DATABASE $DB_NAME;" "Creating new database"

    # Run migrations
    if [ -f "migrations/001_init.sql" ]; then
        execute_sql_file "migrations/001_init.sql" "Running database migrations"
        log_success "Database migrations completed"
    else
        log_error "Migration file not found: migrations/001_init.sql"
        exit 1
    fi
}

# Load dummy data
load_dummy_data() {
    if [ ! -f "dummy-data.sql" ]; then
        log_error "Dummy data file not found: dummy-data.sql"
        exit 1
    fi

    execute_sql_file "dummy-data.sql" "Loading dummy data"
    log_success "Dummy data loaded successfully"
}

# Display data summary
show_summary() {
    log_info "Fetching data summary..."

    local summary_query="
    SELECT
        json_build_object(
            'users', (SELECT COUNT(*) FROM users),
            'admin_users', (SELECT COUNT(*) FROM users WHERE role = 'ADMIN'),
            'regular_users', (SELECT COUNT(*) FROM users WHERE role = 'USER'),
            'events', (SELECT COUNT(*) FROM events),
            'upcoming_events', (SELECT COUNT(*) FROM events WHERE starts_at > NOW()),
            'past_events', (SELECT COUNT(*) FROM events WHERE starts_at <= NOW()),
            'bookings', (SELECT COUNT(*) FROM bookings),
            'confirmed_bookings', (SELECT COUNT(*) FROM bookings WHERE status = 'CONFIRMED'),
            'pending_bookings', (SELECT COUNT(*) FROM bookings WHERE status = 'PENDING'),
            'cancelled_bookings', (SELECT COUNT(*) FROM bookings WHERE status = 'CANCELLED')
        ) as summary;
    "

    local result
    if [ -n "$POSTGRES_CONTAINER" ]; then
        result=$(docker exec -i "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "$summary_query" 2>/dev/null)
    else
        result=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "$summary_query" 2>/dev/null)
    fi

    echo ""
    echo "==========================================="
    echo "📊 DUMMY DATA SUMMARY"
    echo "==========================================="

    # Parse JSON result and display nicely
    if command -v jq >/dev/null 2>&1; then
        echo "$result" | jq -r '
            "👥 Total Users: \(.users)",
            "👨‍💼 Admin Users: \(.admin_users)",
            "👤 Regular Users: \(.regular_users)",
            "",
            "🎫 Total Events: \(.events)",
            "🔮 Upcoming Events: \(.upcoming_events)",
            "📅 Past Events: \(.past_events)",
            "",
            "🎟️  Total Bookings: \(.bookings)",
            "✅ Confirmed Bookings: \(.confirmed_bookings)",
            "⏳ Pending Bookings: \(.pending_bookings)",
            "❌ Cancelled Bookings: \(.cancelled_bookings)"
        '
    else
        echo "$result" | sed 's/[{}]/\n/g' | sed 's/:/: /g' | sed 's/"/ /g' | sed 's/,//g'
    fi

    echo "==========================================="
}

# Main execution
main() {
    echo "==========================================="
    echo "🎫 TICKET BOOKING - DUMMY DATA LOADER"
    echo "==========================================="

    # Check if running with Docker
    POSTGRES_CONTAINER=$(get_postgres_container)
    if [ -n "$POSTGRES_CONTAINER" ]; then
        log_info "Found PostgreSQL container: $POSTGRES_CONTAINER"
    fi

    # Check database connection
    check_db_connection

    # Handle fresh setup
    if [ "$1" = "--fresh" ]; then
        fresh_setup
    fi

    # Load dummy data
    load_dummy_data

    # Show summary
    show_summary

    echo ""
    log_success "🎉 Dummy data loading completed!"
    echo ""
    echo "Next steps:"
    echo "  1. Start the Go application: go run cmd/server/main.go"
    echo "  2. Access Swagger UI: http://localhost:8080/swagger/index.html"
    echo "  3. Test the API with the loaded dummy data"
    echo ""
    echo "Sample accounts:"
    echo "  Admin: admin@ticketbooking.com / admin123"
    echo "  User:  john.doe@example.com / password123"
}

# Handle script arguments
case "$1" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Load dummy data into the ticket booking database"
        echo ""
        echo "Options:"
        echo "  --fresh    Drop and recreate database before loading data"
        echo "  --help     Show this help message"
        echo ""
        echo "Environment variables:"
        echo "  DB_HOST        Database host (default: localhost)"
        echo "  DB_PORT        Database port (default: 5432)"
        echo "  DB_NAME        Database name (default: ticket_booking)"
        echo "  DB_USER        Database user (default: postgres)"
        echo "  DB_PASSWORD    Database password (default: postgres)"
        exit 0
        ;;
    *)
        main "$1"
        ;;
esac
