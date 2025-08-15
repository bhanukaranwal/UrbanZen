# UrbanZen API Documentation

## Overview

UrbanZen provides a comprehensive RESTful API for managing IoT devices, monitoring city infrastructure, and accessing analytics data. All API endpoints are secured with JWT authentication and follow REST conventions.

## Base URL

```
Production: https://api.urbanzen.gov.in/api/v1
Development: http://localhost:8080/api/v1
```

## Authentication

All protected endpoints require a Bearer token in the Authorization header:

```http
Authorization: Bearer <jwt_token>
```

### Getting a Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "admin"
  }
}
```

## API Endpoints

### Device Management

#### List Devices
```http
GET /api/v1/devices
```

Query Parameters:
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)
- `type` (optional): Filter by device type
- `status` (optional): Filter by device status
- `location` (optional): Filter by location (lat,lng,radius)

Response:
```json
{
  "devices": [
    {
      "id": 1,
      "device_id": "WM001",
      "name": "Water Meter - Sector 15",
      "type": "water_meter",
      "status": "active",
      "location": {
        "lat": 28.4595,
        "lng": 77.0266
      },
      "last_seen": "2024-01-01T12:00:00Z",
      "metadata": {}
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "total_items": 500
  }
}
```

#### Create Device
```http
POST /api/v1/devices
Content-Type: application/json

{
  "device_id": "WM002",
  "name": "Water Meter - Sector 16",
  "type_id": 1,
  "location": {
    "lat": 28.4595,
    "lng": 77.0266
  },
  "address": "Sector 16, Block A, Gurgaon",
  "configuration": {
    "measurement_interval": 60,
    "alert_threshold": 1000
  }
}
```

#### Get Device Details
```http
GET /api/v1/devices/{device_id}
```

#### Update Device
```http
PUT /api/v1/devices/{device_id}
Content-Type: application/json

{
  "name": "Updated Device Name",
  "status": "maintenance",
  "configuration": {}
}
```

#### Send Command to Device
```http
POST /api/v1/devices/{device_id}/command
Content-Type: application/json

{
  "command": "reboot",
  "parameters": {
    "delay": 30
  }
}
```

#### Get Device Telemetry
```http
GET /api/v1/devices/{device_id}/telemetry
```

Query Parameters:
- `start_time`: ISO 8601 timestamp
- `end_time`: ISO 8601 timestamp
- `metrics`: Comma-separated list of metrics
- `aggregation`: minute, hour, day (default: raw)

### Analytics

#### Get Consumption Analytics
```http
GET /api/v1/analytics/consumption
```

Query Parameters:
- `utility_type`: water, electricity, gas
- `start_date`: YYYY-MM-DD
- `end_date`: YYYY-MM-DD
- `aggregation`: hourly, daily, monthly
- `zone_id` (optional): Filter by zone

#### Get Anomaly Detection Results
```http
GET /api/v1/analytics/anomalies
```

#### Get Predictive Insights
```http
GET /api/v1/analytics/predictions
```

### User Management

#### List Users
```http
GET /api/v1/users
```

Requires: `admin` role

#### Create User
```http
POST /api/v1/users
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "secure_password",
  "first_name": "John",
  "last_name": "Doe",
  "role": "citizen",
  "phone": "+91-9876543210"
}
```

#### Get User Profile
```http
GET /api/v1/users/profile
```

#### Update User Profile
```http
PUT /api/v1/users/profile
Content-Type: application/json

{
  "first_name": "Updated Name",
  "phone": "+91-9876543211"
}
```

### Billing

#### List Utility Accounts
```http
GET /api/v1/billing/accounts
```

#### Get Billing History
```http
GET /api/v1/billing/accounts/{account_id}/bills
```

#### Generate Bill
```http
POST /api/v1/billing/accounts/{account_id}/generate-bill
Content-Type: application/json

{
  "reading_end": 1500.5,
  "billing_period": "2024-01"
}
```

### Notifications

#### List Notifications
```http
GET /api/v1/notifications
```

#### Send Notification
```http
POST /api/v1/notifications
Content-Type: application/json

{
  "user_id": 123,
  "type": "alert",
  "channel": "email",
  "title": "Water Leak Alert",
  "message": "A water leak has been detected in your area.",
  "data": {
    "location": "Sector 15, Block A",
    "severity": "high"
  }
}
```

### Reporting

#### Generate System Report
```http
POST /api/v1/reports/generate
Content-Type: application/json

{
  "report_type": "consumption",
  "parameters": {
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "utility_type": "water",
    "format": "pdf"
  }
}
```

#### Download Report
```http
GET /api/v1/reports/{report_id}/download
```

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "email",
      "issue": "Invalid email format"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Rate Limiting

API requests are rate limited:
- **Authenticated users**: 1000 requests per hour
- **Public endpoints**: 100 requests per hour

Rate limit headers are included in responses:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Webhooks

UrbanZen can send webhooks for important events:

### Configuration
```http
POST /api/v1/webhooks
Content-Type: application/json

{
  "url": "https://your-app.com/webhook",
  "events": ["device.offline", "alert.created", "bill.generated"],
  "secret": "webhook_secret_key"
}
```

### Payload Format
```json
{
  "event": "device.offline",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "device_id": "WM001",
    "device_name": "Water Meter - Sector 15",
    "last_seen": "2024-01-01T11:45:00Z"
  }
}
```

## SDKs and Libraries

Official SDKs are available for:
- **JavaScript/TypeScript**: `npm install @urbanzen/sdk`
- **Python**: `pip install urbanzen-sdk`
- **Go**: `go get github.com/urbanzen/go-sdk`
- **Java**: Available on Maven Central

## Support

- **Documentation**: https://docs.urbanzen.gov.in
- **API Status**: https://status.urbanzen.gov.in
- **Support Email**: api-support@urbanzen.gov.in
- **GitHub Issues**: https://github.com/bhanukaranwal/UrbanZen/issues