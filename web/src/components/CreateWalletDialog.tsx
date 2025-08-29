'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Wallet, Plus } from 'lucide-react';

interface CreateWalletDialogProps {
    onWalletCreated: (walletData: { name: string }) => void;
    disabled?: boolean;
}

export default function CreateWalletDialog({ onWalletCreated, disabled = false }: CreateWalletDialogProps) {
    const [open, setOpen] = useState(false);
    const [name, setName] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!name.trim()) {
            setError('Wallet name is required');
            return;
        }

        setLoading(true);
        setError(null);

        try {
            // This will be handled by the parent component
            onWalletCreated({ name: name.trim() });
            setOpen(false);
            setName('');
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create wallet');
        } finally {
            setLoading(false);
        }
    };

    const handleOpenChange = (newOpen: boolean) => {
        if (!newOpen) {
            setName('');
            setError(null);
        }
        setOpen(newOpen);
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
                <Button disabled={disabled} className="flex items-center space-x-2">
                    <Plus className="h-4 w-4" />
                    <span>Create Wallet</span>
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center space-x-2">
                        <Wallet className="h-5 w-5" />
                        <span>Create New Wallet</span>
                    </DialogTitle>
                    <DialogDescription>
                        Create a new HD wallet for your organization. You can optionally provide a seed phrase.
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="wallet-name">Wallet Name</Label>
                        <Input
                            id="wallet-name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="Enter wallet name"
                            disabled={loading}
                            required
                        />
                    </div>

                    {error && (
                        <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                            <p className="text-sm text-destructive">{error}</p>
                        </div>
                    )}
                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setOpen(false)}
                            disabled={loading}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={loading || !name.trim()}>
                            {loading ? 'Creating...' : 'Create Wallet'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
