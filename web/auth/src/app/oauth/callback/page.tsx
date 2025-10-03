'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@sonr.io/ui';
import { Alert, AlertDescription } from '@sonr.io/ui';
import { Button } from '@sonr.io/ui';
import { useSignInWithSonr } from '@sonr.io/ui';
import { AlertCircle, CheckCircle, Loader2, XCircle } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { Suspense } from 'react';

function CallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = React.useState<'processing' | 'success' | 'error'>('processing');
  const [message, setMessage] = React.useState<string>('');
  const [userInfo, setUserInfo] = React.useState<Record<string, unknown> | null>(null);

  // Extract OAuth callback parameters
  const code = searchParams.get('code');
  const state = searchParams.get('state');
  const error = searchParams.get('error');
  const errorDescription = searchParams.get('error_description');

  // Get stored OAuth configuration from session
  const getStoredConfig = React.useCallback(() => {
    const stored = sessionStorage.getItem('sonr_oauth_config');
    if (stored) {
      try {
        return JSON.parse(stored);
      } catch {
        return null;
      }
    }
    return null;
  }, []);

  // Initialize OAuth client
  const oauthConfig = React.useMemo(() => {
    const stored = getStoredConfig();
    if (stored) {
      return {
        clientId: stored.clientId,
        redirectUri: stored.redirectUri || `${window.location.origin}/oauth/callback`,
        authorizationUrl: '/oauth2/authorize',
        tokenUrl: '/oauth2/token',
        userInfoUrl: '/oauth2/userinfo',
        scopes: stored.scopes || ['openid', 'profile'],
      };
    }

    // Fallback config for development
    return {
      clientId: 'dev_client_123',
      redirectUri: `${window.location.origin}/oauth/callback`,
      authorizationUrl: '/oauth2/authorize',
      tokenUrl: '/oauth2/token',
      userInfoUrl: '/oauth2/userinfo',
      scopes: ['openid', 'profile'],
    };
  }, [getStoredConfig]);

  const { handleCallback } = useSignInWithSonr(oauthConfig, {
    onSuccess: (user, token) => {
      setUserInfo(user);
      setStatus('success');
      setMessage('Authentication successful! Redirecting...');

      // Store authentication data
      localStorage.setItem('sonr_auth_user', JSON.stringify(user));
      localStorage.setItem('sonr_auth_token', JSON.stringify(token));

      // Get return URL from session storage
      const returnUrl = sessionStorage.getItem('sonr_oauth_return_url');
      sessionStorage.removeItem('sonr_oauth_return_url');
      sessionStorage.removeItem('sonr_oauth_config');

      // Redirect to return URL or dashboard
      setTimeout(() => {
        if (returnUrl) {
          // If it's an external URL, use postMessage to communicate
          if (returnUrl.startsWith('http')) {
            window.opener?.postMessage(
              {
                type: 'sonr_auth_success',
                user,
                token,
              },
              new URL(returnUrl).origin
            );
            window.close();
          } else {
            router.push(returnUrl);
          }
        } else {
          router.push('/dashboard');
        }
      }, 2000);
    },
    onError: (err) => {
      setStatus('error');
      setMessage(err.message || 'Authentication failed');
    },
  });

  // Handle OAuth callback
  React.useEffect(() => {
    const processCallback = async () => {
      // Check for OAuth errors first
      if (error) {
        setStatus('error');
        setMessage(errorDescription || `OAuth error: ${error}`);
        return;
      }

      // Validate state parameter
      const storedState = sessionStorage.getItem('sonr_oauth_state');
      if (state && storedState && state !== storedState) {
        setStatus('error');
        setMessage('Invalid state parameter. Possible CSRF attack.');
        return;
      }

      // Process authorization code
      if (code) {
        try {
          await handleCallback(window.location.href);
        } catch (err) {
          console.error('Callback processing error:', err);
          setStatus('error');
          setMessage(err instanceof Error ? err.message : 'Failed to process callback');
        }
      } else {
        setStatus('error');
        setMessage('No authorization code received');
      }
    };

    processCallback();
  }, [code, state, error, errorDescription, handleCallback]);

  // Handle retry
  const handleRetry = () => {
    const returnUrl = sessionStorage.getItem('sonr_oauth_return_url') || '/dashboard';
    router.push(`/login?return_url=${encodeURIComponent(returnUrl)}`);
  };

  // Handle close for popup mode
  const handleClose = () => {
    if (window.opener) {
      window.opener.postMessage(
        {
          type: 'sonr_auth_cancelled',
        },
        '*'
      );
      window.close();
    } else {
      router.push('/');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4 bg-gradient-to-br from-background to-muted">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {status === 'processing' && (
              <>
                <Loader2 className="h-5 w-5 animate-spin" />
                Processing Authentication
              </>
            )}
            {status === 'success' && (
              <>
                <CheckCircle className="h-5 w-5 text-green-500" />
                Authentication Successful
              </>
            )}
            {status === 'error' && (
              <>
                <XCircle className="h-5 w-5 text-destructive" />
                Authentication Failed
              </>
            )}
          </CardTitle>
          <CardDescription>
            {status === 'processing' && 'Please wait while we complete your authentication...'}
            {status === 'success' && 'You have been successfully authenticated'}
            {status === 'error' && 'There was a problem with your authentication'}
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Status message */}
          {message && (
            <Alert variant={status === 'error' ? 'destructive' : 'default'}>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{message}</AlertDescription>
            </Alert>
          )}

          {/* User info display on success */}
          {status === 'success' && userInfo && (
            <div className="space-y-2 p-4 rounded-lg bg-muted/50">
              <p className="text-sm font-medium">Welcome back!</p>
              <div className="space-y-1 text-sm text-muted-foreground">
                {userInfo.name && <p>Name: {userInfo.name}</p>}
                {userInfo.email && <p>Email: {userInfo.email}</p>}
                {userInfo.did && <p className="font-mono text-xs break-all">DID: {userInfo.did}</p>}
              </div>
            </div>
          )}

          {/* Error details */}
          {status === 'error' && error && (
            <div className="space-y-1 p-3 rounded-lg bg-destructive/10 text-sm">
              <p className="font-medium">Error Code: {error}</p>
              {errorDescription && <p className="text-muted-foreground">{errorDescription}</p>}
            </div>
          )}

          {/* Loading spinner */}
          {status === 'processing' && (
            <div className="flex justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
          )}

          {/* Action buttons */}
          {status === 'error' && (
            <div className="flex gap-2">
              <Button variant="outline" onClick={handleClose} className="flex-1">
                Cancel
              </Button>
              <Button onClick={handleRetry} className="flex-1">
                Try Again
              </Button>
            </div>
          )}

          {/* Success redirect notice */}
          {status === 'success' && (
            <div className="text-center text-sm text-muted-foreground">
              <p>Redirecting you to the application...</p>
              <div className="mt-2">
                <Loader2 className="h-4 w-4 animate-spin inline" />
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default function CallbackPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <CallbackContent />
    </Suspense>
  );
}
