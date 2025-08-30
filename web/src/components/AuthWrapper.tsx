'use client';

import React from 'react';
import { useAuth } from '@/lib/auth-context';
import { SessionLoading } from '@/components/SessionLoading';

interface AuthWrapperProps {
    children: React.ReactNode;
}

export function AuthWrapper({ children }: AuthWrapperProps) {
    const { loading, isAuthenticated } = useAuth();

    console.log('🔄 [AuthWrapper] Render:', {
        loading,
        isAuthenticated
    });

    // Show loading screen while validating session
    if (loading) {
        console.log('⏳ [AuthWrapper] Showing loading screen');
        return <SessionLoading />;
    }

    console.log('✅ [AuthWrapper] Rendering children');
    return <>{children}</>;
}
