'use client';

import React, { useState, useEffect } from 'react';
import { useSessionManagement } from '@/lib/session-middleware';
import { SessionInfo } from '@/types/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { useSnackbar } from '@/components/ui/snackbar';

export function SessionManager() {
    const {
        isAuthenticated,
        getSessionInfo,
        listSessions,
        destroySession,
        refreshSession
    } = useSessionManagement();

    const [currentSession, setCurrentSession] = useState<SessionInfo | null>(null);
    const [allSessions, setAllSessions] = useState<SessionInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const { showSnackbar } = useSnackbar();

    // Load current session info
    const loadCurrentSession = async () => {
        if (!isAuthenticated) return;

        try {
            setLoading(true);
            const sessionInfo = await getSessionInfo();
            setCurrentSession(sessionInfo);
        } catch (error) {
            console.error('Failed to load current session:', error);
            showSnackbar('Failed to load session information', 'error');
        } finally {
            setLoading(false);
        }
    };

    // Load all sessions
    const loadAllSessions = async () => {
        if (!isAuthenticated) return;

        try {
            setLoading(true);
            const sessions = await listSessions();
            setAllSessions(sessions);
        } catch (error) {
            console.error('Failed to load sessions:', error);
            showSnackbar('Failed to load sessions', 'error');
        } finally {
            setLoading(false);
        }
    };

    // Refresh current session
    const handleRefreshSession = async () => {
        try {
            setRefreshing(true);
            await refreshSession();
            await loadCurrentSession();
            showSnackbar('Session refreshed successfully', 'success');
        } catch (error) {
            console.error('Failed to refresh session:', error);
            showSnackbar('Failed to refresh session', 'error');
        } finally {
            setRefreshing(false);
        }
    };

    // Destroy a specific session
    const handleDestroySession = async (sessionId: string) => {
        try {
            await destroySession(sessionId);
            await loadAllSessions();
            showSnackbar('Session destroyed successfully', 'success');
        } catch (error) {
            console.error('Failed to destroy session:', error);
            showSnackbar('Failed to destroy session', 'error');
        }
    };

    // Load data on mount
    useEffect(() => {
        if (isAuthenticated) {
            loadCurrentSession();
            loadAllSessions();
        }
    }, [isAuthenticated]);

    if (!isAuthenticated) {
        return null;
    }

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleString();
    };

    const isExpired = (expiresAt: string) => {
        return new Date(expiresAt) < new Date();
    };

    const getSessionStatus = (session: SessionInfo) => {
        if (isExpired(session.expires_at)) {
            return { label: 'Expired', variant: 'destructive' as const };
        }
        if (session.session_id === currentSession?.session_id) {
            return { label: 'Current', variant: 'default' as const };
        }
        return { label: 'Active', variant: 'secondary' as const };
    };

    return (
        <div className="space-y-6">
            {/* Current Session */}
            <Card>
                <CardHeader>
                    <CardTitle>Current Session</CardTitle>
                    <CardDescription>
                        Information about your current active session
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="flex items-center justify-center py-4">
                            <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                        </div>
                    ) : currentSession ? (
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                    <span className="font-medium">Session ID:</span>
                                    <p className="text-muted-foreground font-mono text-xs">
                                        {currentSession.session_id}
                                    </p>
                                </div>
                                <div>
                                    <span className="font-medium">User ID:</span>
                                    <p className="text-muted-foreground font-mono text-xs">
                                        {currentSession.user_id}
                                    </p>
                                </div>
                                <div>
                                    <span className="font-medium">Email:</span>
                                    <p className="text-muted-foreground">{currentSession.email}</p>
                                </div>
                                <div>
                                    <span className="font-medium">Role:</span>
                                    <p className="text-muted-foreground">{currentSession.role}</p>
                                </div>
                                <div>
                                    <span className="font-medium">Created:</span>
                                    <p className="text-muted-foreground">
                                        {formatDate(currentSession.created_at)}
                                    </p>
                                </div>
                                <div>
                                    <span className="font-medium">Last Activity:</span>
                                    <p className="text-muted-foreground">
                                        {formatDate(currentSession.last_activity)}
                                    </p>
                                </div>
                                <div>
                                    <span className="font-medium">Expires:</span>
                                    <p className="text-muted-foreground">
                                        {formatDate(currentSession.expires_at)}
                                    </p>
                                </div>
                                <div>
                                    <span className="font-medium">Status:</span>
                                    <div className="mt-1">
                                        <Badge
                                            variant={isExpired(currentSession.expires_at) ? 'destructive' : 'default'}
                                        >
                                            {isExpired(currentSession.expires_at) ? 'Expired' : 'Active'}
                                        </Badge>
                                    </div>
                                </div>
                            </div>

                            <div className="flex gap-2">
                                <Button
                                    onClick={handleRefreshSession}
                                    disabled={refreshing}
                                    size="sm"
                                >
                                    {refreshing ? (
                                        <>
                                            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                                            Refreshing...
                                        </>
                                    ) : (
                                        'Refresh Session'
                                    )}
                                </Button>
                                <Button
                                    onClick={loadCurrentSession}
                                    disabled={loading}
                                    variant="outline"
                                    size="sm"
                                >
                                    Reload Info
                                </Button>
                            </div>
                        </div>
                    ) : (
                        <p className="text-muted-foreground">No session information available</p>
                    )}
                </CardContent>
            </Card>

            {/* All Sessions */}
            <Card>
                <CardHeader>
                    <CardTitle>All Sessions</CardTitle>
                    <CardDescription>
                        Manage all your active sessions across devices
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="flex items-center justify-center py-4">
                            <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                        </div>
                    ) : allSessions.length > 0 ? (
                        <div className="space-y-3">
                            {allSessions.map((session) => {
                                const status = getSessionStatus(session);
                                const isCurrent = session.session_id === currentSession?.session_id;

                                return (
                                    <div
                                        key={session.session_id}
                                        className="flex items-center justify-between p-3 border rounded-lg"
                                    >
                                        <div className="flex-1">
                                            <div className="flex items-center gap-2 mb-1">
                                                <span className="font-medium text-sm">
                                                    {session.email}
                                                </span>
                                                <Badge variant={status.variant} className="text-xs">
                                                    {status.label}
                                                </Badge>
                                                {session.oauth_provider && (
                                                    <Badge variant="outline" className="text-xs">
                                                        {session.oauth_provider}
                                                    </Badge>
                                                )}
                                            </div>
                                            <div className="text-xs text-muted-foreground space-y-1">
                                                <p>Session: {session.session_id.substring(0, 8)}...</p>
                                                <p>Last Activity: {formatDate(session.last_activity)}</p>
                                                <p>Expires: {formatDate(session.expires_at)}</p>
                                            </div>
                                        </div>

                                        {!isCurrent && (
                                            <Dialog>
                                                <DialogTrigger asChild>
                                                    <Button
                                                        variant="destructive"
                                                        size="sm"
                                                        disabled={isExpired(session.expires_at)}
                                                    >
                                                        Destroy
                                                    </Button>
                                                </DialogTrigger>
                                                <DialogContent>
                                                    <DialogHeader>
                                                        <DialogTitle>Destroy Session</DialogTitle>
                                                        <DialogDescription>
                                                            Are you sure you want to destroy this session?
                                                            This will log out the user from this device.
                                                        </DialogDescription>
                                                    </DialogHeader>
                                                    <div className="flex justify-end gap-2 mt-4">
                                                        <Button variant="outline">Cancel</Button>
                                                        <Button
                                                            variant="destructive"
                                                            onClick={() => handleDestroySession(session.session_id)}
                                                        >
                                                            Destroy Session
                                                        </Button>
                                                    </div>
                                                </DialogContent>
                                            </Dialog>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    ) : (
                        <p className="text-muted-foreground">No sessions found</p>
                    )}

                    <div className="flex justify-end mt-4">
                        <Button
                            onClick={loadAllSessions}
                            disabled={loading}
                            variant="outline"
                            size="sm"
                        >
                            Refresh Sessions
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
