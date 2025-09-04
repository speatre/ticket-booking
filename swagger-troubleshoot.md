# Swagger Troubleshooting Guide

## ✅ Fixed Issues

1. **Missing Swagger Metadata**: Added proper annotations to `cmd/server/main.go`
2. **Regenerated Documentation**: Updated `docs/docs.go` with correct API info
3. **Proper SwaggerInfo**: Set correct host, basePath, title, and description

## 🚀 How to Access Swagger

### Step 1: Start the Server
```bash
# Option 1: Run directly
go run cmd/server/main.go

# Option 2: Build and run
go build -o bin/server cmd/server/main.go
./bin/server
```

### Step 2: Access Swagger UI
Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

## 🔍 Common Issues & Solutions

### Issue 1: "404 Not Found" on Swagger URL
**Cause**: Server not running or wrong URL
**Solution**: 
- Ensure server is running on port 8080
- Check: `http://localhost:8080/swagger/index.html` (note the `/index.html`)

### Issue 2: Empty/Broken Swagger UI
**Cause**: Outdated swagger docs
**Solution**: Regenerate docs
```bash
# Windows
.\generate-swagger.bat

# Linux/Mac
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

### Issue 3: Server Won't Start
**Cause**: Missing dependencies (PostgreSQL, Redis, RabbitMQ)
**Solution**: 
1. Start required services:
   ```bash
   # Using Docker Compose
   docker-compose up -d postgres redis rabbitmq
   ```
2. Or update `configs/app.yaml` to use different endpoints

### Issue 4: API Routes Not Showing
**Cause**: Missing Swagger annotations in handlers
**Solution**: Handlers already have proper annotations, docs should show all routes

## 📋 Expected API Endpoints

The Swagger UI should show these endpoint groups:

### Public Endpoints
- `GET /api/v1/events` - List events
- `GET /api/v1/events/{id}` - Get event by ID
- `POST /api/v1/users/register` - Register user
- `POST /api/v1/users/login` - Login user
- `POST /api/v1/users/refresh` - Refresh token

### Protected Endpoints (Require JWT)
- `POST /api/v1/bookings` - Create booking
- `GET /api/v1/bookings/{id}` - Get booking
- `PUT /api/v1/users/{id}` - Update profile

### Admin Endpoints (Require Admin Role)
- `POST /api/v1/admin/events` - Create event
- `PUT /api/v1/admin/events/{id}` - Update event
- `DELETE /api/v1/admin/events/{id}` - Delete event

## 🔧 Development Mode

For development, you can run without external dependencies by:
1. Commenting out database operations in main.go
2. Using mock services
3. Running: `go run cmd/server/main.go`

## ✅ Verification Checklist

- [ ] Server starts without errors
- [ ] Can access `http://localhost:8080/swagger/index.html`
- [ ] Swagger UI loads with "Ticket Booking API" title
- [ ] All API endpoints are visible
- [ ] Can expand and test endpoints
- [ ] Authentication section shows Bearer token option

## 📞 Still Having Issues?

If Swagger still doesn't load:
1. Check server logs for errors
2. Verify port 8080 is not in use: `netstat -an | findstr 8080`
3. Try different browser or incognito mode
4. Check firewall/antivirus settings

