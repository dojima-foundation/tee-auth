import Link from 'next/link';

export default function Home() {
    return (
        <main className="min-h-screen bg-background text-foreground">
            <div className="container mx-auto px-4 py-8">
                <h1 className="text-4xl font-bold mb-8">Welcome to ODEYS</h1>
                <p className="text-muted-foreground text-lg mb-8">
                    Secure authentication and user management platform.
                </p>

                <div className="flex flex-col sm:flex-row gap-4">
                    <Link
                        href="/auth/signin"
                        className="inline-block bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-3 px-6 rounded-lg transition-colors duration-200 text-center"
                    >
                        Sign In
                    </Link>

                    <Link
                        href="/dashboard"
                        className="inline-block bg-secondary hover:bg-secondary/80 text-secondary-foreground font-semibold py-3 px-6 rounded-lg transition-colors duration-200 text-center"
                    >
                        Go to Dashboard
                    </Link>
                </div>
            </div>
        </main>
    );
}
