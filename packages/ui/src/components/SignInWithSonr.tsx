import { type VariantProps, cva } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../lib/utils';

export const signInButtonVariants = cva(
  'inline-flex items-center justify-center gap-3 rounded-lg font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default:
          'bg-gradient-to-r from-purple-600 to-indigo-600 text-white hover:from-purple-700 hover:to-indigo-700 shadow-lg hover:shadow-xl',
        outline:
          'border-2 border-purple-600 text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-950',
        ghost: 'text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-950',
        dark: 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-gray-100 dark:text-gray-900 dark:hover:bg-gray-200',
      },
      size: {
        default: 'h-12 px-6 text-base',
        sm: 'h-10 px-4 text-sm',
        lg: 'h-14 px-8 text-lg',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

interface SignInWithSonrProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof signInButtonVariants> {
  /**
   * OAuth client ID for your application
   */
  clientId: string;
  /**
   * OAuth redirect URI after authentication
   */
  redirectUri: string;
  /**
   * OAuth scopes to request
   */
  scopes?: string[];
  /**
   * OAuth state parameter for CSRF protection
   */
  state?: string;
  /**
   * Custom authorization endpoint URL
   */
  authorizationUrl?: string;
  /**
   * Loading state
   */
  isLoading?: boolean;
  /**
   * Custom text for the button
   */
  text?: string;
  /**
   * Show Sonr logo
   */
  showLogo?: boolean;
  /**
   * Callback when authorization starts
   */
  onAuthStart?: () => void;
  /**
   * Callback on authorization error
   */
  onAuthError?: (error: Error) => void;
}

/**
 * SignInWithSonr button component for OAuth authentication
 */
export const SignInWithSonr = React.forwardRef<HTMLButtonElement, SignInWithSonrProps>(
  (
    {
      className,
      variant,
      size,
      clientId,
      redirectUri,
      scopes = ['openid', 'profile'],
      state,
      authorizationUrl = '/oauth2/authorize',
      isLoading = false,
      text = 'Sign in with Sonr',
      showLogo = true,
      onAuthStart,
      onAuthError,
      disabled,
      onClick,
      ...props
    },
    ref
  ) => {
    const handleClick = React.useCallback(
      async (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();

        if (isLoading || disabled) {
          return;
        }

        // Call custom onClick if provided
        if (onClick) {
          onClick(e);
        }

        // Call auth start callback
        if (onAuthStart) {
          onAuthStart();
        }

        try {
          // Build OAuth authorization URL
          const params = new URLSearchParams({
            response_type: 'code',
            client_id: clientId,
            redirect_uri: redirectUri,
            scope: scopes.join(' '),
            state: state || generateRandomState(),
          });

          // Add PKCE challenge for public clients
          const codeVerifier = generateCodeVerifier();
          const codeChallenge = await generateCodeChallenge(codeVerifier);

          // Store code verifier in session storage
          sessionStorage.setItem('sonr_oauth_code_verifier', codeVerifier);

          params.append('code_challenge', codeChallenge);
          params.append('code_challenge_method', 'S256');

          // Construct full authorization URL
          const fullAuthUrl = `${authorizationUrl}?${params.toString()}`;

          // Redirect to authorization endpoint
          window.location.href = fullAuthUrl;
        } catch (error) {
          if (onAuthError) {
            onAuthError(error as Error);
          }
          console.error('Failed to initiate OAuth flow:', error);
        }
      },
      [
        clientId,
        redirectUri,
        scopes,
        state,
        authorizationUrl,
        isLoading,
        disabled,
        onClick,
        onAuthStart,
        onAuthError,
      ]
    );

    return (
      <button
        ref={ref}
        className={cn(signInButtonVariants({ variant, size, className }))}
        onClick={handleClick}
        disabled={disabled || isLoading}
        type="button"
        {...props}
      >
        {isLoading ? <LoadingSpinner /> : showLogo ? <SonrLogo /> : null}
        <span>{isLoading ? 'Signing in...' : text}</span>
      </button>
    );
  }
);

SignInWithSonr.displayName = 'SignInWithSonr';

/**
 * Sonr logo SVG component
 */
const SonrLogo = () => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="shrink-0"
  >
    <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" opacity="0.8" />
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
 * Loading spinner component
 */
const LoadingSpinner = () => (
  <svg
    className="animate-spin h-5 w-5"
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
  >
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
    <path
      className="opacity-75"
      fill="currentColor"
      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
    />
  </svg>
);

/**
 * Generate random state for CSRF protection
 */
function generateRandomState(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Generate PKCE code verifier
 */
function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Generate PKCE code challenge from verifier
 */
async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

export default SignInWithSonr;
