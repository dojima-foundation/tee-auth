'use client';

import React from 'react';
import { Loader2 } from 'lucide-react';

export function SessionLoading() {
    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-background">
            <div className="flex flex-col items-center space-y-4">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <div className="text-center">
                    <h2 className="text-lg font-semibold text-foreground">Validating Session</h2>
                    <p className="text-sm text-muted-foreground">Please wait while we verify your session...</p>
                </div>
            </div>
        </div>
    );
}
