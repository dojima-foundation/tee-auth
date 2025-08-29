'use client';

import DashboardNavbar from '@/components/DashboardNavbar';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Home } from 'lucide-react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/ProtectedRoute';

export default function DashboardPage() {
    return (
        <ProtectedRoute>
            <div className="min-h-screen bg-background">
                <DashboardNavbar />

                <main className="container mx-auto px-6 py-8">
                    <div className="flex items-center justify-between mb-8">
                        <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
                        <Link href="/">
                            <Button variant="outline" size="sm">
                                <ArrowLeft className="mr-2 h-4 w-4" />
                                Back to Home
                            </Button>
                        </Link>
                    </div>

                    {/* Clean Main Content Area */}
                    <div className="bg-card border border-border rounded-lg p-8">
                        <div className="text-center">
                            <Home className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
                            <h2 className="text-2xl font-semibold text-foreground mb-4">
                                Welcome to Your Dashboard
                            </h2>
                            <p className="text-muted-foreground mb-6">
                                Use the menu in the top-left corner to navigate through different sections.
                            </p>

                            {/* Quick Navigation Cards */}
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-2xl mx-auto">
                                <Link href="/dashboard/users">
                                    <div className="p-4 border rounded-lg hover:bg-accent/50 transition-colors cursor-pointer">
                                        <h3 className="font-semibold text-foreground">Users</h3>
                                        <p className="text-sm text-muted-foreground">Manage application users</p>
                                    </div>
                                </Link>
                                <Link href="/dashboard/wallets">
                                    <div className="p-4 border rounded-lg hover:bg-accent/50 transition-colors cursor-pointer">
                                        <h3 className="font-semibold text-foreground">Wallets</h3>
                                        <p className="text-sm text-muted-foreground">Manage cryptocurrency wallets</p>
                                    </div>
                                </Link>
                                <Link href="/dashboard/pkeys">
                                    <div className="p-4 border rounded-lg hover:bg-accent/50 transition-colors cursor-pointer">
                                        <h3 className="font-semibold text-foreground">Private Keys</h3>
                                        <p className="text-sm text-muted-foreground">Manage cryptographic keys</p>
                                    </div>
                                </Link>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </ProtectedRoute>
    );
}
