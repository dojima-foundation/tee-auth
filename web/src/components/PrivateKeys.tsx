'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Key, Copy, Eye, EyeOff } from 'lucide-react';
import CreatePrivateKeyDialog from './CreatePrivateKeyDialog';
import { useSnackbar } from '@/components/ui/snackbar';
import { useAuth } from '@/lib/auth-context';
import {
    fetchPrivateKeys,
    createPrivateKey,
    selectPrivateKeys,
    selectPrivateKeysLoading,
    selectPrivateKeysError,
    selectPrivateKeysPagination
} from '@/store/privateKeysSlice';
import { selectOrganizationId } from '@/store/authSlice';

export default function PrivateKeys() {
    const dispatch = useAppDispatch();
    const { isAuthenticated } = useAuth();

    // Get private keys data from Redux store
    const privateKeys = useAppSelector(selectPrivateKeys);
    const loading = useAppSelector(selectPrivateKeysLoading);
    // const error = useAppSelector(selectPrivateKeysError);
    // const pagination = useAppSelector(selectPrivateKeysPagination);

    // Get organization ID from auth store
    const organizationId = useAppSelector(selectOrganizationId);

    const [showPrivateKeys, setShowPrivateKeys] = useState(false);
    const [isCreatingKey, setIsCreatingKey] = useState(false);
    const { showSnackbar } = useSnackbar();

    useEffect(() => {

        if (organizationId && isAuthenticated) {
            dispatch(fetchPrivateKeys({}));
        } else {

        }
    }, [dispatch, organizationId, isAuthenticated]);

    const handleCreatePrivateKey = async (privateKeyData: { wallet_id: string; name: string; curve: string; tags?: string[] }) => {
        try {
            setIsCreatingKey(true);

            if (!organizationId) {
                throw new Error('No organization ID available');
            }

            // Call Redux action to create private key
            await dispatch(createPrivateKey(privateKeyData)).unwrap();

            showSnackbar({
                type: 'success',
                title: 'Private Key Created',
                message: 'Private key created successfully!'
            });

            // Refresh private keys list after creating key
            dispatch(fetchPrivateKeys({}));
        } catch (error) {
            console.error('Error creating private key:', error);
            const errorMessage = error instanceof Error ? error.message : 'Failed to create private key';

            // Handle specific error cases with better descriptions
            if (errorMessage.includes('duplicate key value violates unique constraint')) {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: `A private key named "${privateKeyData.name}" already exists in this wallet. Please choose a different name.`
                });
            } else if (errorMessage.includes('invalid curve')) {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: 'Invalid curve type selected. Please choose a valid curve from the dropdown.'
                });
            } else if (errorMessage.includes('invalid organization ID')) {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: 'Invalid organization ID. Please try logging in again.'
                });
            } else if (errorMessage.includes('wallet not found')) {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: 'Selected wallet not found. Please refresh and try again.'
                });
            } else if (errorMessage.includes('network') || errorMessage.includes('connection')) {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: 'Network connection error. Please check your internet connection and try again.'
                });
            } else {
                showSnackbar({
                    type: 'error',
                    title: 'Private Key Creation Failed',
                    message: 'An unexpected error occurred while creating the private key. Please try again or contact support.'
                });
            }
        } finally {
            setIsCreatingKey(false);
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text).then(() => {
            // You could add a toast notification here
            console.log('Key copied to clipboard');
        });
    };

    const togglePrivateKeyVisibility = () => {
        setShowPrivateKeys(!showPrivateKeys);
    };

    return (
        <div className="space-y-6">
            {/* Header with Title and Buttons */}
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold text-foreground">Private Keys</h1>
                <div className="flex space-x-2">
                    <Button
                        onClick={togglePrivateKeyVisibility}
                        variant="outline"
                    >
                        {showPrivateKeys ? (
                            <>
                                <EyeOff className="mr-2 h-4 w-4" />
                                Hide Keys
                            </>
                        ) : (
                            <>
                                <Eye className="mr-2 h-4 w-4" />
                                Show Keys
                            </>
                        )}
                    </Button>
                    <CreatePrivateKeyDialog
                        onPrivateKeyCreated={handleCreatePrivateKey}
                        disabled={isCreatingKey}
                    />
                </div>
            </div>



            {/* Private Keys Table */}
            <div className="bg-card border border-border rounded-lg overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-muted/50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Private Key
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Public Key
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Curve
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Path
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {loading ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-8 text-center">
                                        <div className="flex items-center justify-center">
                                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                                            <span className="ml-2 text-muted-foreground">Loading private keys...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : privateKeys.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-8 text-center">
                                        <div className="flex flex-col items-center">
                                            <Key className="h-12 w-12 text-muted-foreground mb-4" />
                                            <p className="text-muted-foreground">No private keys found</p>
                                            <p className="text-sm text-muted-foreground">Create your first private key to get started</p>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                privateKeys.map((key) => (
                                    <tr key={key.id} className="hover:bg-accent/50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center mr-3">
                                                    <Key className="h-4 w-4 text-primary" />
                                                </div>
                                                <div>
                                                    <div className="text-sm font-medium text-foreground">
                                                        {key.name}
                                                    </div>
                                                    <div className="text-sm text-muted-foreground">
                                                        ID: {key.id}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center space-x-2">
                                                <div className="text-sm font-mono text-foreground">
                                                    {showPrivateKeys
                                                        ? key.public_key
                                                        : `${key.public_key.slice(0, 8)}...${key.public_key.slice(-6)}`
                                                    }
                                                </div>
                                                <button
                                                    onClick={() => copyToClipboard(key.public_key)}
                                                    className="text-muted-foreground hover:text-foreground transition-colors"
                                                >
                                                    <Copy className="h-4 w-4" />
                                                </button>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm text-foreground">{key.curve}</div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm font-mono text-muted-foreground">{key.path}</div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${key.is_active
                                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                                                }`}>
                                                {key.is_active ? 'Active' : 'Inactive'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                            <div className="flex space-x-2">
                                                <button className="text-primary hover:text-primary/80 transition-colors">
                                                    View
                                                </button>
                                                <button className="text-destructive hover:text-destructive/80 transition-colors">
                                                    Delete
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
