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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Key, Plus } from 'lucide-react';
import { useAppSelector } from '@/store/hooks';
import { selectWallets } from '@/store/walletsSlice';

interface CreatePrivateKeyDialogProps {
    onPrivateKeyCreated: (data: { wallet_id: string; name: string; curve: string; tags?: string[] }) => void;
    disabled?: boolean;
}

export default function CreatePrivateKeyDialog({ onPrivateKeyCreated, disabled = false }: CreatePrivateKeyDialogProps) {
    const [open, setOpen] = useState(false);
    const [walletId, setWalletId] = useState('');
    const [name, setName] = useState('');
    const [curve, setCurve] = useState('');
    const [tags, setTags] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const wallets = useAppSelector(selectWallets);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!walletId.trim()) {
            setError('Wallet is required');
            return;
        }
        if (!name.trim()) {
            setError('Private key name is required');
            return;
        }
        if (!curve.trim()) {
            setError('Curve is required');
            return;
        }


        setLoading(true);
        setError(null);

        try {
            // This will be handled by the parent component
            onPrivateKeyCreated({
                wallet_id: walletId,
                name,
                curve,
                tags: parseTags(tags)
            });
            setOpen(false);
            resetForm();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create private key');
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setWalletId('');
        setName('');
        setCurve('');
        setTags('');
        setError(null);
    };

    const handleOpenChange = (newOpen: boolean) => {
        if (!newOpen) {
            resetForm();
        }
        setOpen(newOpen);
    };

    const parseTags = (tagsString: string): string[] => {
        return tagsString.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
                <Button disabled={disabled} className="flex items-center space-x-2">
                    <Plus className="h-4 w-4" />
                    <span>Create Private Key</span>
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center space-x-2">
                        <Key className="h-5 w-5" />
                        <span>Create New Private Key</span>
                    </DialogTitle>
                    <DialogDescription>
                        Create a new private key for a wallet. Select the wallet and provide the key details.
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="wallet-select">Wallet</Label>
                        <Select value={walletId} onValueChange={setWalletId} disabled={loading}>
                            <SelectTrigger id="wallet-select" aria-label="Wallet">
                                <SelectValue placeholder="Select a wallet" />
                            </SelectTrigger>
                            <SelectContent>
                                {wallets.map((wallet) => (
                                    <SelectItem key={wallet.id} value={wallet.id}>
                                        {wallet.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="private-key-name">Private Key Name</Label>
                        <Input
                            id="private-key-name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="Enter private key name"
                            disabled={loading}
                            required
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="curve-select">Curve</Label>
                        <Select value={curve} onValueChange={setCurve} disabled={loading}>
                            <SelectTrigger id="curve-select" aria-label="Curve">
                                <SelectValue placeholder="Select a curve" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="CURVE_SECP256K1">SECP256K1 (Ethereum)</SelectItem>
                                <SelectItem value="CURVE_ED25519">ED25519 (Solana)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="tags">Tags (Optional)</Label>
                        <Input
                            id="tags"
                            value={tags}
                            onChange={(e) => setTags(e.target.value)}
                            placeholder="Enter tags separated by commas"
                            disabled={loading}
                        />
                        <p className="text-xs text-muted-foreground">
                            Optional tags to categorize the private key
                        </p>
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
                        <Button type="submit" disabled={loading || !walletId || !name || !curve}>
                            {loading ? 'Creating...' : 'Create Private Key'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}

