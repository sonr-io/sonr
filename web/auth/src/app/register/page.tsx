'use client';

import { useWebAuthn } from '@/hooks/useWebAuthn';
import { Button, ErrorAlert, Input } from '@sonr.io/ui';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function RegisterPage() {
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [isSupported, setIsSupported] = useState<boolean | null>(null);
  const { registerUser, isLoading, error, clearError } = useWebAuthn();
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
      const success = await registerUser(username.trim(), displayName.trim() || undefined);
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
          <div className="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
            <svg
              className="h-6 w-6 text-blue-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
              role="img"
              aria-label="Account Creation Icon"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              />
            </svg>
          </div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Create your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Register with a passkey for secure, passwordless authentication
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {error && <ErrorAlert message={error} onDismiss={clearError} />}

          <div className="space-y-4">
            <Input
              label="Username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              required
              autoComplete="username"
              helpText="Choose a unique username for your account"
            />

            <Input
              label="Display Name (Optional)"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Enter your display name"
              autoComplete="name"
              helpText="This will be shown in your profile"
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
              {isLoading ? 'Creating Account...' : 'Create Account with Passkey'}
            </Button>
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Already have an account?{' '}
              <a href="/login" className="font-medium text-blue-600 hover:text-blue-500">
                Sign in here
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
              <span className="px-2 bg-gray-50 text-gray-500">What is a passkey?</span>
            </div>
          </div>
          <div className="mt-4 text-sm text-gray-600 space-y-2">
            <p>• No passwords to remember or type</p>
            <p>• Uses your device's built-in security (fingerprint, face, PIN)</p>
            <p>• More secure than traditional passwords</p>
            <p>• Works across all your devices</p>
          </div>
        </div>
      </div>
    </div>
  );
}
