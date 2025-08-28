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
import { Plus, Key, Shield } from 'lucide-react';

interface CreatePrivateKeyDialogProps {
    onPrivateKeyCreated: (keyData: {
        name: string;
        type: string;
        walletId?: string;
    }) => void;
    loading?: boolean;
}

export default function CreatePrivateKeyDialog({ onPrivateKeyCreated, loading = false }: CreatePrivateKeyDialogProps) {
    const [open, setOpen] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        type: 'secp256k1',
        walletId: ''
    });
    const [errors, setErrors] = useState<Record<string, string>>({});

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        // Reset errors
        setErrors({});

        // Validate form
        const newErrors: Record<string, string> = {};

        if (!formData.name.trim()) {
            newErrors.name = 'Private key name is required';
        }

        if (!formData.type) {
            newErrors.type = 'Key type is required';
        }

        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        // Call the parent handler
        onPrivateKeyCreated(formData);

        // Reset form and close dialog
        setFormData({ name: '', type: 'secp256k1', walletId: '' });
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
                    Create Private Key
                </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center">
                        <Key className="mr-2 h-5 w-5" />
                        Create New Private Key
                    </DialogTitle>
                    <DialogDescription>
                        Generate a new cryptographic private key. This will be encrypted and stored securely.
                    </DialogDescription>
                </DialogHeader>

                {/* Security Warning */}
                <div className="flex items-start space-x-3 p-3 bg-orange-50 border border-orange-200 rounded-lg dark:bg-orange-950 dark:border-orange-800">
                    <Shield className="h-5 w-5 text-orange-600 dark:text-orange-400 mt-0.5" />
                    <div>
                        <h4 className="font-semibold text-orange-800 dark:text-orange-200">Security Notice</h4>
                        <p className="text-sm text-orange-700 dark:text-orange-300">
                            Private keys are sensitive cryptographic material. They will be encrypted and stored securely.
                        </p>
                    </div>
                </div>

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="name">Private Key Name</Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => handleInputChange('name', e.target.value)}
                            placeholder="Enter a descriptive name for this key"
                            className={errors.name ? 'border-red-500' : ''}
                        />
                        {errors.name && (
                            <p className="text-sm text-red-500">{errors.name}</p>
                        )}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="type">Key Type</Label>
                        <Select
                            value={formData.type}
                            onValueChange={(value) => handleInputChange('type', value)}
                        >
                            <SelectTrigger className={errors.type ? 'border-red-500' : ''}>
                                <SelectValue placeholder="Select a key type" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="secp256k1">Secp256k1 (Bitcoin/Ethereum)</SelectItem>
                                <SelectItem value="ed25519">Ed25519 (Solana/Polkadot)</SelectItem>
                                <SelectItem value="rsa2048">RSA 2048</SelectItem>
                                <SelectItem value="rsa4096">RSA 4096</SelectItem>
                                <SelectItem value="p256">P-256 (NIST)</SelectItem>
                                <SelectItem value="p384">P-384 (NIST)</SelectItem>
                            </SelectContent>
                        </Select>
                        {errors.type && (
                            <p className="text-sm text-red-500">{errors.type}</p>
                        )}
                    </div>

                    <div className="space-y-2">
                        <Label htmlFor="walletId">Associated Wallet (Optional)</Label>
                        <Input
                            id="walletId"
                            value={formData.walletId}
                            onChange={(e) => handleInputChange('walletId', e.target.value)}
                            placeholder="Enter wallet ID to associate with this key"
                        />
                        <p className="text-xs text-muted-foreground">
                            Associate this private key with a specific wallet for better organization
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
                            {loading ? 'Creating...' : 'Create Private Key'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
