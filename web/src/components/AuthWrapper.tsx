'use client';

import React from 'react';
import { useAuth } from '@/lib/auth-context';
import { SessionLoading } from '@/components/SessionLoading';

interface AuthWrapperProps {
    children: React.ReactNode;
}

export function AuthWrapper({ children }: AuthWrapperProps) {
    const { loading, isAuthenticated } = useAuth();


    // Show loading screen while validating session
    if (loading) {
        return <SessionLoading />;
    }

    return <>{children}</>;
}
