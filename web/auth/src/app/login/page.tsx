'use client';

import { useWebAuthn } from '@/hooks/useWebAuthn';
import { Button, ErrorAlert, Input } from '@sonr.io/ui';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [isSupported, setIsSupported] = useState<boolean | null>(null);
  const { authenticateUser, isLoading, error, clearError } = useWebAuthn();
  const router = useRouter();

  // Check WebAuthn support on component mount
  useState(() => {
    const checkSupport = async () => {
      if (typeof window !== 'undefined') {
        const supported =
          window.PublicKeyCredential &&
          typeof window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable ===
            'function';
        setIsSupported(supported);
      }
    };
    checkSupport();
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();

    if (!username.trim()) {
      return;
    }

    try {
      const success = await authenticateUser(username.trim());
      if (success) {
        router.push('/dashboard');
      }
    } catch (_err) {
      // Error is handled by the hook
    }
  };

  if (isSupported === false) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div>
            <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
              WebAuthn Not Supported
            </h2>
            <p className="mt-2 text-center text-sm text-gray-600">
              Your browser doesn't support WebAuthn. Please use a modern browser like Chrome,
              Firefox, Safari, or Edge.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-green-100">
            <svg
              className="h-6 w-6 text-green-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
              role="img"
              aria-label="Authentication Successful Icon"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Use your passkey to securely access your account
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && <ErrorAlert message={error} onDismiss={clearError} />}

          <div>
            <Input
              label="Username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              required
              autoComplete="username"
              helpText="Enter the username you registered with"
            />
          </div>

          <div>
            <Button
              type="submit"
              size="lg"
              isLoading={isLoading}
              disabled={!username.trim() || isLoading}
              className="w-full"
            >
              {isLoading ? 'Signing In...' : 'Sign In with Passkey'}
            </Button>
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <a href="/register" className="font-medium text-blue-600 hover:text-blue-500">
                Create one here
              </a>
            </p>
          </div>
        </form>

        <div className="mt-6">
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-gray-50 text-gray-500">Secure & Simple</span>
            </div>
          </div>
          <div className="mt-4 text-sm text-gray-600 space-y-2">
            <p>• No passwords to type or remember</p>
            <p>• Authenticate with your device's biometrics</p>
            <p>• Protected against phishing and breaches</p>
          </div>
        </div>
      </div>
    </div>
  );
}
