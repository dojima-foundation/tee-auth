import Link from 'next/link';

export default function Home() {
    return (
        <main className="min-h-screen bg-background text-foreground">
            <div className="container mx-auto px-4 py-8">
                <h1 className="text-4xl font-bold mb-8">Welcome</h1>
                <p className="text-muted-foreground text-lg mb-8">
                    Your application is ready for development.
                </p>

                <Link
                    href="/dashboard"
                    className="inline-block bg-primary hover:bg-primary/90 text-primary-foreground font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
                >
                    Go to Dashboard
                </Link>
            </div>
        </main>
    );
}
