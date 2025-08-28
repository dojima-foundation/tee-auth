'use client';

import { ThemeToggle } from './ThemeToggle';
import { useAppSelector } from '@/store/hooks';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuShortcut,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Users, Wallet, Key, Settings, Database, Shield } from 'lucide-react';
import { useRouter } from 'next/navigation';

// Custom 9-dot icon component
const NineDotsIcon = ({ className }: { className?: string }) => (
    <svg
        className={className}
        viewBox="0 0 24 24"
        fill="currentColor"
    >
        <circle cx="4" cy="4" r="2" />
        <circle cx="12" cy="4" r="2" />
        <circle cx="20" cy="4" r="2" />
        <circle cx="4" cy="12" r="2" />
        <circle cx="12" cy="12" r="2" />
        <circle cx="20" cy="12" r="2" />
        <circle cx="4" cy="20" r="2" />
        <circle cx="12" cy="20" r="2" />
        <circle cx="20" cy="20" r="2" />
    </svg>
);

const DashboardNavbar = () => {
    const router = useRouter();
    const { loading: userLoading } = useAppSelector((state) => state.user);
    const { loading: walletLoading } = useAppSelector((state) => state.wallet);

    const handleMenuItemClick = (item: string) => {
        switch (item) {
            case 'Users':
                router.push('/dashboard/users');
                break;
            case 'Wallet':
                router.push('/dashboard/wallets');
                break;
            case 'Private Keys':
                router.push('/dashboard/pkeys');
                break;
            default:
                console.log(`Clicked on ${item}`);
        }
    };

    return (
        <nav className="bg-card border-b border-border px-6 py-4">
            <div className="flex items-center justify-between">
                {/* Left side - Logo and 9 dots menu */}
                <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-2">
                        <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                            <span className="text-primary-foreground font-bold text-lg">E</span>
                        </div>
                        <span className="text-xl font-semibold text-foreground">ODEYS</span>
                    </div>

                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <button
                                className="p-2 hover:bg-accent rounded-lg transition-colors duration-200"
                                disabled={userLoading || walletLoading}
                            >
                                <NineDotsIcon className="h-5 w-5 text-muted-foreground" />
                            </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="w-64" align="start">
                            <DropdownMenuLabel>Application Menu</DropdownMenuLabel>
                            <DropdownMenuSeparator />

                            <DropdownMenuGroup>
                                <DropdownMenuItem
                                    onClick={() => handleMenuItemClick('Users')}
                                    disabled={userLoading}
                                    className="cursor-pointer"
                                >
                                    <Users className="mr-2 h-4 w-4" />
                                    <span>Users Management</span>
                                    {userLoading && (
                                        <DropdownMenuShortcut>Loading...</DropdownMenuShortcut>
                                    )}
                                </DropdownMenuItem>

                                <DropdownMenuItem
                                    onClick={() => handleMenuItemClick('Wallet')}
                                    disabled={walletLoading}
                                    className="cursor-pointer"
                                >
                                    <Wallet className="mr-2 h-4 w-4" />
                                    <span>Wallet Management</span>
                                    {walletLoading && (
                                        <DropdownMenuShortcut>Loading...</DropdownMenuShortcut>
                                    )}
                                </DropdownMenuItem>

                                <DropdownMenuItem
                                    onClick={() => handleMenuItemClick('Private Keys')}
                                    disabled={walletLoading}
                                    className="cursor-pointer"
                                >
                                    <Key className="mr-2 h-4 w-4" />
                                    <span>Private Keys</span>
                                    {walletLoading && (
                                        <DropdownMenuShortcut>Loading...</DropdownMenuShortcut>
                                    )}
                                </DropdownMenuItem>
                            </DropdownMenuGroup>

                            <DropdownMenuSeparator />

                            <DropdownMenuGroup>
                                <DropdownMenuItem className="cursor-pointer">
                                    <Database className="mr-2 h-4 w-4" />
                                    <span>Database</span>
                                    <DropdownMenuShortcut>⌘D</DropdownMenuShortcut>
                                </DropdownMenuItem>

                                <DropdownMenuItem className="cursor-pointer">
                                    <Shield className="mr-2 h-4 w-4" />
                                    <span>Security</span>
                                    <DropdownMenuShortcut>⌘S</DropdownMenuShortcut>
                                </DropdownMenuItem>

                                <DropdownMenuItem className="cursor-pointer">
                                    <Settings className="mr-2 h-4 w-4" />
                                    <span>Settings</span>
                                    <DropdownMenuShortcut>⌘,</DropdownMenuShortcut>
                                </DropdownMenuItem>
                            </DropdownMenuGroup>

                            <DropdownMenuSeparator />

                            <DropdownMenuItem className="cursor-pointer text-destructive focus:text-destructive">
                                <span>Logout</span>
                                <DropdownMenuShortcut>⇧⌘Q</DropdownMenuShortcut>
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>

                {/* Right side - Theme toggle */}
                <div className="flex items-center space-x-2">
                    <ThemeToggle />
                </div>
            </div>
        </nav>
    );
};

export default DashboardNavbar;
