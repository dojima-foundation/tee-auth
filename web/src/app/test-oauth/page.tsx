'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function TestOAuthPage() {
    const router = useRouter();
    const [code, setCode] = useState('test-code-123');
    const [state, setState] = useState('test-state-456');

    const simulateOAuthCallback = () => {
        const callbackUrl = `/auth/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`;
        router.push(callbackUrl);
    };

    const simulateErrorCallback = () => {
        const errorUrl = `/auth/callback?error=${encodeURIComponent('access_denied')}`;
        router.push(errorUrl);
    };

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
            <div className="w-full max-w-md">
                <div className="text-center mb-8">
                    <h1 className="text-3xl font-bold text-foreground mb-2">
                        OAuth Test Page
                    </h1>
                    <p className="text-muted-foreground">
                        Test the OAuth callback flow
                    </p>
                </div>

                <div className="bg-card border border-border rounded-lg p-6 shadow-sm space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                            OAuth Code
                        </label>
                        <input
                            type="text"
                            value={code}
                            onChange={(e) => setCode(e.target.value)}
                            className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md"
                            placeholder="Enter OAuth code"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                            OAuth State
                        </label>
                        <input
                            type="text"
                            value={state}
                            onChange={(e) => setState(e.target.value)}
                            className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md"
                            placeholder="Enter OAuth state"
                        />
                    </div>

                    <div className="space-y-2">
                        <button
                            onClick={simulateOAuthCallback}
                            className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-2 px-4 rounded-lg transition-colors duration-200"
                        >
                            Simulate Successful OAuth Callback
                        </button>

                        <button
                            onClick={simulateErrorCallback}
                            className="w-full bg-destructive hover:bg-destructive/90 text-destructive-foreground font-semibold py-2 px-4 rounded-lg transition-colors duration-200"
                        >
                            Simulate OAuth Error
                        </button>
                    </div>
                </div>

                <div className="mt-6 text-center">
                    <p className="text-sm text-muted-foreground">
                        This page is for testing the OAuth callback flow
                    </p>
                </div>
            </div>
        </div>
    );
}
