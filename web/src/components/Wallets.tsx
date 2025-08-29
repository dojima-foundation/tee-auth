'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { useEffect, useState } from 'react';
import { Wallet, Copy } from 'lucide-react';
import CreateWalletDialog from './CreateWalletDialog';
import { useSnackbar } from '@/components/ui/snackbar';
import {
    fetchWallets,
    createWallet,
    selectWallets,
    selectWalletsLoading,
    selectWalletsError,
    selectWalletsPagination
} from '@/store/walletsSlice';
import { selectOrganizationId } from '@/store/authSlice';

export default function Wallets() {
    const dispatch = useAppDispatch();

    // Get wallets data from Redux store
    const wallets = useAppSelector(selectWallets);
    const loading = useAppSelector(selectWalletsLoading);
    const error = useAppSelector(selectWalletsError);
    const pagination = useAppSelector(selectWalletsPagination);

    // Get organization ID from auth store
    const organizationId = useAppSelector(selectOrganizationId);

    const [isCreatingWallet, setIsCreatingWallet] = useState(false);
    const { showSnackbar } = useSnackbar();

    useEffect(() => {
        if (organizationId) {
            dispatch(fetchWallets({ organizationId }));
        }
    }, [dispatch, organizationId]);

    const handleCreateWallet = async (walletData: { name: string }) => {
        try {
            setIsCreatingWallet(true);

            if (!organizationId) {
                throw new Error('No organization ID available');
            }

            // Call Redux action to create wallet
            await dispatch(createWallet({
                organizationId,
                walletData: {
                    name: walletData.name,
                    seed_phrase: undefined,
                }
            })).unwrap();

            showSnackbar({
                type: 'success',
                title: 'Wallet Created',
                message: `Wallet "${walletData.name}" created successfully!`
            });

            // Refresh wallets list after creating wallet
            dispatch(fetchWallets({ organizationId }));
        } catch (error) {
            console.error('Error creating wallet:', error);
            const errorMessage = error instanceof Error ? error.message : 'Failed to create wallet';

            // Handle specific error cases with better descriptions
            if (errorMessage.includes('duplicate key value violates unique constraint')) {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: `A wallet named "${walletData.name}" already exists in your organization. Please choose a different name.`
                });
            } else if (errorMessage.includes('invalid organization ID')) {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: 'Invalid organization ID. Please try logging in again.'
                });
            } else if (errorMessage.includes('failed to generate seed')) {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: 'Failed to generate wallet seed. Please try again or contact support.'
                });
            } else if (errorMessage.includes('failed to derive address')) {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: 'Failed to generate wallet addresses. Please try again.'
                });
            } else if (errorMessage.includes('network') || errorMessage.includes('connection')) {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: 'Network connection error. Please check your internet connection and try again.'
                });
            } else {
                showSnackbar({
                    type: 'error',
                    title: 'Wallet Creation Failed',
                    message: 'An unexpected error occurred while creating the wallet. Please try again or contact support.'
                });
            }
        } finally {
            setIsCreatingWallet(false);
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text).then(() => {
            // You could add a toast notification here
            console.log('Address copied to clipboard');
        });
    };

    return (
        <div className="space-y-6">
            {/* Header with Title and Create Button */}
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold text-foreground">Wallets</h1>
                <CreateWalletDialog
                    onWalletCreated={handleCreateWallet}
                    disabled={isCreatingWallet}
                />
            </div>



            {/* Wallets Table */}
            <div className="bg-card border border-border rounded-lg overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-muted/50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Wallet
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Public Key
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Created
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
                                            <span className="ml-2 text-muted-foreground">Loading wallets...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : wallets.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-8 text-center">
                                        <div className="flex flex-col items-center">
                                            <Wallet className="h-12 w-12 text-muted-foreground mb-4" />
                                            <p className="text-muted-foreground">No wallets found</p>
                                            <p className="text-sm text-muted-foreground">Create your first wallet to get started</p>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                wallets.map((wallet) => (
                                    <tr key={wallet.id} className="hover:bg-accent/50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center mr-3">
                                                    <Wallet className="h-4 w-4 text-primary" />
                                                </div>
                                                <div>
                                                    <div className="text-sm font-medium text-foreground">
                                                        {wallet.name}
                                                    </div>
                                                    <div className="text-sm text-muted-foreground">
                                                        ID: {wallet.id}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center space-x-2">
                                                <div className="text-sm font-mono text-foreground">
                                                    {wallet.public_key.slice(0, 8)}...{wallet.public_key.slice(-6)}
                                                </div>
                                                <button
                                                    onClick={() => copyToClipboard(wallet.public_key)}
                                                    className="text-muted-foreground hover:text-foreground transition-colors"
                                                >
                                                    <Copy className="h-4 w-4" />
                                                </button>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                            {new Date(wallet.created_at).toLocaleDateString()}
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${wallet.is_active
                                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                                                }`}>
                                                {wallet.is_active ? 'Active' : 'Inactive'}
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
