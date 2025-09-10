'use client';

import React, { useEffect, useState, useRef, Suspense, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { useDispatch } from 'react-redux';
import { setAuthSession } from '@/store/authSlice';
import type { AppDispatch } from '@/store';
import { gauthApi } from '@/services/gauthApi';

function OAuthCallbackContent() {
    const { handleOAuthCallback, setSession, loading, error } = useAuth();
    const router = useRouter();
    const searchParams = useSearchParams();
    const dispatch = useDispatch<AppDispatch>();
    const [processing, setProcessing] = useState(false);
    const hasProcessed = useRef(false);

    const processCallback = useCallback(async () => {
        // Prevent multiple executions
        if (hasProcessed.current) {
            return;
        }

        // Check for session parameters (successful OAuth flow)
        const sessionToken = searchParams.get('session_token') || searchParams.get('session_id');
        const expiresAt = searchParams.get('expires_at');
        const userId = searchParams.get('user_id');
        const email = searchParams.get('email');
        const organizationId = searchParams.get('organization_id');

        // Check for OAuth parameters (if backend redirects with code/state)
        const code = searchParams.get('code') || searchParams.get('authorization_code');
        const state = searchParams.get('state') || searchParams.get('oauth_state');
        const error = searchParams.get('error') || searchParams.get('oauth_error');


        if (error) {
            console.error('OAuth error:', error);
            router.push('/auth/error?error=' + encodeURIComponent(error));
            return;
        }

        // If we have session data, handle successful login
        if (sessionToken && userId && email && organizationId && organizationId.trim() !== '') {
            try {
                setProcessing(true);
                hasProcessed.current = true;

                // If expires_at is not provided, set a default expiration (24 hours from now)
                let sessionExpiresAt = expiresAt;
                if (!sessionExpiresAt) {
                    const defaultExpiry = Date.now() + (24 * 60 * 60 * 1000); // 24 hours
                    sessionExpiresAt = defaultExpiry.toString();
                }

                // Ensure expires_at is a valid timestamp string
                let expiresAtTimestamp: string;
                if (typeof sessionExpiresAt === 'string' && /^\d+$/.test(sessionExpiresAt)) {
                    // Already a timestamp string
                    expiresAtTimestamp = sessionExpiresAt;
                } else {
                    // Try to parse as date and convert to timestamp
                    const date = new Date(sessionExpiresAt);
                    if (isNaN(date.getTime())) {
                        // If parsing fails, use default
                        expiresAtTimestamp = (Date.now() + (24 * 60 * 60 * 1000)).toString();
                    } else {
                        expiresAtTimestamp = date.getTime().toString();
                    }
                }

                // Store session data in the format expected by auth context
                const sessionData = {
                    session_token: sessionToken,
                    expires_at: expiresAtTimestamp,
                    user: {
                        id: userId,
                        email: email,
                        username: email.split('@')[0], // Use email prefix as username
                        organization_id: organizationId,
                        is_active: true,
                        created_at: Date.now().toString(),
                        updated_at: Date.now().toString()
                    },
                    auth_method: {
                        id: 'google-oauth',
                        type: 'google_oauth',
                        name: 'Google OAuth'
                    }
                };

                // Update auth context state using setSession
                setSession(sessionData);

                // Store in Redux store for global access
                dispatch(setAuthSession(sessionData));

                // Try to fetch actual session info from backend to get correct expiration
                try {
                    const sessionInfo = await gauthApi.getSessionInfo();
                    if (sessionInfo.success) {
                        // Update session with actual backend data
                        const updatedSessionData = {
                            ...sessionData,
                            expires_at: new Date(sessionInfo.data.expires_at).getTime().toString(),
                        };
                        setSession(updatedSessionData);
                        dispatch(setAuthSession(updatedSessionData));
                    }
                } catch (error) {
                    console.warn('Failed to fetch session info from backend, using default expiration:', error);
                    // Continue with default expiration
                }

                // Redirect to dashboard
                router.push('/dashboard');
                return;
            } catch (err) {
                console.error('Error processing session data:', err);
                router.push('/auth/error?error=' + encodeURIComponent('Failed to process session'));
                return;
            } finally {
                setProcessing(false);
            }
        }

        // If we have code and state, handle OAuth callback
        if (code && state) {
            try {
                setProcessing(true);
                hasProcessed.current = true;

                await handleOAuthCallback(code, state);
                return;
            } catch (err) {
                console.error('Error processing OAuth callback:', err);
                router.push('/auth/error?error=' + encodeURIComponent('Failed to process OAuth callback'));
                return;
            } finally {
                setProcessing(false);
            }
        }

        // Check if we have session data but missing organization ID
        if (sessionToken && userId && email && (!organizationId || organizationId.trim() === '')) {

            // If expires_at is not provided, set a default expiration (24 hours from now)
            let sessionExpiresAt = expiresAt;
            if (!sessionExpiresAt) {
                const defaultExpiry = Date.now() + (24 * 60 * 60 * 1000); // 24 hours
                sessionExpiresAt = defaultExpiry.toString();
            }

            // Ensure expires_at is a valid timestamp string
            let expiresAtTimestamp: string;
            if (typeof sessionExpiresAt === 'string' && /^\d+$/.test(sessionExpiresAt)) {
                // Already a timestamp string
                expiresAtTimestamp = sessionExpiresAt;
            } else {
                // Try to parse as date and convert to timestamp
                const date = new Date(sessionExpiresAt);
                if (isNaN(date.getTime())) {
                    // If parsing fails, use default
                    expiresAtTimestamp = (Date.now() + (24 * 60 * 60 * 1000)).toString();
                } else {
                    expiresAtTimestamp = date.getTime().toString();
                }
            }

            // Proceed with empty organization ID
            const sessionData = {
                session_token: sessionToken,
                expires_at: expiresAtTimestamp,
                user: {
                    id: userId,
                    email: email,
                    username: email.split('@')[0],
                    organization_id: '', // Empty organization ID
                    is_active: true,
                    created_at: Date.now().toString(),
                    updated_at: Date.now().toString()
                },
                auth_method: {
                    id: 'google-oauth',
                    type: 'google_oauth',
                    name: 'Google OAuth'
                }
            };

            setSession(sessionData);
            dispatch(setAuthSession(sessionData));

            // Try to fetch actual session info from backend to get correct expiration
            try {
                const sessionInfo = await gauthApi.getSessionInfo();
                if (sessionInfo.success) {
                    // Update session with actual backend data
                    const updatedSessionData = {
                        ...sessionData,
                        expires_at: new Date(sessionInfo.data.expires_at).getTime().toString(),
                    };
                    setSession(updatedSessionData);
                    dispatch(setAuthSession(updatedSessionData));
                }
            } catch (error) {
                console.warn('Failed to fetch session info from backend, using default expiration:', error);
                // Continue with default expiration
            }

            router.push('/dashboard');
            return;
        }

        // No valid parameters found
        console.error('Missing OAuth parameters');
        router.push('/auth/error?error=' + encodeURIComponent('Missing OAuth parameters'));
        return;
    }, [searchParams, handleOAuthCallback, router, setSession, dispatch]);

    useEffect(() => {
        processCallback();
    }, [processCallback]);

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
            <div className="text-center">
                {processing || loading ? (
                    <>
                        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                        <h2 className="text-xl font-semibold text-foreground mb-2">
                            Completing Sign In
                        </h2>
                        <p className="text-muted-foreground">
                            Please wait while we complete your authentication...
                        </p>
                    </>
                ) : error ? (
                    <>
                        <div className="w-12 h-12 bg-destructive/10 rounded-full flex items-center justify-center mx-auto mb-4">
                            <svg className="w-6 h-6 text-destructive" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </div>
                        <h2 className="text-xl font-semibold text-foreground mb-2">
                            Authentication Failed
                        </h2>
                        <p className="text-muted-foreground mb-4">
                            {error}
                        </p>
                        <button
                            onClick={() => router.push('/auth/signin')}
                            className="bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-2 px-4 rounded-lg transition-colors duration-200"
                        >
                            Try Again
                        </button>
                    </>
                ) : (
                    <>
                        <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                            <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                        </div>
                        <h2 className="text-xl font-semibold text-foreground mb-2">
                            Sign In Successful
                        </h2>
                        <p className="text-muted-foreground">
                            Redirecting to dashboard...
                        </p>
                    </>
                )}
            </div>
        </div>
    );
}

export default function OAuthCallbackPage() {
    return (
        <Suspense fallback={
            <div className="min-h-screen bg-background flex items-center justify-center p-4">
                <div className="text-center">
                    <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <h2 className="text-xl font-semibold text-foreground mb-2">
                        Loading...
                    </h2>
                    <p className="text-muted-foreground">
                        Please wait...
                    </p>
                </div>
            </div>
        }>
            <OAuthCallbackContent />
        </Suspense>
    );
}
