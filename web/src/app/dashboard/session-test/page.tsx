'use client';

import React, { useState, useEffect } from 'react';
import { useSessionManagement } from '@/lib/session-middleware';
import { SessionInfo } from '@/types/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useSnackbar } from '@/components/ui/snackbar';

export default function SessionTestPage() {
    const {
        isAuthenticated,
        getSessionInfo,
        listSessions,
        refreshSession,
        validateSession,
        destroySession
    } = useSessionManagement();

    const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
    const [allSessions, setAllSessions] = useState<SessionInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const { showSnackbar } = useSnackbar();

    const loadSessionInfo = async () => {
        if (!isAuthenticated) return;

        try {
            setLoading(true);
            const info = await getSessionInfo();
            setSessionInfo(info);
            showSnackbar({ title: 'Session info loaded', type: 'success' });
        } catch (error) {
            console.error('Failed to load session info:', error);
            showSnackbar({ title: 'Failed to load session info', type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    const loadAllSessions = async () => {
        if (!isAuthenticated) return;

        try {
            setLoading(true);
            const sessions = await listSessions();
            setAllSessions(sessions);
            showSnackbar({ title: `Loaded ${sessions.length} sessions`, type: 'success' });
        } catch (error) {
            console.error('Failed to load sessions:', error);
            showSnackbar({ title: 'Failed to load sessions', type: 'error' });
        } finally {
            setLoading(false);
        }
    };

    const handleRefreshSession = async () => {
        try {
            await refreshSession();
            await loadSessionInfo();
            showSnackbar({ title: 'Session refreshed successfully', type: 'success' });
        } catch (error) {
            console.error('Failed to refresh session:', error);
            showSnackbar({ title: 'Failed to refresh session', type: 'error' });
        }
    };

    const handleValidateSession = async () => {
        try {
            const isValid = await validateSession();
            showSnackbar({ title: `Session validation: ${isValid ? 'Valid' : 'Invalid'}`, type: isValid ? 'success' : 'error' });
        } catch (error) {
            console.error('Failed to validate session:', error);
            showSnackbar({ title: 'Failed to validate session', type: 'error' });
        }
    };

    const handleDestroySession = async (sessionId: string) => {
        try {
            await destroySession(sessionId);
            await loadAllSessions();
            showSnackbar({ title: 'Session destroyed successfully', type: 'success' });
        } catch (error) {
            console.error('Failed to destroy session:', error);
            showSnackbar({ title: 'Failed to destroy session', type: 'error' });
        }
    };

    useEffect(() => {
        if (isAuthenticated) {
            loadSessionInfo();
            loadAllSessions();
        }
    }, [isAuthenticated]);

    if (!isAuthenticated) {
        return (
            <div className="container mx-auto py-6">
                <Card>
                    <CardHeader>
                        <CardTitle>Session Test</CardTitle>
                        <CardDescription>Please log in to test session management</CardDescription>
                    </CardHeader>
                </Card>
            </div>
        );
    }

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleString();
    };

    const isExpired = (expiresAt: string) => {
        return new Date(expiresAt) < new Date();
    };

    return (
        <div className="container mx-auto py-6 space-y-6">
            <div>
                <h1 className="text-3xl font-bold text-foreground">Session Management Test</h1>
                <p className="text-muted-foreground mt-2">
                    Test the session management functionality
                </p>
            </div>

            {/* Test Actions */}
            <Card>
                <CardHeader>
                    <CardTitle>Test Actions</CardTitle>
                    <CardDescription>Test various session management operations</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-wrap gap-2">
                        <Button
                            onClick={loadSessionInfo}
                            disabled={loading}
                            variant="outline"
                        >
                            Load Session Info
                        </Button>
                        <Button
                            onClick={loadAllSessions}
                            disabled={loading}
                            variant="outline"
                        >
                            Load All Sessions
                        </Button>
                        <Button
                            onClick={handleRefreshSession}
                            disabled={loading}
                            variant="outline"
                        >
                            Refresh Session
                        </Button>
                        <Button
                            onClick={handleValidateSession}
                            disabled={loading}
                            variant="outline"
                        >
                            Validate Session
                        </Button>
                    </div>
                </CardContent>
            </Card>

            {/* Current Session Info */}
            {sessionInfo && (
                <Card>
                    <CardHeader>
                        <CardTitle>Current Session Info</CardTitle>
                        <CardDescription>Information about your current session</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <span className="font-medium">Session ID:</span>
                                <p className="text-muted-foreground font-mono text-xs">
                                    {sessionInfo.session_id}
                                </p>
                            </div>
                            <div>
                                <span className="font-medium">User ID:</span>
                                <p className="text-muted-foreground font-mono text-xs">
                                    {sessionInfo.user_id}
                                </p>
                            </div>
                            <div>
                                <span className="font-medium">Email:</span>
                                <p className="text-muted-foreground">{sessionInfo.email}</p>
                            </div>
                            <div>
                                <span className="font-medium">Role:</span>
                                <p className="text-muted-foreground">{sessionInfo.role}</p>
                            </div>
                            <div>
                                <span className="font-medium">Created:</span>
                                <p className="text-muted-foreground">
                                    {formatDate(sessionInfo.created_at)}
                                </p>
                            </div>
                            <div>
                                <span className="font-medium">Last Activity:</span>
                                <p className="text-muted-foreground">
                                    {formatDate(sessionInfo.last_activity)}
                                </p>
                            </div>
                            <div>
                                <span className="font-medium">Expires:</span>
                                <p className="text-muted-foreground">
                                    {formatDate(sessionInfo.expires_at)}
                                </p>
                            </div>
                            <div>
                                <span className="font-medium">Status:</span>
                                <div className="mt-1">
                                    <Badge
                                        variant={isExpired(sessionInfo.expires_at) ? 'destructive' : 'default'}
                                    >
                                        {isExpired(sessionInfo.expires_at) ? 'Expired' : 'Active'}
                                    </Badge>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* All Sessions */}
            {allSessions.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle>All Sessions ({allSessions.length})</CardTitle>
                        <CardDescription>All your active sessions</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3">
                            {allSessions.map((session) => (
                                <div
                                    key={session.session_id}
                                    className="flex items-center justify-between p-3 border rounded-lg"
                                >
                                    <div className="flex-1">
                                        <div className="flex items-center gap-2 mb-1">
                                            <span className="font-medium text-sm">
                                                {session.email}
                                            </span>
                                            <Badge
                                                variant={isExpired(session.expires_at) ? 'destructive' : 'default'}
                                                className="text-xs"
                                            >
                                                {isExpired(session.expires_at) ? 'Expired' : 'Active'}
                                            </Badge>
                                            {session.oauth_provider && (
                                                <Badge variant="outline" className="text-xs">
                                                    {session.oauth_provider}
                                                </Badge>
                                            )}
                                        </div>
                                        <div className="text-xs text-muted-foreground space-y-1">
                                            <p>Session: {session.session_id.substring(0, 12)}...</p>
                                            <p>Last Activity: {formatDate(session.last_activity)}</p>
                                            <p>Expires: {formatDate(session.expires_at)}</p>
                                        </div>
                                    </div>

                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={() => handleDestroySession(session.session_id)}
                                        disabled={isExpired(session.expires_at)}
                                    >
                                        Destroy
                                    </Button>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
