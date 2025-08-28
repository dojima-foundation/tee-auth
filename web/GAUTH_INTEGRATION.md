# GAuth API Integration

This document describes the integration of the GAuth API with the ODEYS dashboard.

## Overview

The GAuth API integration allows the dashboard to create organizations and manage users through the GAuth service. The integration is implemented in the Users component.

## API Service

### Location
- `web/src/services/gauthApi.ts`

### Features
- **Create Organization**: Creates a new organization with an initial admin user
- **Get Organizations**: Retrieves list of organizations
- **Get Organization**: Retrieves specific organization details
- **Create User**: Creates a new user within an organization
- **Get Users**: Retrieves users for an organization

### Configuration
The API service uses the following environment variable:
- `NEXT_PUBLIC_GAUTH_API_URL`: Base URL for the GAuth API (defaults to `http://localhost:8082`)

## Integration in Users Component

### Location
- `web/src/components/Users.tsx`

### Functionality
When the "Create Organization" button is clicked, the component:

1. **Calls GAuth API**: Makes a POST request to `/api/v1/organizations`
2. **Request Payload**:
   ```json
   {
     "name": "ODEYS Organization",
     "initial_user_email": "admin@odeys.com",
     "initial_user_public_key": "0x1234567890abcdef1234567890abcdef12345678"
   }
   ```
3. **Handles Response**: Shows success/error messages
4. **Updates UI**: Refreshes the users list after successful creation

### UI Features
- **Loading State**: Shows spinner while creating organization
- **Success Message**: Displays organization ID and user ID on success
- **Error Handling**: Shows error messages if API call fails
- **Button State**: Disables button during API call

## API Endpoints Used

### Create Organization
- **Endpoint**: `POST /api/v1/organizations`
- **Purpose**: Creates a new organization with initial admin user
- **Response**: Organization details, status, and user ID

### Get Users
- **Endpoint**: `GET /api/v1/users?organization_id={id}`
- **Purpose**: Retrieves users for a specific organization
- **Response**: List of users with pagination

## Error Handling

The integration includes comprehensive error handling:
- **Network Errors**: Handles connection issues
- **API Errors**: Displays server error messages
- **Validation Errors**: Shows validation failure messages
- **Timeout Handling**: Manages request timeouts

## Security Considerations

- **Public Key**: Uses a placeholder public key for demonstration
- **Email Validation**: GAuth API validates email format
- **Organization Isolation**: Users are scoped to organizations
- **Error Messages**: Sanitized error messages to prevent information leakage

## Future Enhancements

1. **Dynamic Organization Name**: Allow users to input organization name
2. **Real Public Key**: Integrate with wallet to get actual public key
3. **User Management**: Add create/edit/delete user functionality
4. **Organization Management**: Add organization listing and management
5. **Authentication**: Add proper authentication flow
6. **Real-time Updates**: Add WebSocket support for real-time updates

## Testing

To test the integration:

1. **Start GAuth Server**: Ensure GAuth API is running on `http://localhost:8082`
2. **Navigate to Users**: Go to `/dashboard/users`
3. **Click Create Organization**: Test the organization creation flow
4. **Check Response**: Verify success/error messages
5. **Verify Data**: Check if users list updates after creation

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure GAuth server is running
2. **CORS Errors**: Check GAuth server CORS configuration
3. **Validation Errors**: Verify request payload format
4. **Timeout Errors**: Check network connectivity and server response time

### Debug Steps

1. Check browser console for error messages
2. Verify GAuth server logs
3. Test API endpoint directly with curl/Postman
4. Check environment variable configuration
