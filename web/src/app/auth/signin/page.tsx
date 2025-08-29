'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { GoogleOAuthLogin } from '@/components/GoogleOAuthLogin';

export default function SignInPage() {
    const { isAuthenticated, loading } = useAuth();
    const router = useRouter();

    // Redirect if already authenticated
    useEffect(() => {
        if (isAuthenticated && !loading) {
            router.push('/dashboard');
        }
    }, [isAuthenticated, loading, router]);

    if (loading) {
        return (
            <div className="min-h-screen bg-background flex items-center justify-center">
                <div className="text-center">
                    <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-muted-foreground">Loading...</p>
                </div>
            </div>
        );
    }

    if (isAuthenticated) {
        return null; // Will redirect
    }

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                <div className="text-center mb-8">
                    <h1 className="text-3xl font-bold text-foreground mb-2">
                        Welcome Back
                    </h1>
                    <p className="text-muted-foreground">
                        Sign in with Google to continue
                    </p>
                </div>

                <div className="bg-card border border-border rounded-lg p-6 shadow-sm">
                    <GoogleOAuthLogin
                        onSuccess={() => {
                            router.push('/dashboard');
                        }}
                        onError={(error) => {
                            console.error('Login error:', error);
                        }}
                    />
                </div>

                <div className="mt-6 text-center">
                    <p className="text-sm text-muted-foreground">
                        New users will have an organization created automatically
                    </p>
                </div>
            </div>
        </div>
    );
}
