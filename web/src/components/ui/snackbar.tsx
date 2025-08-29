'use client';

import React, { createContext, useContext, useState, useCallback } from 'react';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

export type SnackbarType = 'success' | 'error' | 'info' | 'warning';

export interface SnackbarMessage {
    id: string;
    type: SnackbarType;
    title: string;
    message?: string;
    duration?: number;
}

interface SnackbarContextType {
    showSnackbar: (message: Omit<SnackbarMessage, 'id'>) => void;
    hideSnackbar: (id: string) => void;
}

const SnackbarContext = createContext<SnackbarContextType | undefined>(undefined);

export const useSnackbar = () => {
    const context = useContext(SnackbarContext);
    if (!context) {
        throw new Error('useSnackbar must be used within a SnackbarProvider');
    }
    return context;
};

interface SnackbarProviderProps {
    children: React.ReactNode;
}

export const SnackbarProvider: React.FC<SnackbarProviderProps> = ({ children }) => {
    const [snackbars, setSnackbars] = useState<SnackbarMessage[]>([]);

    const showSnackbar = useCallback((message: Omit<SnackbarMessage, 'id'>) => {
        const id = Math.random().toString(36).substr(2, 9);
        const newSnackbar: SnackbarMessage = {
            ...message,
            id,
            duration: message.duration ?? 5000,
        };

        setSnackbars(prev => [...prev, newSnackbar]);

        // Auto-hide after duration
        if (newSnackbar.duration && newSnackbar.duration > 0) {
            setTimeout(() => {
                hideSnackbar(id);
            }, newSnackbar.duration);
        }
    }, []);

    const hideSnackbar = useCallback((id: string) => {
        setSnackbars(prev => prev.filter(snackbar => snackbar.id !== id));
    }, []);

    return (
        <SnackbarContext.Provider value={{ showSnackbar, hideSnackbar }}>
            {children}
            <SnackbarContainer snackbars={snackbars} onHide={hideSnackbar} />
        </SnackbarContext.Provider>
    );
};

interface SnackbarContainerProps {
    snackbars: SnackbarMessage[];
    onHide: (id: string) => void;
}

const SnackbarContainer: React.FC<SnackbarContainerProps> = ({ snackbars, onHide }) => {
    if (snackbars.length === 0) return null;

    return (
        <div className="fixed top-4 right-4 z-50 space-y-2">
            {snackbars.map((snackbar) => (
                <SnackbarItem key={snackbar.id} snackbar={snackbar} onHide={onHide} />
            ))}
        </div>
    );
};

interface SnackbarItemProps {
    snackbar: SnackbarMessage;
    onHide: (id: string) => void;
}

const SnackbarItem: React.FC<SnackbarItemProps> = ({ snackbar, onHide }) => {
    const getIcon = () => {
        switch (snackbar.type) {
            case 'success':
                return <CheckCircle className="h-5 w-5 text-green-600" />;
            case 'error':
                return <AlertCircle className="h-5 w-5 text-red-600" />;
            case 'warning':
                return <AlertCircle className="h-5 w-5 text-yellow-600" />;
            case 'info':
                return <Info className="h-5 w-5 text-blue-600" />;
            default:
                return null;
        }
    };

    const getStyles = () => {
        switch (snackbar.type) {
            case 'success':
                return 'bg-green-50 border-green-200 text-green-800 dark:bg-green-950 dark:border-green-800 dark:text-green-200';
            case 'error':
                return 'bg-red-50 border-red-200 text-red-800 dark:bg-red-950 dark:border-red-800 dark:text-red-200';
            case 'warning':
                return 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-950 dark:border-yellow-800 dark:text-yellow-200';
            case 'info':
                return 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-950 dark:border-blue-800 dark:text-blue-200';
            default:
                return 'bg-gray-50 border-gray-200 text-gray-800 dark:bg-gray-950 dark:border-gray-800 dark:text-gray-200';
        }
    };

    return (
        <div
            className={cn(
                'flex items-start space-x-3 p-4 border rounded-lg shadow-lg max-w-sm animate-in slide-in-from-right-full duration-300',
                getStyles()
            )}
        >
            <div className="flex-shrink-0 mt-0.5">
                {getIcon()}
            </div>
            <div className="flex-1 min-w-0">
                <p className="text-sm font-medium">{snackbar.title}</p>
                {snackbar.message && (
                    <p className="text-sm mt-1 opacity-90">{snackbar.message}</p>
                )}
            </div>
            <button
                onClick={() => onHide(snackbar.id)}
                className="flex-shrink-0 ml-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            >
                <X className="h-4 w-4" />
            </button>
        </div>
    );
};
