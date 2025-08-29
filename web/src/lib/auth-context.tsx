'use client';

import React, { createContext, useContext, useReducer, useEffect, ReactNode, useCallback } from 'react';
import { AuthState, AuthSession, LoginCredentials } from '@/types/auth';
import { gauthApi } from '@/services/gauthApi';

// Auth Action Types
type AuthAction =
    | { type: 'AUTH_START' }
    | { type: 'AUTH_SUCCESS'; payload: AuthSession }
    | { type: 'AUTH_FAILURE'; payload: string }
    | { type: 'AUTH_LOGOUT' }
    | { type: 'CLEAR_ERROR' };

// Initial state
const initialState: AuthState = {
    isAuthenticated: false,
    user: null,
    session: null,
    loading: false,
    error: null,
};

// Auth reducer
function authReducer(state: AuthState, action: AuthAction): AuthState {
    switch (action.type) {
        case 'AUTH_START':
            return {
                ...state,
                loading: true,
                error: null,
            };
        case 'AUTH_SUCCESS':
            return {
                ...state,
                isAuthenticated: true,
                user: action.payload.user,
                session: action.payload,
                loading: false,
                error: null,
            };
        case 'AUTH_FAILURE':
            return {
                ...state,
                isAuthenticated: false,
                user: null,
                session: null,
                loading: false,
                error: action.payload,
            };
        case 'AUTH_LOGOUT':
            return {
                ...state,
                isAuthenticated: false,
                user: null,
                session: null,
                loading: false,
                error: null,
            };
        case 'CLEAR_ERROR':
            return {
                ...state,
                error: null,
            };
        default:
            return state;
    }
}

// Auth context
interface AuthContextType extends AuthState {
    loginWithGoogle: (credentials: LoginCredentials) => Promise<void>;
    handleOAuthCallback: (code: string, state: string) => Promise<void>;
    setSession: (session: AuthSession) => void;
    logout: () => void;
    clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
interface AuthProviderProps {
    children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
    const [state, dispatch] = useReducer(authReducer, initialState);

    // Check for existing session on mount
    useEffect(() => {
        const sessionToken = localStorage.getItem('gauth_session_token');
        const sessionData = localStorage.getItem('gauth_session_data');

        if (sessionToken && sessionData) {
            try {
                const session: AuthSession = JSON.parse(sessionData);
                const expiresAt = new Date(session.expires_at);

                if (expiresAt > new Date()) {
                    dispatch({ type: 'AUTH_SUCCESS', payload: session });
                } else {
                    // Session expired, clear storage
                    localStorage.removeItem('gauth_session_token');
                    localStorage.removeItem('gauth_session_data');
                }
            } catch (error) {
                console.error('Failed to parse session data:', error);
                localStorage.removeItem('gauth_session_token');
                localStorage.removeItem('gauth_session_data');
            }
        }
    }, []);

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

    const logout = () => {
        // Clear session data
        localStorage.removeItem('gauth_session_token');
        localStorage.removeItem('gauth_session_data');
        localStorage.removeItem('gauth_oauth_state');
        localStorage.removeItem('gauth_organization_id');

        dispatch({ type: 'AUTH_LOGOUT' });
    };

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

    const value: AuthContextType = {
        ...state,
        loginWithGoogle,
        handleOAuthCallback,
        setSession,
        logout,
        clearError,
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
