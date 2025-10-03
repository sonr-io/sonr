import { DashboardSidebar, SidebarInset, SidebarProvider, SidebarTrigger } from '@sonr.io/ui';
import type { Metadata } from 'next';
import { AuthWrapper } from '../components/auth-wrapper';
import { ThemeProvider } from '../components/theme-provider';
import './globals.css';

export const metadata: Metadata = {
  title: 'Sonr Developer Dashboard',
  description: 'Manage your Sonr Services, domains, and analytics',
};

export default function RootLayout({ children }: { children: React.ReactNode }): React.JSX.Element {
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body className="min-h-screen bg-background font-sans antialiased">
        <ThemeProvider defaultTheme="system" storageKey="sonr-dashboard-theme">
          <AuthWrapper>
            <SidebarProvider>
              <div className="relative flex min-h-screen w-full">
                <DashboardSidebar />
                <SidebarInset>
                  <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
                    <SidebarTrigger className="-ml-1" />
                    <div className="flex items-center gap-2 px-3">
                      <h1 className="text-lg font-semibold">Sonr Dashboard</h1>
                    </div>
                  </header>
                  <div className="flex-1 px-4 py-6 lg:px-8">{children}</div>
                </SidebarInset>
              </div>
            </SidebarProvider>
          </AuthWrapper>
        </ThemeProvider>
      </body>
    </html>
  );
}
