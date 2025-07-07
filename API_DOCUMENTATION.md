# BikunTracker Backend API Documentation

## Overview
BikunTracker is a real-time bus tracking system for campus shuttle buses. This API provides endpoints for managing buses, tracking lap history, authentication, and real-time WebSocket connections.

**Base URL**: `http://localhost:8080` (development)

## Table of Contents
1. [Authentication](#authentication)
2. [Bus Management](#bus-management)
3. [Lap Tracking](#lap-tracking)
4. [Real-time WebSocket](#real-time-websocket)
5. [Data Models](#data-models)
6. [Error Handling](#error-handling)

---

## Authentication

### POST `/auth/sso/login`
Login using Single Sign-On (SSO) with UI CAS system.

**Request Body:**
```json
{
  "ticket": "string",
  "service": "string"
}
```

**Response:**
```json
{
  "access": "jwt_access_token",
  "refresh": "jwt_refresh_token",
  "user": {
    "id": 1,
    "name": "John Doe",
    "npm": "1234567890",
    "email": "john@ui.ac.id",
    "role": "admin",
    "created_at": 1640995200,
    "updated_at": 1640995200
  }
}
```

### POST `/auth/refresh`
Refresh JWT access token using refresh token.

**Request Body:**
```json
{
  "refresh": "jwt_refresh_token"
}
```

**Response:**
```json
{
  "access": "new_jwt_access_token",
  "refresh": "new_jwt_refresh_token"
}
```

### GET `/auth/me`
Get current authenticated user information.

**Headers:**
- `Authorization: Bearer <jwt_access_token>`

**Response:**
```json
{
  "id": 1,
  "name": "John Doe",
  "npm": "1234567890",
  "email": "john@ui.ac.id",
  "role": "admin",
  "created_at": 1640995200,
  "updated_at": 1640995200
}
```

---

## Bus Management

### GET `/bus`
Get all buses.

**Response:**
```json
[
  {
    "id": 1,
    "vehicle_no": "UI-001",
    "imei": "123456789012345",
    "is_active": true,
    "color": "blue",
    "current_halte": "Asrama UI",
    "next_halte": "Menwa",
    "created_at": 1640995200,
    "updated_at": 1640995200
  }
]
```

### POST `/bus`
Create a new bus.

**Headers:**
- `X-API-Key: <admin_api_key>` (Admin only)

**Request Body:**
```json
{
  "vehicle_no": "UI-001",
  "imei": "123456789012345",
  "is_active": true,
  "color": "blue",
  "current_halte": "Asrama UI",
  "next_halte": "Menwa"
}
```

**Response:**
```json
{
  "id": 1,
  "vehicle_no": "UI-001",
  "imei": "123456789012345",
  "is_active": true,
  "color": "blue",
  "current_halte": "Asrama UI",
  "next_halte": "Menwa",
  "created_at": 1640995200,
  "updated_at": 1640995200
}
```

### PUT `/bus/:id`
Update a bus by ID.

**Headers:**
- `X-API-Key: <admin_api_key>` (Admin only)

**Request Body:**
```json
{
  "vehicle_no": "UI-001-Updated",
  "is_active": false,
  "color": "red"
}
```

**Response:**
```json
{
  "id": 1,
  "vehicle_no": "UI-001-Updated",
  "imei": "123456789012345",
  "is_active": false,
  "color": "red",
  "current_halte": "Asrama UI",
  "next_halte": "Menwa",
  "created_at": 1640995200,
  "updated_at": 1640995300
}
```

### DELETE `/bus/:id`
Delete a bus by ID.

**Headers:**
- `X-API-Key: <admin_api_key>` (Admin only)

**Response:**
```json
{
  "message": "Bus deleted successfully"
}
```

---

## Lap Tracking

### GET `/bus/lap-history`
Get filtered lap history with pagination.

**Query Parameters:**
- `imei` (optional): Filter by specific bus IMEI
- `bus_id` (optional): Filter by bus ID
- `route_color` (optional): Filter by route color (blue, red, grey, etc.)
- `from_date` (optional): Filter from start date (ISO 8601 format)
- `to_date` (optional): Filter to end date (ISO 8601 format)
- `start_time` (optional): Filter by time of day (HH:MM format)
- `end_time` (optional): Filter by time of day (HH:MM format)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Number of results per page (default: 10)

**Example:**
```
GET /bus/lap-history?imei=123456789012345&route_color=blue&page=1&limit=20
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "bus_id": 1,
      "imei": "123456789012345",
      "lap_number": 1,
      "start_time": "2024-01-01T08:00:00Z",
      "end_time": "2024-01-01T08:30:00Z",
      "route_color": "blue",
      "created_at": "2024-01-01T08:00:00Z",
      "updated_at": "2024-01-01T08:30:00Z"
    }
  ],
  "hasNext": true,
  "totalPages": 5,
  "currentPage": 1,
  "totalCount": 100
}
```

### GET `/bus/:imei/lap-history`
Get lap history for a specific bus by IMEI.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Number of results per page (default: 10)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "bus_id": 1,
      "imei": "123456789012345",
      "lap_number": 1,
      "start_time": "2024-01-01T08:00:00Z",
      "end_time": "2024-01-01T08:30:00Z",
      "route_color": "blue",
      "created_at": "2024-01-01T08:00:00Z",
      "updated_at": "2024-01-01T08:30:00Z"
    }
  ],
  "hasNext": false,
  "totalPages": 1,
  "currentPage": 1,
  "totalCount": 1
}
```

### GET `/bus/:imei/active-lap`
Get the currently active lap for a specific bus.

**Response:**
```json
{
  "id": 1,
  "bus_id": 1,
  "imei": "123456789012345",
  "lap_number": 2,
  "start_time": "2024-01-01T09:00:00Z",
  "end_time": null,
  "route_color": "blue",
  "created_at": "2024-01-01T09:00:00Z",
  "updated_at": "2024-01-01T09:00:00Z"
}
```

**Response (No Active Lap):**
```json
{
  "message": "No active lap found for this bus"
}
```

---

## Real-time WebSocket

### WebSocket `/ws`
Real-time bus coordinates and operational status.

**Connection:** Upgrade HTTP to WebSocket

**Message Format:**
```json
{
  "coordinates": [
    {
      "id": 1,
      "color": "blue",
      "imei": "123456789012345",
      "vehicle_name": "UI-001",
      "longitude": 106.8456,
      "latitude": -6.3676,
      "status": "moving",
      "speed": 25,
      "total_mileage": 1250.5,
      "gps_time": "2024-01-01T08:00:00Z",
      "current_halte": "Asrama UI",
      "message": "Arriving at Asrama UI",
      "next_halte": "Menwa"
    }
  ],
  "operational_status": {
    "total_buses": 5,
    "active_buses": 3,
    "last_updated": "2024-01-01T08:00:00Z"
  }
}
```

**Frequency:** Updates every 1 second

---

## Data Models

### Bus
```json
{
  "id": "integer",
  "vehicle_no": "string",
  "imei": "string",
  "is_active": "boolean",
  "color": "string",
  "current_halte": "string",
  "next_halte": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### BusCoordinate
```json
{
  "id": "integer",
  "color": "string",
  "imei": "string",
  "vehicle_name": "string",
  "longitude": "float64",
  "latitude": "float64",
  "status": "string",
  "speed": "integer",
  "total_mileage": "float64",
  "gps_time": "timestamp",
  "current_halte": "string",
  "message": "string",
  "next_halte": "string"
}
```

### BusLapHistory
```json
{
  "id": "integer",
  "bus_id": "integer",
  "imei": "string",
  "lap_number": "integer",
  "start_time": "timestamp",
  "end_time": "timestamp|null",
  "route_color": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### User
```json
{
  "id": "integer",
  "name": "string",
  "npm": "string",
  "email": "string",
  "role": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

---

## Error Handling

### Standard Error Response
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

### Common HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

### Common Error Codes
- `INVALID_REQUEST` - Invalid request format
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Data validation failed

---

## Lap Detection Logic

The system automatically detects laps based on bus movement patterns:

### Lap Start
- **Condition:** Bus moves from "Asrama UI" to "Menwa"
- **Behavior:** Creates a new lap record with start time
- **Override:** If an active lap exists, it will be ended and a new one started

### Lap End
- **Conditions:** 
  - Bus reaches "Parking" (final destination)
  - Bus returns to "Asrama UI" from any halte other than "Menwa"
- **Behavior:** Updates the lap record with end time

### Route Colors
- `blue` - Regular blue route
- `red` - Regular red route
- `express-blue` - Express blue route (morning)
- `express-red` - Express red route (morning)
- `grey` - No route detected or inactive

---

## Rate Limiting

- WebSocket connections: No limit
- API endpoints: 100 requests per minute per IP
- Admin operations: 50 requests per minute per API key

---

## Authentication Requirements

### Public Endpoints
- `GET /bus`
- `GET /bus/lap-history`
- `GET /bus/:imei/lap-history`
- `GET /bus/:imei/active-lap`
- `POST /auth/sso/login`
- `POST /auth/refresh`
- WebSocket `/ws`

### JWT Required
- `GET /auth/me`

### Admin API Key Required
- `POST /bus`
- `PUT /bus/:id`
- `DELETE /bus/:id`

---

## Development Notes

### Debug Endpoints (Development Only)
These endpoints are available for testing and should be removed in production:

- `POST /bus/test-lap-data` - Create test lap data
- `GET /bus/check-table` - Check lap history table structure

### Environment Variables
```env
PORT=8080
DB_DSN=postgres://user:password@localhost/bikuntracker
WS_URL=ws://external-gps-provider/ws
WS_UPGRADE_WHITELIST=http://localhost:3000,https://bikuntracker.ui.ac.id
ADMIN_API_KEY=your_admin_api_key
JWT_SECRET=your_jwt_secret
PRINT_CSV_LOGS=false
```

### GPS Data Flow
1. External GPS provider sends data via WebSocket
2. System processes coordinates and detects halte visits
3. Lap detection logic runs on every halte change
4. Data is stored in database and broadcasted via WebSocket
5. Clients receive real-time updates

---

## Support

For support and questions, please contact the development team or create an issue in the project repository.
