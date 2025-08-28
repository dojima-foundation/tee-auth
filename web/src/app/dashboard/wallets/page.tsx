'use client';

import DashboardNavbar from '@/components/DashboardNavbar';
import Wallets from '@/components/Wallets';

export default function WalletsPage() {
    return (
        <div className="min-h-screen bg-background">
            <DashboardNavbar />
            <main className="container mx-auto px-6 py-8">
                <Wallets />
            </main>
        </div>
    );
}
