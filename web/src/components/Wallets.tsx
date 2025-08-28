'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { walletActions } from '@/store/sagas/walletSaga';
import { useEffect, useState } from 'react';
import { Wallet, Copy } from 'lucide-react';
import CreateWalletDialog from './CreateWalletDialog';

export default function Wallets() {
    const dispatch = useAppDispatch();
    const { wallets, loading, error } = useAppSelector((state) => state.wallet);
    const [isCreatingWallet, setIsCreatingWallet] = useState(false);
    const [walletError, setWalletError] = useState<string | null>(null);
    const [walletSuccess, setWalletSuccess] = useState<string | null>(null);

    useEffect(() => {
        dispatch(walletActions.fetchWallets());
    }, [dispatch]);

    const handleCreateWallet = async (walletData: { name: string; currency: string; address?: string }) => {
        try {
            setIsCreatingWallet(true);
            setWalletError(null);
            setWalletSuccess(null);

            // Call Redux action to create wallet
            dispatch(walletActions.createWallet({
                name: walletData.name,
                currency: walletData.currency,
                address: walletData.address || `0x${Math.random().toString(36).substr(2, 40)}`
            }));

            setWalletSuccess(`Wallet "${walletData.name}" created successfully!`);

            // Refresh wallets list after creating wallet
            dispatch(walletActions.fetchWallets());
        } catch (error) {
            console.error('Error creating wallet:', error);
            setWalletError(error instanceof Error ? error.message : 'Failed to create wallet');
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
                    loading={isCreatingWallet}
                />
            </div>

            {/* Success Display */}
            {walletSuccess && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg dark:bg-green-950 dark:border-green-800">
                    <div className="text-green-800 dark:text-green-200">
                        <h3 className="font-semibold">Success:</h3>
                        <p>{walletSuccess}</p>
                    </div>
                </div>
            )}

            {/* Error Display */}
            {(error || walletError) && (
                <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
                    <div className="text-destructive">
                        <h3 className="font-semibold">Error:</h3>
                        <p>{walletError || error}</p>
                    </div>
                </div>
            )}

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
                                    Address
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Balance
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Currency
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
                                                    {wallet.address.slice(0, 8)}...{wallet.address.slice(-6)}
                                                </div>
                                                <button
                                                    onClick={() => copyToClipboard(wallet.address)}
                                                    className="text-muted-foreground hover:text-foreground transition-colors"
                                                >
                                                    <Copy className="h-4 w-4" />
                                                </button>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm font-medium text-foreground">
                                                {wallet.balance} {wallet.currency}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                                                {wallet.currency}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                                                Active
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
