'use client';

import React, { createContext, useContext, useReducer, useEffect, ReactNode, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { AuthState, AuthSession, LoginCredentials, SessionInfo } from '@/types/auth';
import { gauthApi } from '@/services/gauthApi';
import { setAuthSession, clearAuth } from '@/store/authSlice';

// Type definitions for window object extensions
interface WindowWithMockAuth extends Window {
    __MOCK_AUTH__?: boolean;
}

// Type definitions for session with organization
interface SessionWithOrganization {
    organization_id?: string;
}

// Auth Action Types
type AuthAction =
    | { type: 'AUTH_START' }
    | { type: 'AUTH_SUCCESS'; payload: AuthSession }
    | { type: 'AUTH_FAILURE'; payload: string }
    | { type: 'AUTH_LOGOUT' }
    | { type: 'CLEAR_ERROR' }
    | { type: 'SESSION_REFRESH'; payload: string }
    | { type: 'SESSION_VALIDATE'; payload: SessionInfo };

// Initial state
const initialState: AuthState = {
    isAuthenticated: false,
    user: null,
    session: null,
    loading: true, // Start with loading true to validate session on mount
    error: null,
};

// Auth reducer
function authReducer(state: AuthState, action: AuthAction): AuthState {

    let newState: AuthState;

    switch (action.type) {
        case 'AUTH_START':
            newState = {
                ...state,
                loading: true,
                error: null,
            };
            break;
        case 'AUTH_SUCCESS':
            newState = {
                ...state,
                isAuthenticated: true,
                user: action.payload.user,
                session: action.payload,
                loading: false,
                error: null,
            };
            break;
        case 'AUTH_FAILURE':
            newState = {
                ...state,
                isAuthenticated: false,
                user: null,
                session: null,
                loading: false,
                error: action.payload,
            };
            break;
        case 'AUTH_LOGOUT':
            newState = {
                ...state,
                isAuthenticated: false,
                user: null,
                session: null,
                loading: false,
                error: null,
            };
            break;
        case 'CLEAR_ERROR':
            newState = {
                ...state,
                error: null,
            };
            break;
        case 'SESSION_REFRESH':
            newState = {
                ...state,
                session: state.session ? {
                    ...state.session,
                    expires_at: action.payload,
                } : null,
            };
            break;
        case 'SESSION_VALIDATE':
            newState = {
                ...state,
                isAuthenticated: true,
                loading: false,
                error: null,
            };
            break;
        default:
            newState = state;
    }


    return newState;
}

// Auth context
interface AuthContextType extends AuthState {
    loginWithGoogle: (credentials: LoginCredentials) => Promise<void>;
    handleOAuthCallback: (code: string, state: string) => Promise<void>;
    setSession: (session: AuthSession) => void;
    logout: () => void;
    clearError: () => void;
    refreshSession: () => Promise<void>;
    validateSession: () => Promise<boolean>;
    getSessionInfo: () => Promise<SessionInfo | null>;
    listSessions: () => Promise<SessionInfo[]>;
    destroySession: (sessionId: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
interface AuthProviderProps {
    children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [state, dispatch] = useReducer(authReducer, initialState);
    const reduxDispatch = useDispatch();

    // Check for existing session on mount and validate with backend
    useEffect(() => {
        const initializeSession = async () => {

            // Check if we're in test mode with mock authentication enabled
            const isTestMode = process.env.NODE_ENV === 'test' || process.env.NEXT_PUBLIC_TEST_MODE === 'true';
            const mockAuthEnabled = process.env.NEXT_PUBLIC_MOCK_AUTH === 'true' ||
                (typeof window !== 'undefined' && (window as WindowWithMockAuth).__MOCK_AUTH__);

            if (isTestMode && mockAuthEnabled) {
                // Create mock session data for testing
                const mockSession: AuthSession = {
                    session_token: 'test-session-token',
                    user: {
                        id: 'test-user-id',
                        email: 'test@example.com',
                        username: 'testuser',
                        organization_id: 'test-org-id',
                        public_key: 'test-public-key',
                        is_active: true,
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                    expires_at: (Date.now() + 24 * 60 * 60 * 1000).toString(), // 24 hours from now
                    auth_method: {
                        id: 'test-auth-method-id',
                        type: 'mock',
                        name: 'Mock Authentication'
                    }
                };

                dispatch({ type: 'AUTH_SUCCESS', payload: mockSession });
                reduxDispatch(setAuthSession(mockSession));
                return;
            }

            const sessionToken = localStorage.getItem('gauth_session_token');
            const sessionData = localStorage.getItem('gauth_session_data');


            if (sessionToken && sessionData) {
                try {
                    const session: AuthSession = JSON.parse(sessionData);


                    // Try to parse the expires_at value
                    let expiresAt: Date;

                    if (typeof session.expires_at === 'string') {
                        // Try parsing as timestamp first (if it's a number string)
                        if (/^\d+$/.test(session.expires_at)) {
                            expiresAt = new Date(parseInt(session.expires_at));
                        } else {
                            // Try parsing as ISO string
                            expiresAt = new Date(session.expires_at);
                        }
                    } else if (typeof session.expires_at === 'number') {
                        expiresAt = new Date(session.expires_at);
                    } else {
                        console.error('❌ [AuthProvider] Invalid expires_at type:', typeof session.expires_at, session.expires_at);
                        localStorage.removeItem('gauth_session_token');
                        localStorage.removeItem('gauth_session_data');
                        dispatch({ type: 'AUTH_FAILURE', payload: 'Invalid session data' });
                        return;
                    }

                    // Check if the date is valid
                    if (isNaN(expiresAt.getTime())) {
                        console.error('❌ [AuthProvider] Invalid expires_at value after parsing:', session.expires_at, '→', expiresAt);
                        localStorage.removeItem('gauth_session_token');
                        localStorage.removeItem('gauth_session_data');
                        dispatch({ type: 'AUTH_FAILURE', payload: 'Invalid session data' });
                        return;
                    }



                    // Check if session is expired locally first
                    if (expiresAt <= new Date()) {
                        localStorage.removeItem('gauth_session_token');
                        localStorage.removeItem('gauth_session_data');
                        dispatch({ type: 'AUTH_FAILURE', payload: 'Session expired' });
                        return;
                    }

                    // Validate session with backend
                    dispatch({ type: 'AUTH_START' });

                    const response = await gauthApi.validateSession();

                    if (response.success) {
                        // Session is valid, update with fresh data from backend
                        const updatedSession: AuthSession = {
                            ...session,
                            expires_at: new Date(response.data.expires_at).getTime().toString(),
                            user: {
                                ...session.user,
                                organization_id: response.data.organization_id, // Update organization_id from backend
                            }
                        };

                        dispatch({ type: 'AUTH_SUCCESS', payload: updatedSession });

                        // Update Redux store with fresh session data
                        reduxDispatch(setAuthSession(updatedSession));

                        // Update local storage with fresh data
                        localStorage.setItem('gauth_session_data', JSON.stringify(updatedSession));
                    } else {
                        localStorage.removeItem('gauth_session_token');
                        localStorage.removeItem('gauth_session_data');
                        dispatch({ type: 'AUTH_FAILURE', payload: 'Session validation failed' });

                        // Clear Redux store
                        reduxDispatch(clearAuth());
                    }
                } catch (error) {
                    console.error('💥 [AuthProvider] Failed to validate session:', error);
                    localStorage.removeItem('gauth_session_token');
                    localStorage.removeItem('gauth_session_data');
                    dispatch({ type: 'AUTH_FAILURE', payload: 'Session validation failed' });

                    // Clear Redux store
                    reduxDispatch(clearAuth());
                }
            } else {
                dispatch({ type: 'AUTH_FAILURE', payload: '' });

                // Clear Redux store
                reduxDispatch(clearAuth());
            }
        };

        initializeSession();
    }, [reduxDispatch]);

    const loginWithGoogle = async (credentials: LoginCredentials) => {
        try {
            dispatch({ type: 'AUTH_START' });

            const response = await gauthApi.initiateGoogleOAuth(credentials);

            if (response.success && response.data.auth_url) {
                // Store state for callback verification
                localStorage.setItem('gauth_oauth_state', response.data.state);
                localStorage.setItem('gauth_organization_id', credentials.organization_id);

                // Redirect to Google OAuth
                window.location.href = response.data.auth_url;
            } else {
                throw new Error('Failed to get OAuth URL');
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Login failed';
            dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
        }
    };

    const handleOAuthCallback = async (code: string, state: string) => {
        try {
            dispatch({ type: 'AUTH_START' });

            const response = await gauthApi.handleGoogleOAuthCallback(code, state);

            if (response.success && response.data.session_token) {
                // Store session data
                localStorage.setItem('gauth_session_token', response.data.session_token);
                localStorage.setItem('gauth_session_data', JSON.stringify(response.data));

                dispatch({ type: 'AUTH_SUCCESS', payload: response.data });

                // Clear OAuth state
                localStorage.removeItem('gauth_oauth_state');
                localStorage.removeItem('gauth_organization_id');
            } else {
                throw new Error('Failed to complete OAuth callback');
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'OAuth callback failed';
            dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
        }
    };

    const logout = useCallback(async () => {
        try {
            // Call logout API to destroy session on server
            await gauthApi.logoutSession();
        } catch (error) {
            console.error('Failed to logout from server:', error);
            // Continue with local logout even if server logout fails
        }

        // Clear session data
        localStorage.removeItem('gauth_session_token');
        localStorage.removeItem('gauth_session_data');
        localStorage.removeItem('gauth_oauth_state');
        localStorage.removeItem('gauth_organization_id');

        dispatch({ type: 'AUTH_LOGOUT' });
    }, []);

    const clearError = () => {
        dispatch({ type: 'CLEAR_ERROR' });
    };

    const setSession = useCallback((session: AuthSession) => {
        // Store session data in localStorage
        localStorage.setItem('gauth_session_token', session.session_token);
        localStorage.setItem('gauth_session_data', JSON.stringify(session));

        // Update auth state
        dispatch({ type: 'AUTH_SUCCESS', payload: session });
    }, []);

    const refreshSession = useCallback(async () => {
        try {
            const response = await gauthApi.refreshSession();
            if (response.success) {
                dispatch({ type: 'SESSION_REFRESH', payload: response.data.expires_at });

                // Update localStorage with new expiration
                const sessionData = localStorage.getItem('gauth_session_data');
                if (sessionData) {
                    const session: AuthSession = JSON.parse(sessionData);
                    session.expires_at = response.data.expires_at;
                    localStorage.setItem('gauth_session_data', JSON.stringify(session));
                }
            }
        } catch (error) {
            console.error('Failed to refresh session:', error);
            // If refresh fails, logout the user
            logout();
        }
    }, [logout]);

    const validateSession = useCallback(async (): Promise<boolean> => {
        try {
            const response = await gauthApi.validateSession();
            if (response.success) {
                dispatch({ type: 'SESSION_VALIDATE', payload: response.data });
                return true;
            }
            return false;
        } catch (error) {
            console.error('Session validation failed:', error);
            return false;
        }
    }, []);

    const getSessionInfo = useCallback(async (): Promise<SessionInfo | null> => {
        try {
            const response = await gauthApi.getSessionInfo();
            if (response.success) {
                return response.data;
            }
            return null;
        } catch (error) {
            console.error('Failed to get session info:', error);
            return null;
        }
    }, []);

    const listSessions = useCallback(async (): Promise<SessionInfo[]> => {
        try {
            const response = await gauthApi.listSessions();
            if (response.success) {
                return response.data.sessions;
            }
            return [];
        } catch (error) {
            console.error('Failed to list sessions:', error);
            return [];
        }
    }, []);

    const destroySession = useCallback(async (sessionId: string) => {
        try {
            await gauthApi.destroySession(sessionId);
        } catch (error) {
            console.error('Failed to destroy session:', error);
            throw error;
        }
    }, []);

    const value: AuthContextType = {
        ...state,
        loginWithGoogle,
        handleOAuthCallback,
        setSession,
        logout,
        clearError,
        refreshSession,
        validateSession,
        getSessionInfo,
        listSessions,
        destroySession,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}

// Hook to use auth context
export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
