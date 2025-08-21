# 🚀 GAuth Service Postman Collections

This directory contains comprehensive Postman collections for testing the GAuth service API, including both HTTP REST endpoints and gRPC endpoints.

## 📁 Files Overview

- **`gauth-service.postman_collection.json`** - Main HTTP REST API collection
- **`gauth-grpc.postman_collection.json`** - gRPC API collection
- **`gauth-service.postman_environment.json`** - Environment variables
- **`README.md`** - This documentation file

## 🎯 Quick Start

### 1. Import Collections into Postman

1. Open Postman
2. Click **Import** button
3. Import the following files:
   - `gauth-service.postman_collection.json`
   - `gauth-grpc.postman_collection.json`
   - `gauth-service.postman_environment.json`

### 2. Set Up Environment

1. Select the **"GAuth Service Environment"** from the environment dropdown
2. Verify the base URL is set to your service endpoint (default: `http://localhost:8080`)
3. For gRPC testing, verify the gRPC URL is set (default: `localhost:9090`)

### 3. Start Testing

1. **Health Check**: Start with the health check endpoint to verify service is running
2. **Create Organization**: Create your first organization
3. **Create User**: Add users to the organization
4. **Authentication**: Test authentication flows
5. **Seed Generation**: Test cryptographic operations

## 📊 Collection Structure

### HTTP REST API Collection

#### 🔍 Health & Status
- **Health Check** - Verify service health and dependencies
- **Service Status** - Get detailed service information

#### 🏢 Organizations
- **Create Organization** - Create new organization with admin user
- **Get Organization** - Retrieve organization details
- **Update Organization** - Modify organization information
- **List Organizations** - Paginated list of organizations

#### 👥 Users
- **Create User** - Add new user to organization
- **Get User** - Retrieve user details
- **Update User** - Modify user information
- **List Users** - Paginated list of users in organization

#### 🔐 Authentication
- **Authenticate User** - Get session token
- **Authorize Action** - Verify permissions for specific actions

#### 📝 Activities
- **Create Activity** - Log audit trail activities
- **Get Activity** - Retrieve activity details
- **List Activities** - Paginated list with filtering

#### 🔑 Seed Generation
- **Request Seed Generation** - Generate BIP39 seed phrases via TEE
- **Validate Seed** - Validate seed phrase format and checksum

#### 🛡️ TEE Integration
- **Get Enclave Info** - TEE enclave information
- **Enclave Health Check** - TEE health status

### gRPC API Collection

The gRPC collection provides the same functionality as the HTTP collection but uses gRPC protocol for direct service communication.

## 🔧 Environment Variables

### Core Variables
- `base_url` - HTTP API base URL (default: `http://localhost:8080`)
- `grpc_url` - gRPC service URL (default: `localhost:9090`)

### Dynamic Variables (Auto-populated)
- `organization_id` - Created organization ID
- `user_id` - Created user ID
- `activity_id` - Created activity ID
- `session_token` - Authentication session token
- `page_token` - Pagination token
- `timestamp` - Current timestamp for authentication

### Test Data Variables
- `test_organization_name` - Default organization name
- `test_user_email` - Default user email
- `test_user_public_key` - Default user public key
- `test_seed_phrase` - Sample BIP39 seed phrase
- `test_signature` - Sample signature for testing

## 🧪 Testing Workflows

### Basic Service Testing
1. **Health Check** → Verify service is running
2. **Service Status** → Check version and metrics
3. **TEE Health** → Verify enclave connectivity

### Organization Management
1. **Create Organization** → Set up new organization
2. **Get Organization** → Verify creation
3. **List Organizations** → View all organizations
4. **Update Organization** → Modify details

### User Management
1. **Create User** → Add user to organization
2. **Get User** → Verify user creation
3. **List Users** → View organization users
4. **Update User** → Modify user details

### Authentication Flow
1. **Create Organization** → Set up organization
2. **Create User** → Add user
3. **Authenticate User** → Get session token
4. **Authorize Action** → Test permissions

### Seed Generation Workflow
1. **Create Organization** → Set up organization
2. **Create User** → Add user
3. **Authenticate User** → Get session token
4. **Request Seed Generation** → Generate seed phrase
5. **Validate Seed** → Verify generated seed

## 🔄 Automated Testing Features

### Pre-request Scripts
- **Auto-timestamp**: Sets current timestamp for authentication
- **Auto-UUID**: Generates UUIDs for testing if not present
- **Environment Setup**: Prepares test data

### Test Scripts
- **Status Validation**: Verifies response codes (200/201)
- **Content Type Check**: Ensures JSON responses
- **Performance Check**: Validates response times
- **ID Extraction**: Automatically stores IDs from responses
- **Token Management**: Captures session tokens

### Environment Management
- **Dynamic Variables**: Auto-populated from responses
- **Test Data**: Pre-configured test values
- **State Persistence**: Maintains context across requests

## 🚨 Error Handling

### Common Error Scenarios
- **Service Unavailable**: Check if service is running
- **Authentication Failed**: Verify credentials and timestamps
- **Invalid UUID**: Check ID format and existence
- **Permission Denied**: Verify user permissions and session tokens

### Debugging Tips
1. Check service logs for detailed error messages
2. Verify environment variables are set correctly
3. Ensure database and Redis are running
4. Check TEE enclave connectivity for seed operations

## 📈 Performance Testing

### Response Time Expectations
- **Health Checks**: < 100ms
- **CRUD Operations**: < 500ms
- **Seed Generation**: < 2000ms (depends on TEE)
- **List Operations**: < 1000ms

### Load Testing
- Use Postman's **Runner** feature for concurrent testing
- Monitor response times and error rates
- Test with various data sizes and pagination

## 🔒 Security Testing

### Authentication Testing
- Test with invalid credentials
- Verify session token expiration
- Test authorization with different user roles
- Validate signature verification

### Input Validation
- Test with malformed JSON
- Verify UUID format validation
- Test with invalid email formats
- Validate seed phrase format

## 🛠️ Setup Instructions

### Prerequisites
1. **GAuth Service Running**
   ```bash
   # Build and start the service
   make build-local
   export GRPC_PORT=9090
   export DB_HOST=localhost
   export REDIS_HOST=localhost
   ./bin/gauth
   ```

2. **Database Setup**
   ```bash
   # PostgreSQL
   docker run -d --name gauth-postgres \
     -e POSTGRES_USER=gauth \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=gauth \
     -p 5432:5432 postgres:15

   # Redis
   docker run -d --name gauth-redis \
     -p 6379:6379 redis:7
   ```

3. **TEE Integration** (Optional)
   ```bash
   # Start renclave-v2 service for seed generation
   docker run -d --name gauth-renclave \
     -p 3000:3000 renclave-v2:latest
   ```

### Environment Configuration
1. **Development**
   - `base_url`: `http://localhost:8080`
   - `grpc_url`: `localhost:9090`

2. **Staging**
   - `base_url`: `https://staging-gauth.example.com`
   - `grpc_url`: `staging-gauth.example.com:9090`

3. **Production**
   - `base_url`: `https://gauth.example.com`
   - `grpc_url`: `gauth.example.com:9090`

## 📋 Testing Checklist

### Service Health
- [ ] Health check returns 200
- [ ] Service status shows correct version
- [ ] TEE health check passes
- [ ] Response times are acceptable

### Organization Management
- [ ] Create organization succeeds
- [ ] Get organization returns correct data
- [ ] Update organization works
- [ ] List organizations with pagination

### User Management
- [ ] Create user succeeds
- [ ] Get user returns correct data
- [ ] Update user works
- [ ] List users with filtering

### Authentication
- [ ] Authentication succeeds with valid credentials
- [ ] Session token is generated
- [ ] Authorization works with session token
- [ ] Invalid credentials are rejected

### Seed Generation
- [ ] Seed generation succeeds
- [ ] Generated seed is valid BIP39
- [ ] Seed validation works
- [ ] Invalid seeds are rejected

### Error Handling
- [ ] Invalid UUIDs return 400
- [ ] Missing required fields return 400
- [ ] Unauthorized requests return 401
- [ ] Not found resources return 404

## 🔄 Continuous Integration

### Postman CLI Integration
```bash
# Install Newman (Postman CLI)
npm install -g newman

# Run collection tests
newman run gauth-service.postman_collection.json \
  --environment gauth-service.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export results.json
```

### GitHub Actions Integration
```yaml
- name: API Tests
  run: |
    newman run postman/gauth-service.postman_collection.json \
      --environment postman/gauth-service.postman_environment.json \
      --reporters cli,junit \
      --reporter-junit-export test-results.xml
```

## 📚 Additional Resources

- **GAuth Service Documentation**: See main project README
- **gRPC Testing**: Use gRPC collection for direct service testing
- **Performance Monitoring**: Monitor response times and error rates
- **Security Testing**: Validate authentication and authorization flows

## 🤝 Contributing

When adding new endpoints or modifying existing ones:

1. **Update Collections**: Add new requests to appropriate folders
2. **Update Environment**: Add new variables if needed
3. **Update Tests**: Add appropriate test scripts
4. **Update Documentation**: Keep this README current

## 🆘 Troubleshooting

### Common Issues

**Service Not Responding**
- Check if service is running: `curl http://localhost:8080/health`
- Verify port configuration
- Check service logs

**Database Connection Issues**
- Verify PostgreSQL is running
- Check connection credentials
- Ensure database exists

**gRPC Connection Issues**
- Verify gRPC service is running
- Check port configuration
- Ensure protobuf definitions are current

**TEE Integration Issues**
- Verify renclave-v2 service is running
- Check enclave health endpoint
- Ensure proper network connectivity

For additional support, please refer to the main project documentation or create an issue in the repository.
