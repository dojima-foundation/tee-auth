# gAuth REST API

This directory contains the REST API implementation for the gAuth service. The REST API acts as a gateway that translates HTTP requests to gRPC calls internally.

## Architecture

```
HTTP Client → REST API → gRPC Client → gRPC Server → Business Logic
```

The REST API server:
1. Receives HTTP requests
2. Validates and transforms request payloads
3. Makes gRPC calls to the internal gRPC server
4. Transforms gRPC responses back to JSON
5. Returns HTTP responses

## API Endpoints

### Health & Status
- `GET /api/v1/health` - Service health check
- `GET /api/v1/status` - Service status information

### Organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/:id` - Get organization by ID
- `PUT /api/v1/organizations/:id` - Update organization
- `GET /api/v1/organizations` - List organizations (with pagination)

### Users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `GET /api/v1/users?organization_id=:id` - List users (with pagination)

### Activities
- `POST /api/v1/activities` - Create activity
- `GET /api/v1/activities/:id` - Get activity by ID
- `GET /api/v1/activities?organization_id=:id` - List activities (with pagination, filtering)

### Authentication
- `POST /api/v1/auth/authenticate` - Authenticate user
- `POST /api/v1/auth/authorize` - Authorize action

### Renclave Integration
- `GET /api/v1/renclave/info` - Get enclave information
- `POST /api/v1/renclave/seed/generate` - Generate seed phrase
- `POST /api/v1/renclave/seed/validate` - Validate seed phrase

## Configuration

The REST API server uses the same configuration as the gRPC server, with additional settings:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"

security:
  cors_enabled: true
  cors_origins: ["*"]
  rate_limit_enabled: true
  rate_limit_rps: 100
  rate_limit_burst: 200
```

## Request/Response Format

### Standard Response Format

All successful responses follow this format:
```json
{
  "success": true,
  "data": {
    // Response data here
  }
}
```

Error responses follow this format:
```json
{
  "error": "Error message",
  "details": "Detailed error information"
}
```

### Pagination

List endpoints support pagination with these query parameters:
- `page_size`: Number of items per page (default: 10, max: 100)
- `page_token`: Token for the next page

Paginated responses include a `next_page_token` field when there are more pages.

## Example Usage

### Create Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Organization",
    "initial_user_email": "admin@example.com",
    "initial_user_public_key": "public-key-here"
  }'
```

### Get Health Status
```bash
curl http://localhost:8080/api/v1/health
```

### List Users with Pagination
```bash
curl "http://localhost:8080/api/v1/users?organization_id=org-id&page_size=20"
```

## Middleware

The REST API includes several middleware components:

1. **Logging Middleware**: Logs all HTTP requests
2. **Recovery Middleware**: Handles panics gracefully
3. **CORS Middleware**: Handles Cross-Origin Resource Sharing
4. **Rate Limiting Middleware**: Prevents abuse (configurable)

## Testing

Run REST API tests:
```bash
make test-rest
```

The tests include:
- Unit tests for individual components
- Integration tests with actual gRPC backend
- Error handling scenarios
- Request validation

## Development

To run the REST API in development:
```bash
# Start both gRPC and REST servers
make run

# Or run with live reload
make dev
```

The REST API will be available at `http://localhost:8080/api/v1/`

## Error Handling

The REST API handles errors from the gRPC backend and translates them to appropriate HTTP status codes:

- `400 Bad Request` - Invalid request payload or parameters
- `401 Unauthorized` - Authentication failed
- `403 Forbidden` - Authorization failed
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server-side errors
- `503 Service Unavailable` - Service health issues
