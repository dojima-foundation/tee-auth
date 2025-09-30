# 🚀 GAuth Service Postman Collections

This directory contains comprehensive Postman collections for testing the GAuth service API and the Renclave-v2 secure enclave service, including both HTTP REST endpoints and gRPC endpoints.

## 📁 Files Overview

- **`gauth-rest-api.postman_collection.json`** - GAuth HTTP REST API collection
- **`gauth-grpc.postman_collection.json`** - GAuth gRPC API collection
- **`gauth-rest-api.postman_environment.json`** - GAuth REST API environment variables
- **`gauth-grpc.postman_environment.json`** - GAuth gRPC API environment variables
- **`renclave-v2-api.postman_collection.json`** - Renclave-v2 secure enclave API collection
- **`renclave-v2-api.postman_environment.json`** - Renclave-v2 API environment variables
- **`README.md`** - This documentation file

## 🎯 Quick Start

### 1. Import Collections into Postman

1. Open Postman
2. Click **Import** button
3. Import the following files:
   - `gauth-rest-api.postman_collection.json`
   - `gauth-grpc.postman_collection.json`
   - `renclave-v2-api.postman_collection.json`
   - `gauth-rest-api.postman_environment.json`
   - `gauth-grpc.postman_environment.json`
   - `renclave-v2-api.postman_environment.json`

### 2. Set Up Environment

1. Select the appropriate environment from the environment dropdown:
   - **"GAuth REST API Environment"** for REST API testing
   - **"GAuth gRPC API Environment"** for gRPC API testing
   - **"Renclave-v2 API Environment"** for enclave service testing
2. Verify the base URLs are set to your service endpoints:
   - GAuth REST: `http://localhost:8082` (default)
   - GAuth gRPC: `localhost:9090` (default)
   - Renclave-v2: `http://localhost:3000` (default)

### 3. Start Testing

1. **Health Check**: Start with the health check endpoints to verify services are running
2. **Create Organization**: Create your first organization
3. **Create User**: Add users to the organization
4. **Create Wallet**: Create wallets with accounts
5. **Create Private Keys**: Generate private keys for signing
6. **Authentication**: Test authentication flows
7. **Seed Generation**: Test cryptographic operations via Renclave-v2

## 📊 Collection Structure

### GAuth REST API Collection

#### 🔍 Health & Status
- **Health Check** - Verify service health and dependencies
- **Service Status** - Get detailed service information and metrics

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

#### 💰 Wallets
- **Create Wallet** - Create new wallet with accounts
- **Get Wallet** - Retrieve wallet details
- **Update Wallet** - Modify wallet information
- **List Wallets** - Paginated list of wallets in organization
- **Delete Wallet** - Delete wallet (with export option)

#### 🔑 Private Keys
- **Create Private Key** - Create new private key in wallet
- **Get Private Key** - Retrieve private key details
- **Update Private Key** - Modify private key information
- **List Private Keys** - Paginated list of private keys in organization
- **Delete Private Key** - Delete private key (with export option)

#### 📝 Activities
- **Create Activity** - Log audit trail activities
- **Get Activity** - Retrieve activity details
- **List Activities** - Paginated list with filtering

#### 🔐 Authentication
- **Authenticate** - Get session token with signature
- **Authorize** - Verify permissions for specific actions

### GAuth gRPC API Collection

The gRPC collection provides the same functionality as the HTTP collection but uses gRPC protocol for direct service communication.

### Renclave-v2 API Collection

#### 🔍 Health & Status
- **Health Check** - Verify enclave service health
- **Get Enclave Info** - Get service information and capabilities
- **Enclave Info** - Get detailed enclave information

#### 🌱 Seed Management
- **Generate Seed** - Generate BIP39 seed phrases (128-256 bits)
- **Validate Seed** - Validate seed phrase format and checksum

#### 🔑 Key Derivation
- **Derive Key** - Derive private keys from seed phrases using BIP32 paths

#### 📍 Address Derivation
- **Derive Address** - Derive public addresses from seed phrases using BIP32 paths

#### 🌐 Network Management
- **Network Status** - Get network connectivity status
- **Test Connectivity** - Test network connectivity to specific hosts

## 🔧 Environment Variables

### GAuth REST/gRPC Variables
- `base_url` / `grpc_host` - Service base URL (default: `http://localhost:8082` / `localhost`)
- `grpc_port` - gRPC service port (default: `9090`)

### Dynamic Variables (Auto-populated)
- `organization_id` - Created organization ID
- `user_id` - Created user ID
- `wallet_id` - Created wallet ID
- `private_key_id` - Created private key ID
- `activity_id` - Created activity ID
- `session_token` - Authentication session token
- `page_token` - Pagination token
- `timestamp` - Current timestamp for authentication

### Test Data Variables
- `test_organization_name` - Default organization name
- `test_user_email` - Default user email
- `test_user_public_key` - Default user public key
- `wallet_name` - Default wallet name
- `private_key_name` - Default private key name
- `activity_type` - Default activity type
- `parameters` - Default activity parameters

### Renclave-v2 Variables
- `renclave_base_url` - Enclave service URL (default: `http://localhost:3000`)
- `test_seed_phrase` - Sample BIP39 seed phrase for testing
- `test_encrypted_seed_phrase` - Sample encrypted seed phrase (hex-encoded) for testing
- `test_derivation_path` - Sample BIP32 derivation path
- `test_curve` - Sample cryptographic curve
- `test_passphrase` - Sample passphrase for seed generation

## 🧪 Testing Workflows

### Basic Service Testing
1. **Health Check** → Verify services are running
2. **Service Status** → Check version and metrics
3. **Enclave Health** → Verify enclave connectivity

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

### Wallet Management
1. **Create Wallet** → Create wallet with accounts
2. **Get Wallet** → Verify wallet creation
3. **List Wallets** → View organization wallets
4. **Create Private Key** → Generate signing keys
5. **List Private Keys** → View available keys

### Cryptographic Operations
1. **Generate Seed** → Create new encrypted seed phrase via Renclave-v2
2. **Validate Seed** → Verify seed phrase validity (supports both encrypted and plain seeds)
3. **Derive Address** → Generate addresses from encrypted seed phrase
4. **Derive Key** → Generate encrypted private keys from encrypted seed phrase

### Authentication Flow
1. **Authenticate** → Get session token
2. **Authorize** → Verify permissions
3. **Create Activity** → Log audit trail

## 🔐 Encrypted API Structure

The Renclave-v2 API now uses encrypted seed phrases and private keys for enhanced security:

### Key Changes
- **Derive Key API**: Now accepts `encrypted_seed_phrase` and returns encrypted `private_key`
- **Derive Address API**: Now accepts `encrypted_seed_phrase` for address derivation
- **Enhanced Security**: All sensitive data is encrypted using quorum keys

### API Request Format
```json
{
  "encrypted_seed_phrase": "hex_encoded_encrypted_seed_phrase",
  "path": "m/44'/60'/0'/0/0",
  "curve": "CURVE_SECP256K1"
}
```

### API Response Format (Derive Key)
```json
{
  "private_key": "hex_encoded_encrypted_private_key",
  "public_key": "hex_encoded_public_key",
  "address": "derived_address",
  "path": "m/44'/60'/0'/0/0",
  "curve": "CURVE_SECP256K1"
}
```

### Environment Variables
- `test_encrypted_seed_phrase`: Use this variable for testing with encrypted seed phrases
- `test_seed_phrase`: Keep for backward compatibility with validate-seed API

## 🔄 Continuous Integration

### Postman CLI Integration
```bash
# Install Newman (Postman CLI)
npm install -g newman

# Run GAuth REST collection tests
newman run gauth-rest-api.postman_collection.json \
  --environment gauth-rest-api.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export results.json

# Run GAuth gRPC collection tests
newman run gauth-grpc.postman_collection.json \
  --environment gauth-grpc.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export grpc-results.json

# Run Renclave-v2 collection tests
newman run renclave-v2-api.postman_collection.json \
  --environment renclave-v2-api.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export renclave-results.json
```

### GitHub Actions Integration
```yaml
- name: API Tests
  run: |
    newman run postman/gauth-rest-api.postman_collection.json \
      --environment postman/gauth-rest-api.postman_environment.json \
      --reporters cli,junit \
      --reporter-junit-export test-results.xml
    
    newman run postman/gauth-grpc.postman_collection.json \
      --environment postman/gauth-grpc.postman_environment.json \
      --reporters cli,junit \
      --reporter-junit-export grpc-test-results.xml
    
    newman run postman/renclave-v2-api.postman_collection.json \
      --environment postman/renclave-v2-api.postman_environment.json \
      --reporters cli,junit \
      --reporter-junit-export renclave-test-results.xml
```

## 📚 Additional Resources

- **GAuth Service Documentation**: See main project README
- **gRPC Testing**: Use gRPC collection for direct service testing
- **Renclave-v2 Documentation**: See renclave-v2 project README
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
- Check if service is running: `curl http://localhost:8082/health`
- Check if gRPC service is running: `grpcurl -plaintext localhost:9090 list`
- Check if Renclave-v2 is running: `curl http://localhost:3000/health`

**Authentication Errors**
- Verify organization_id and user_id are set
- Check if session tokens are valid
- Ensure proper signature format

**gRPC Connection Issues**
- Verify gRPC port is correct (default: 9090)
- Check if gRPC server is running
- Ensure proper TLS configuration if using secure connections

**Enclave Communication Issues**
- Verify Renclave-v2 service is running
- Check network connectivity
- Ensure proper enclave configuration

### Debug Mode
Enable debug logging in your services to get more detailed error information.

## 📋 Testing Checklist

### GAuth Service
- [ ] Health check returns 200
- [ ] Service status includes version info
- [ ] Create organization works
- [ ] Get organization returns correct data
- [ ] Update organization works
- [ ] List organizations with pagination
- [ ] Create user succeeds
- [ ] Get user returns correct data
- [ ] Update user works
- [ ] List users with filtering
- [ ] Create wallet with accounts
- [ ] Get wallet returns correct data
- [ ] Update wallet works
- [ ] List wallets with pagination
- [ ] Create private key succeeds
- [ ] Get private key returns correct data
- [ ] Update private key works
- [ ] List private keys with pagination
- [ ] Authentication succeeds with valid credentials
- [ ] Session token is generated
- [ ] Authorization works with session token
- [ ] Invalid credentials are rejected
- [ ] Create activity succeeds
- [ ] Get activity returns correct data
- [ ] List activities with pagination

### Renclave-v2 Service
- [ ] Health check returns 200
- [ ] Enclave info includes version and capabilities
- [ ] Seed generation succeeds
- [ ] Generated seed is valid BIP39
- [ ] Seed validation works
- [ ] Invalid seeds are rejected
- [ ] Key derivation works
- [ ] Address derivation works
- [ ] Network status is available
- [ ] Connectivity testing works

### Error Handling
- [ ] Invalid UUIDs return 400
- [ ] Missing required fields return 400
- [ ] Unauthorized requests return 401
- [ ] Not found resources return 404
- [ ] Server errors return 500
