'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';

interface ProtectedRouteProps {
    children: React.ReactNode;
    fallback?: React.ReactNode;
}

export function ProtectedRoute({ children, fallback }: ProtectedRouteProps) {
    const { isAuthenticated, loading } = useAuth();
    const router = useRouter();

    // Check if we're in test mode with mock authentication enabled
    const isTestMode = process.env.NODE_ENV === 'test' || process.env.NEXT_PUBLIC_TEST_MODE === 'true';
    const mockAuthEnabled = process.env.NEXT_PUBLIC_MOCK_AUTH === 'true' || 
        (typeof window !== 'undefined' && (window as any).__MOCK_AUTH__);

    useEffect(() => {
        // In test mode with mock auth, don't redirect
        // Otherwise, redirect unauthenticated users to sign-in
        if (!loading && !isAuthenticated && !(isTestMode && mockAuthEnabled)) {
            router.push('/auth/signin');
        }
    }, [isAuthenticated, loading, router, isTestMode, mockAuthEnabled]);

    if (loading) {
        return fallback || (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-muted-foreground">Loading...</p>
                </div>
            </div>
        );
    }

    // In test mode with mock auth, allow access even without authentication
    if (!isAuthenticated && !(isTestMode && mockAuthEnabled)) {
        return null; // Will redirect
    }

    return <>{children}</>;
}
