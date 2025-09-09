'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useSessionManagement } from '@/lib/session-middleware';
import { SessionInfo } from '@/types/auth';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useSnackbar } from '@/components/ui/snackbar';

interface SessionStatusProps {
    showDetails?: boolean;
    className?: string;
}

export function SessionStatus({ showDetails = false, className = '' }: SessionStatusProps) {
    const {
        isAuthenticated,
        getSessionInfo,
        refreshSession
    } = useSessionManagement();

    const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
    // const [loading, setLoading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const { showSnackbar } = useSnackbar();

    // Load session info
    const loadSessionInfo = useCallback(async () => {
        if (!isAuthenticated) return;

        try {
            const info = await getSessionInfo();
            setSessionInfo(info);
        } catch (error) {
            console.error('Failed to load session info:', error);
        }
    }, [isAuthenticated, getSessionInfo]);

    // Refresh session
    const handleRefresh = async () => {
        try {
            setRefreshing(true);
            await refreshSession();
            await loadSessionInfo();
            showSnackbar({ title: 'Session refreshed', type: 'success' });
        } catch (error) {
            console.error('Failed to refresh session:', error);
            showSnackbar({ title: 'Failed to refresh session', type: 'error' });
        } finally {
            setRefreshing(false);
        }
    };

    // Load session info on mount and when authentication status changes
    useEffect(() => {
        if (isAuthenticated) {
            loadSessionInfo();
        } else {
            setSessionInfo(null);
        }
    }, [isAuthenticated, loadSessionInfo]);

    // Auto-refresh session info every minute
    useEffect(() => {
        if (!isAuthenticated) return;

        const interval = setInterval(loadSessionInfo, 60000); // 1 minute
        return () => clearInterval(interval);
    }, [isAuthenticated, loadSessionInfo]);

    if (!isAuthenticated || !sessionInfo) {
        return null;
    }

    const isExpired = (expiresAt: string) => {
        return new Date(expiresAt) < new Date();
    };

    const getTimeUntilExpiry = (expiresAt: string) => {
        const now = new Date();
        const expiry = new Date(expiresAt);
        const diff = expiry.getTime() - now.getTime();

        if (diff <= 0) return 'Expired';

        const minutes = Math.floor(diff / (1000 * 60));
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) return `${days}d ${hours % 24}h`;
        if (hours > 0) return `${hours}h ${minutes % 60}m`;
        return `${minutes}m`;
    };

    const getStatusColor = () => {
        if (isExpired(sessionInfo.expires_at)) return 'destructive';

        const timeUntilExpiry = getTimeUntilExpiry(sessionInfo.expires_at);
        if (timeUntilExpiry.includes('m') && parseInt(timeUntilExpiry) < 30) {
            return 'destructive'; // Less than 30 minutes
        }
        if (timeUntilExpiry.includes('h') && parseInt(timeUntilExpiry) < 2) {
            return 'secondary'; // Less than 2 hours
        }
        return 'default';
    };

    const timeUntilExpiry = getTimeUntilExpiry(sessionInfo.expires_at);
    const statusColor = getStatusColor();

    if (!showDetails) {
        return (
            <div className={`flex items-center gap-2 ${className}`}>
                <Badge variant={statusColor} className="text-xs">
                    {isExpired(sessionInfo.expires_at) ? 'Expired' : `Expires in ${timeUntilExpiry}`}
                </Badge>
                {!isExpired(sessionInfo.expires_at) && (
                    <Button
                        onClick={handleRefresh}
                        disabled={refreshing}
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2 text-xs"
                    >
                        {refreshing ? (
                            <div className="w-3 h-3 border border-current border-t-transparent rounded-full animate-spin" />
                        ) : (
                            '↻'
                        )}
                    </Button>
                )}
            </div>
        );
    }

    return (
        <div className={`space-y-2 ${className}`}>
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <Badge variant={statusColor} className="text-xs">
                        {isExpired(sessionInfo.expires_at) ? 'Expired' : `Expires in ${timeUntilExpiry}`}
                    </Badge>
                    {sessionInfo.oauth_provider && (
                        <Badge variant="outline" className="text-xs">
                            {sessionInfo.oauth_provider}
                        </Badge>
                    )}
                </div>
                <Button
                    onClick={handleRefresh}
                    disabled={refreshing}
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2 text-xs"
                >
                    {refreshing ? (
                        <div className="w-3 h-3 border border-current border-t-transparent rounded-full animate-spin" />
                    ) : (
                        'Refresh'
                    )}
                </Button>
            </div>

            <div className="text-xs text-muted-foreground space-y-1">
                <p>Session: {sessionInfo.session_id.substring(0, 12)}...</p>
                <p>Last Activity: {new Date(sessionInfo.last_activity).toLocaleString()}</p>
                <p>Expires: {new Date(sessionInfo.expires_at).toLocaleString()}</p>
            </div>
        </div>
    );
}
