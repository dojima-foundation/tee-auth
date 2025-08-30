'use client';

import React from 'react';
import { SessionManager } from '@/components/SessionManager';
import { SessionStatus } from '@/components/SessionStatus';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export default function SessionsPage() {
    return (
        <div className="container mx-auto py-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Session Management</h1>
                    <p className="text-muted-foreground mt-2">
                        Manage your active sessions and monitor session status
                    </p>
                </div>

                {/* Session Status Summary */}
                <div className="flex items-center gap-4">
                    <Card className="w-64">
                        <CardHeader className="pb-2">
                            <CardTitle className="text-sm font-medium">Current Session</CardTitle>
                            <CardDescription className="text-xs">
                                Real-time session information
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <SessionStatus showDetails={true} />
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Session Manager Component */}
            <SessionManager />
        </div>
    );
}
