import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/ThemeProvider";
import { ReduxProvider } from '@/components/ReduxProvider';
import { AuthProvider } from '@/lib/auth-context';
import { SessionMiddleware } from '@/lib/session-middleware';
import { AuthWrapper } from '@/components/AuthWrapper';
import { SnackbarProvider } from '@/components/ui/snackbar';

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "ODEYS - Secure Authentication Platform",
  description: "Secure authentication and user management platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ReduxProvider>
          <ThemeProvider defaultTheme="system" storageKey="ui-theme">
            <AuthProvider>
              <AuthWrapper>
                <SessionMiddleware>
                  <SnackbarProvider>
                    {children}
                  </SnackbarProvider>
                </SessionMiddleware>
              </AuthWrapper>
            </AuthProvider>
          </ThemeProvider>
        </ReduxProvider>
      </body>
    </html>
  );
}
