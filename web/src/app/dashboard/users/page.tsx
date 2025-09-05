'use client';

import DashboardNavbar from '@/components/DashboardNavbar';
import Users from '@/components/Users';
import { ProtectedRoute } from '@/components/ProtectedRoute';

export default function UsersPage() {
    return (
        <ProtectedRoute>
            <div className="min-h-screen bg-background">
                <DashboardNavbar />
                <main className="container mx-auto px-6 py-8">
                    <Users />
                </main>
            </div>
        </ProtectedRoute>
    );
}
