'use client';

import DashboardNavbar from '@/components/DashboardNavbar';
import Users from '@/components/Users';

export default function UsersPage() {
    return (
        <div className="min-h-screen bg-background">
            <DashboardNavbar />
            <main className="container mx-auto px-6 py-8">
                <Users />
            </main>
        </div>
    );
}
