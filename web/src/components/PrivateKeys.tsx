'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { walletActions } from '@/store/sagas/walletSaga';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Key, Copy, Eye, EyeOff } from 'lucide-react';
import CreatePrivateKeyDialog from './CreatePrivateKeyDialog';

export default function PrivateKeys() {
    const dispatch = useAppDispatch();
    const { privateKeys, loading, error } = useAppSelector((state) => state.wallet);
    const [showPrivateKeys, setShowPrivateKeys] = useState(false);
    const [isCreatingKey, setIsCreatingKey] = useState(false);
    const [keyError, setKeyError] = useState<string | null>(null);
    const [keySuccess, setKeySuccess] = useState<string | null>(null);

    useEffect(() => {
        dispatch(walletActions.fetchPrivateKeys());
    }, [dispatch]);

    const handleCreatePrivateKey = async (keyData: { name: string; type: string; walletId?: string }) => {
        try {
            setIsCreatingKey(true);
            setKeyError(null);
            setKeySuccess(null);

            // Call Redux action to create private key
            dispatch(walletActions.createPrivateKey({
                name: keyData.name,
                publicKey: `0x${Math.random().toString(36).substr(2, 40)}`,
                encryptedPrivateKey: `encrypted_${Math.random().toString(36).substr(2, 20)}`,
                walletId: keyData.walletId || 'wallet_1'
            }));

            setKeySuccess(`Private key "${keyData.name}" created successfully!`);

            // Refresh private keys list after creating key
            dispatch(walletActions.fetchPrivateKeys());
        } catch (error) {
            console.error('Error creating private key:', error);
            setKeyError(error instanceof Error ? error.message : 'Failed to create private key');
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
                        loading={isCreatingKey}
                    />
                </div>
            </div>

            {/* Success Display */}
            {keySuccess && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg dark:bg-green-950 dark:border-green-800">
                    <div className="text-green-800 dark:text-green-200">
                        <h3 className="font-semibold">Success:</h3>
                        <p>{keySuccess}</p>
                    </div>
                </div>
            )}

            {/* Error Display */}
            {(error || keyError) && (
                <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
                    <div className="text-destructive">
                        <h3 className="font-semibold">Error:</h3>
                        <p>{keyError || error}</p>
                    </div>
                </div>
            )}

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
                                    Wallet ID
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
                                    <td colSpan={5} className="px-6 py-8 text-center">
                                        <div className="flex items-center justify-center">
                                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                                            <span className="ml-2 text-muted-foreground">Loading private keys...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : privateKeys.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center">
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
                                                        ? key.publicKey
                                                        : `${key.publicKey.slice(0, 8)}...${key.publicKey.slice(-6)}`
                                                    }
                                                </div>
                                                <button
                                                    onClick={() => copyToClipboard(key.publicKey)}
                                                    className="text-muted-foreground hover:text-foreground transition-colors"
                                                >
                                                    <Copy className="h-4 w-4" />
                                                </button>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm text-foreground">{key.walletId}</div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${key.isActive
                                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                : 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200'
                                                }`}>
                                                {key.isActive ? 'Active' : 'Inactive'}
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
