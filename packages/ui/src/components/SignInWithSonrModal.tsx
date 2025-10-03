import { X } from 'lucide-react';
import * as React from 'react';
import { cn } from '../lib/utils';
import { SignInWithSonr } from './SignInWithSonr';

interface SignInWithSonrModalProps {
  /**
   * Whether the modal is open
   */
  isOpen: boolean;
  /**
   * Callback when modal should close
   */
  onClose: () => void;
  /**
   * OAuth client ID
   */
  clientId: string;
  /**
   * OAuth redirect URI
   */
  redirectUri: string;
  /**
   * OAuth scopes to request
   */
  scopes?: string[];
  /**
   * Modal title
   */
  title?: string;
  /**
   * Modal description
   */
  description?: string;
  /**
   * Show terms and privacy links
   */
  showTerms?: boolean;
  /**
   * Terms URL
   */
  termsUrl?: string;
  /**
   * Privacy URL
   */
  privacyUrl?: string;
  /**
   * Custom authorization URL
   */
  authorizationUrl?: string;
  /**
   * Additional content to show below the button
   */
  children?: React.ReactNode;
  /**
   * Callback when auth starts
   */
  onAuthStart?: () => void;
  /**
   * Callback on auth error
   */
  onAuthError?: (error: Error) => void;
}

/**
 * Modal component for SignInWithSonr authentication
 */
export const SignInWithSonrModal: React.FC<SignInWithSonrModalProps> = ({
  isOpen,
  onClose,
  clientId,
  redirectUri,
  scopes = ['openid', 'profile'],
  title = 'Sign in to continue',
  description = 'Use your Sonr account to securely sign in',
  showTerms = true,
  termsUrl = '/terms',
  privacyUrl = '/privacy',
  authorizationUrl,
  children,
  onAuthStart,
  onAuthError,
}) => {
  const [isLoading, setIsLoading] = React.useState(false);

  // Handle escape key
  React.useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  // Prevent body scroll when modal is open
  React.useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const handleAuthStart = () => {
    setIsLoading(true);
    if (onAuthStart) {
      onAuthStart();
    }
  };

  const handleAuthError = (error: Error) => {
    setIsLoading(false);
    if (onAuthError) {
      onAuthError(error);
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal */}
      <div
        className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 transform"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        aria-describedby="modal-description"
      >
        <div className="relative bg-white dark:bg-gray-900 rounded-xl shadow-xl">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none"
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </button>

          {/* Content */}
          <div className="p-8">
            {/* Header */}
            <div className="mb-6 text-center">
              <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-r from-purple-600 to-indigo-600">
                <SonrLogoLarge />
              </div>
              <h2
                id="modal-title"
                className="text-2xl font-semibold text-gray-900 dark:text-gray-100"
              >
                {title}
              </h2>
              <p id="modal-description" className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                {description}
              </p>
            </div>

            {/* Sign in button */}
            <SignInWithSonr
              clientId={clientId}
              redirectUri={redirectUri}
              scopes={scopes}
              authorizationUrl={authorizationUrl}
              isLoading={isLoading}
              onAuthStart={handleAuthStart}
              onAuthError={handleAuthError}
              className="w-full"
              size="lg"
            />

            {/* Alternative sign in methods */}
            <div className="mt-6">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-300 dark:border-gray-700" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-white dark:bg-gray-900 px-2 text-gray-500 dark:text-gray-400">
                    Or continue with
                  </span>
                </div>
              </div>

              {/* Alternative auth buttons */}
              <div className="mt-6 grid grid-cols-2 gap-3">
                <button
                  type="button"
                  className={cn(
                    'inline-flex w-full items-center justify-center gap-2',
                    'rounded-lg border border-gray-300 dark:border-gray-700',
                    'bg-white dark:bg-gray-800 px-4 py-2.5',
                    'text-sm font-medium text-gray-700 dark:text-gray-300',
                    'hover:bg-gray-50 dark:hover:bg-gray-700',
                    'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2',
                    'disabled:opacity-50 disabled:cursor-not-allowed'
                  )}
                  disabled={isLoading}
                >
                  <WalletIcon />
                  Wallet
                </button>
                <button
                  type="button"
                  className={cn(
                    'inline-flex w-full items-center justify-center gap-2',
                    'rounded-lg border border-gray-300 dark:border-gray-700',
                    'bg-white dark:bg-gray-800 px-4 py-2.5',
                    'text-sm font-medium text-gray-700 dark:text-gray-300',
                    'hover:bg-gray-50 dark:hover:bg-gray-700',
                    'focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2',
                    'disabled:opacity-50 disabled:cursor-not-allowed'
                  )}
                  disabled={isLoading}
                >
                  <PasskeyIcon />
                  Passkey
                </button>
              </div>
            </div>

            {/* Additional content */}
            {children && <div className="mt-6">{children}</div>}

            {/* Terms and privacy */}
            {showTerms && (
              <p className="mt-6 text-center text-xs text-gray-500 dark:text-gray-400">
                By signing in, you agree to our{' '}
                <a
                  href={termsUrl}
                  className="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Terms of Service
                </a>{' '}
                and{' '}
                <a
                  href={privacyUrl}
                  className="text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Privacy Policy
                </a>
              </p>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

/**
 * Large Sonr logo for modal header
 */
const SonrLogoLarge = () => (
  <svg
    width="32"
    height="32"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="text-white"
  >
    <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" opacity="0.9" />
    <path
      d="M2 17L12 22L22 17"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M2 12L12 17L22 12"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

/**
 * Wallet icon
 */
const WalletIcon = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="shrink-0"
  >
    <path
      d="M21 4H3C1.89543 4 1 4.89543 1 6V18C1 19.1046 1.89543 20 3 20H21C22.1046 20 23 19.1046 23 18V6C23 4.89543 22.1046 4 21 4Z"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M1 10H23"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

/**
 * Passkey icon
 */
const PasskeyIcon = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="shrink-0"
  >
    <path
      d="M21 2L19 4M15 10L17 8L21 4L15 10ZM15 10L9 16L3 20L7 14L13 8L19 2L15 10Z"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2" />
  </svg>
);

export default SignInWithSonrModal;
