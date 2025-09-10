'use client';

import { useEffect, useRef, useCallback } from 'react';
import { useAuth } from './auth-context';

interface SessionMiddlewareProps {
    children: React.ReactNode;
    refreshInterval?: number; // in milliseconds, default 5 minutes
    validateOnMount?: boolean;
}

export function SessionMiddleware({
    children,
    refreshInterval = 5 * 60 * 1000, // 5 minutes
    validateOnMount = false // AuthProvider now handles initial validation
}: SessionMiddlewareProps) {
    const { isAuthenticated, validateSession, refreshSession, logout } = useAuth();
    const refreshTimerRef = useRef<NodeJS.Timeout | null>(null);
    const isRefreshingRef = useRef(false);

    // Function to refresh session
    const handleRefresh = useCallback(async () => {
        if (isRefreshingRef.current || !isAuthenticated) {
            return;
        }

        try {
            isRefreshingRef.current = true;
            await refreshSession();
        } catch (error) {
            console.error('Session refresh failed:', error);
            // If refresh fails, logout the user
            await logout();
        } finally {
            isRefreshingRef.current = false;
        }
    }, [isAuthenticated, refreshSession, logout]);

    // Function to validate session
    const handleValidation = useCallback(async () => {
        if (!isAuthenticated) {
            return;
        }

        try {
            const isValid = await validateSession();
            if (!isValid) {
                console.warn('Session validation failed, logging out');
                await logout();
            }
        } catch (error) {
            console.error('Session validation error:', error);
            await logout();
        }
    }, [isAuthenticated, validateSession, logout]);

    // Set up automatic session refresh
    useEffect(() => {
        if (!isAuthenticated) {
            if (refreshTimerRef.current) {
                clearInterval(refreshTimerRef.current);
                refreshTimerRef.current = null;
            }
            return;
        }

        // Set up refresh timer
        refreshTimerRef.current = setInterval(handleRefresh, refreshInterval);

        return () => {
            if (refreshTimerRef.current) {
                clearInterval(refreshTimerRef.current);
                refreshTimerRef.current = null;
            }
        };
    }, [isAuthenticated, handleRefresh, refreshInterval]);

    // Validate session on mount if enabled
    useEffect(() => {


        if (validateOnMount && isAuthenticated) {
            handleValidation();
        }
    }, [validateOnMount, isAuthenticated, handleValidation]);

    // Set up visibility change handler to refresh session when user returns
    useEffect(() => {
        const handleVisibilityChange = () => {
            if (document.visibilityState === 'visible' && isAuthenticated) {
                handleValidation();
            }
        };

        document.addEventListener('visibilitychange', handleVisibilityChange);
        return () => {
            document.removeEventListener('visibilitychange', handleVisibilityChange);
        };
    }, [isAuthenticated, handleValidation]);

    // Set up beforeunload handler to refresh session before page unload
    useEffect(() => {
        const handleBeforeUnload = () => {
            if (isAuthenticated && !isRefreshingRef.current) {
                // Use sendBeacon for reliable request during page unload
                const sessionToken = localStorage.getItem('gauth_session_token');
                if (sessionToken) {
                    navigator.sendBeacon(
                        `${process.env.NEXT_PUBLIC_GAUTH_API_URL || 'http://localhost:8082'}/api/v1/sessions/refresh`,
                        JSON.stringify({})
                    );
                }
            }
        };

        window.addEventListener('beforeunload', handleBeforeUnload);
        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, [isAuthenticated]);

    return <>{children}</>;
}

// Hook for manual session management
export function useSessionManagement() {
    const {
        isAuthenticated,
        validateSession,
        refreshSession,
        getSessionInfo,
        listSessions,
        destroySession
    } = useAuth();

    const refreshSessionManually = useCallback(async () => {
        if (!isAuthenticated) {
            throw new Error('No active session to refresh');
        }
        await refreshSession();
    }, [isAuthenticated, refreshSession]);

    const validateSessionManually = useCallback(async () => {
        if (!isAuthenticated) {
            return false;
        }
        return await validateSession();
    }, [isAuthenticated, validateSession]);

    const getCurrentSessionInfo = useCallback(async () => {
        if (!isAuthenticated) {
            return null;
        }
        return await getSessionInfo();
    }, [isAuthenticated, getSessionInfo]);

    const getAllSessions = useCallback(async () => {
        if (!isAuthenticated) {
            return [];
        }
        return await listSessions();
    }, [isAuthenticated, listSessions]);

    const destroySessionById = useCallback(async (sessionId: string) => {
        if (!isAuthenticated) {
            throw new Error('No active session');
        }
        await destroySession(sessionId);
    }, [isAuthenticated, destroySession]);

    return {
        isAuthenticated,
        refreshSession: refreshSessionManually,
        validateSession: validateSessionManually,
        getSessionInfo: getCurrentSessionInfo,
        listSessions: getAllSessions,
        destroySession: destroySessionById,
    };
}
