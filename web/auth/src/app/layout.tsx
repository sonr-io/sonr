import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Highway - Sonr Authentication Gateway',
  description:
    'Passwordless authentication using WebAuthn passkeys for the Sonr blockchain ecosystem',
  keywords: ['webauthn', 'passkey', 'authentication', 'sonr', 'blockchain'],
  authors: [{ name: 'Sonr' }],
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={inter.className}>{children}</body>
    </html>
  );
}
