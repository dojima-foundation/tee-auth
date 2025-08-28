'use client';

import DashboardNavbar from '@/components/DashboardNavbar';
import PrivateKeys from '@/components/PrivateKeys';

export default function PrivateKeysPage() {
    return (
        <div className="min-h-screen bg-background">
            <DashboardNavbar />
            <main className="container mx-auto px-6 py-8">
                <PrivateKeys />
            </main>
        </div>
    );
}
