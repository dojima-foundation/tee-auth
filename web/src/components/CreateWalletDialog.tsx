'use client';

import { useState } from 'react';
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
import { Plus, Wallet } from 'lucide-react';

interface CreateWalletDialogProps {
    onWalletCreated: (walletData: {
        name: string;
        currency: string;
        address?: string;
    }) => void;
    loading?: boolean;
}

export default function CreateWalletDialog({ onWalletCreated, loading = false }: CreateWalletDialogProps) {
    const [open, setOpen] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        currency: 'ETH',
        address: ''
    });
    const [errors, setErrors] = useState<Record<string, string>>({});

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        // Reset errors
        setErrors({});

        // Validate form
        const newErrors: Record<string, string> = {};

        if (!formData.name.trim()) {
            newErrors.name = 'Wallet name is required';
        }

        if (!formData.currency) {
            newErrors.currency = 'Currency is required';
        }

        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        // Call the parent handler
        onWalletCreated(formData);

        // Reset form and close dialog
        setFormData({ name: '', currency: 'ETH', address: '' });
        setOpen(false);
    };

    const handleInputChange = (field: string, value: string) => {
        setFormData(prev => ({ ...prev, [field]: value }));
        // Clear error when user starts typing
        if (errors[field]) {
            setErrors(prev => ({ ...prev, [field]: '' }));
        }
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button disabled={loading}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Wallet
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center">
                        <Wallet className="mr-2 h-5 w-5" />
                        Create New Wallet
                    </DialogTitle>
                    <DialogDescription>
                        Create a new cryptocurrency wallet. Fill in the details below.
                    </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="name">Wallet Name</Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => handleInputChange('name', e.target.value)}
                            placeholder="Enter wallet name"
                            className={errors.name ? 'border-red-500' : ''}
                        />
                        {errors.name && (
                            <p className="text-sm text-red-500">{errors.name}</p>
                        )}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="currency">Currency</Label>
                        <Select
                            value={formData.currency}
                            onValueChange={(value) => handleInputChange('currency', value)}
                        >
                            <SelectTrigger className={errors.currency ? 'border-red-500' : ''}>
                                <SelectValue placeholder="Select a currency" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="ETH">Ethereum (ETH)</SelectItem>
                                <SelectItem value="BTC">Bitcoin (BTC)</SelectItem>
                                <SelectItem value="USDC">USD Coin (USDC)</SelectItem>
                                <SelectItem value="USDT">Tether (USDT)</SelectItem>
                                <SelectItem value="DAI">Dai (DAI)</SelectItem>
                                <SelectItem value="MATIC">Polygon (MATIC)</SelectItem>
                                <SelectItem value="SOL">Solana (SOL)</SelectItem>
                            </SelectContent>
                        </Select>
                        {errors.currency && (
                            <p className="text-sm text-red-500">{errors.currency}</p>
                        )}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="address">Wallet Address (Optional)</Label>
                        <Input
                            id="address"
                            value={formData.address}
                            onChange={(e) => handleInputChange('address', e.target.value)}
                            placeholder="Enter wallet address (optional)"
                        />
                        <p className="text-xs text-muted-foreground">
                            Leave empty to generate a new address automatically
                        </p>
                    </div>

                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => setOpen(false)}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={loading}>
                            {loading ? 'Creating...' : 'Create Wallet'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
