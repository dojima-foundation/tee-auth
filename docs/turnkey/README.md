# Turnkey Documentation for Cursor

This directory contains Turnkey API documentation for use with Cursor's LLM features.

## Files

- `llms-full.txt` - Complete Turnkey API documentation (3MB)
- `llms.txt` - Concise API reference (41KB)
- `README.md` - This file

## How to Use with Cursor

### Method 1: Direct File Reference
When working with Turnkey API in Cursor, you can reference these files directly:

```
Reference: docs/turnkey/llms-full.txt
```

### Method 2: Copy Relevant Sections
Copy specific sections from the documentation files into your code comments:

```go
// Turnkey API Integration
// Reference: docs/turnkey/llms-full.txt
// Endpoint: /public/v1/submit/approve_activity
// Authentication: X-Stamp header required
```

### Method 3: Use in Chat
When asking Cursor about Turnkey API, mention:
- "Check docs/turnkey/llms-full.txt for the complete API reference"
- "See docs/turnkey/llms.txt for a concise overview"

## Key Information

### Authentication
All Turnkey API requests require:
- `X-Stamp` header with cryptographically signed request
- Organization ID
- API keypair (public/private keys)

### Common Endpoints
- `/public/v1/submit/approve_activity`
- `/public/v1/submit/create_api_keys`
- `/public/v1/submit/sign_transaction`
- `/public/v1/submit/create_authenticators`

### Request Format
```json
{
  "type": "ACTIVITY_TYPE_*",
  "timestampMs": "timestamp_in_milliseconds",
  "organizationId": "your_organization_id",
  "parameters": {
    // endpoint-specific parameters
  }
}
```

## Updating Documentation

To update the documentation files:

```bash
curl -o docs/turnkey/llms-full.txt https://docs.turnkey.com/llms-full.txt
curl -o docs/turnkey/llms.txt https://docs.turnkey.com/llms.txt
```

## Online Resources

- [Turnkey Documentation](https://docs.turnkey.com/)
- [API Reference](https://docs.turnkey.com/llms-full.txt)
- [Developer Guide](https://docs.turnkey.com/developer-reference/using-llms)
