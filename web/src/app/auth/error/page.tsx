'use client';

import React, { Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';

function AuthErrorContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const error = searchParams.get('error') || 'An unknown error occurred';

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
            <div className="w-full max-w-md text-center">
                <div className="w-16 h-16 bg-destructive/10 rounded-full flex items-center justify-center mx-auto mb-6">
                    <svg className="w-8 h-8 text-destructive" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                </div>

                <h1 className="text-2xl font-bold text-foreground mb-4">
                    Authentication Error
                </h1>

                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 mb-6">
                    <p className="text-sm text-destructive">
                        {error}
                    </p>
                </div>

                <div className="space-y-3">
                    <button
                        onClick={() => router.push('/auth/signin')}
                        className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-3 px-4 rounded-lg transition-colors duration-200"
                    >
                        Try Again
                    </button>

                    <Link
                        href="/"
                        className="block w-full bg-secondary hover:bg-secondary/80 text-secondary-foreground font-semibold py-3 px-4 rounded-lg transition-colors duration-200"
                    >
                        Go Home
                    </Link>
                </div>

                <div className="mt-6 text-sm text-muted-foreground">
                    <p>
                        If this problem persists, please contact support.
                    </p>
                </div>
            </div>
        </div>
    );
}

export default function AuthErrorPage() {
    return (
        <Suspense fallback={
            <div className="min-h-screen bg-background flex items-center justify-center p-4">
                <div className="w-full max-w-md text-center">
                    <div className="w-16 h-16 bg-muted rounded-full animate-pulse mx-auto mb-6" />
                    <div className="h-8 bg-muted rounded animate-pulse mb-4" />
                    <div className="h-4 bg-muted rounded animate-pulse mb-6" />
                </div>
            </div>
        }>
            <AuthErrorContent />
        </Suspense>
    );
}
