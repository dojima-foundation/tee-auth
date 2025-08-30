# Session Management Integration

This document describes the comprehensive session management integration that has been added to the frontend application.

## Overview

The session management integration provides:
- **Automatic session validation and refresh**
- **Real-time session status monitoring**
- **Session management UI components**
- **Cross-device session management**
- **Secure session handling with server-side validation**

## Architecture

### Components

1. **Session API Service** (`src/services/gauthApi.ts`)
   - Session management API methods
   - Automatic token inclusion in requests
   - Cookie-based session handling

2. **Session Middleware** (`src/lib/session-middleware.tsx`)
   - Automatic session refresh (every 5 minutes)
   - Session validation on page visibility changes
   - Session cleanup on page unload

3. **Auth Context Updates** (`src/lib/auth-context.tsx`)
   - Enhanced with session management methods
   - Server-side logout integration
   - Session state management

4. **UI Components**
   - `SessionManager` - Full session management interface
   - `SessionStatus` - Real-time session status indicator
   - Session test page for debugging

## Features

### 1. Automatic Session Management

The `SessionMiddleware` component automatically:
- Refreshes sessions every 5 minutes
- Validates sessions when the user returns to the tab
- Handles session cleanup on page unload
- Logs out users when sessions expire

### 2. Real-time Session Status

The `SessionStatus` component displays:
- Time until session expiry
- Session status (Active/Expired)
- Quick refresh button
- OAuth provider information

### 3. Session Management Interface

The `SessionManager` component provides:
- Current session information
- List of all active sessions
- Session destruction capabilities
- Manual session refresh

### 4. Cross-device Session Management

Users can:
- View all active sessions across devices
- Destroy specific sessions remotely
- Monitor session activity
- See session creation and expiry times

## API Endpoints

The integration uses the following backend endpoints:

- `GET /api/v1/sessions/info` - Get current session information
- `POST /api/v1/sessions/refresh` - Refresh current session
- `POST /api/v1/sessions/logout` - Logout and destroy session
- `GET /api/v1/sessions/validate` - Validate current session
- `GET /api/v1/sessions/list` - List all user sessions
- `DELETE /api/v1/sessions/:id` - Destroy specific session

## Usage

### 1. Basic Integration

The session management is automatically integrated into the app layout:

```tsx
// src/app/layout.tsx
<AuthProvider>
  <SessionMiddleware>
    <SnackbarProvider>
      {children}
    </SnackbarProvider>
  </SessionMiddleware>
</AuthProvider>
```

### 2. Using Session Status

Add session status to any component:

```tsx
import { SessionStatus } from '@/components/SessionStatus';

// Simple status indicator
<SessionStatus />

// Detailed status with information
<SessionStatus showDetails={true} />
```

### 3. Manual Session Management

Use the session management hook:

```tsx
import { useSessionManagement } from '@/lib/session-middleware';

function MyComponent() {
  const { 
    refreshSession, 
    validateSession, 
    getSessionInfo, 
    listSessions 
  } = useSessionManagement();

  const handleRefresh = async () => {
    await refreshSession();
  };
}
```

### 4. Full Session Management UI

Add the complete session management interface:

```tsx
import { SessionManager } from '@/components/SessionManager';

function SessionsPage() {
  return <SessionManager />;
}
```

## Configuration

### Session Refresh Interval

Configure the automatic refresh interval:

```tsx
<SessionMiddleware refreshInterval={10 * 60 * 1000}> // 10 minutes
  {children}
</SessionMiddleware>
```

### Validation on Mount

Control whether to validate sessions on component mount:

```tsx
<SessionMiddleware validateOnMount={false}>
  {children}
</SessionMiddleware>
```

## Security Features

### 1. Server-side Session Validation

All session operations are validated on the server:
- Session tokens are verified
- Expiration times are checked
- User permissions are validated

### 2. Automatic Logout

Users are automatically logged out when:
- Session refresh fails
- Session validation fails
- Session expires

### 3. Secure Token Handling

- Session tokens are stored in localStorage
- Tokens are automatically included in API requests
- Cookies are used for additional security

## Error Handling

The integration includes comprehensive error handling:

- **Network errors**: Graceful fallback to local logout
- **Session expiry**: Automatic user logout
- **Invalid sessions**: Clear error messages and logout
- **API failures**: User-friendly error notifications

## Testing

### Session Test Page

A dedicated test page is available at `/dashboard/session-test` for:
- Testing session operations
- Debugging session issues
- Monitoring session state
- Manual session management

### Manual Testing

1. **Session Refresh**: Click refresh button and verify expiry time updates
2. **Session Validation**: Use test page to validate sessions
3. **Cross-device**: Login from multiple devices and manage sessions
4. **Expiry**: Wait for session to expire and verify automatic logout

## Monitoring

### Session Status Indicators

- **Green**: Session active with >2 hours remaining
- **Yellow**: Session active with <2 hours remaining  
- **Red**: Session expired or <30 minutes remaining

### Logs

Session management operations are logged to the console:
- Session refresh attempts
- Validation results
- Error conditions
- Automatic logout events

## Best Practices

### 1. Session Security

- Always use HTTPS in production
- Implement proper CORS policies
- Use secure cookie settings
- Monitor for suspicious session activity

### 2. User Experience

- Provide clear session status indicators
- Allow manual session refresh
- Show session expiry warnings
- Enable cross-device session management

### 3. Error Handling

- Gracefully handle network failures
- Provide clear error messages
- Implement automatic retry logic
- Log errors for debugging

## Troubleshooting

### Common Issues

1. **Session not refreshing**
   - Check network connectivity
   - Verify API endpoint availability
   - Check browser console for errors

2. **Automatic logout**
   - Verify session hasn't expired
   - Check server-side session validation
   - Review authentication flow

3. **Cross-device issues**
   - Ensure proper CORS configuration
   - Verify session sharing between devices
   - Check session storage implementation

### Debug Tools

- Use the session test page for debugging
- Check browser developer tools for network requests
- Review server logs for session operations
- Monitor localStorage for session data

## Future Enhancements

Potential improvements for the session management system:

1. **Session Analytics**
   - Track session usage patterns
   - Monitor security events
   - Generate session reports

2. **Advanced Security**
   - Device fingerprinting
   - Location-based validation
   - Suspicious activity detection

3. **User Preferences**
   - Configurable session timeouts
   - Remember device options
   - Session notification settings

4. **Performance Optimization**
   - Batch session operations
   - Optimize refresh intervals
   - Implement session caching
