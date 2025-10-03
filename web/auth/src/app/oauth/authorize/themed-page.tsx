'use client';

import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@sonr.io/ui';
import { Button } from '@sonr.io/ui';
import { Alert, AlertDescription } from '@sonr.io/ui';
import { OAuth2Client, parseCallbackUrl } from '@sonr.io/ui';
import { AlertCircle, Check, Shield } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import React from 'react';

interface ClientBranding {
  id: string;
  name: string;
  description?: string;
  logo?: string;
  verified: boolean;
  theme?: {
    primaryColor?: string;
    accentColor?: string;
    backgroundColor?: string;
    cardBackground?: string;
    borderRadius?: string;
    fontFamily?: string;
  };
  customCSS?: string;
}

export default function ThemedAuthorizePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = React.useState(false);

  // Extract OAuth parameters
  const clientId = searchParams.get('client_id');
  const redirectUri = searchParams.get('redirect_uri');
  const responseType = searchParams.get('response_type');
  const scope = searchParams.get('scope');
  const state = searchParams.get('state');
  const codeChallenge = searchParams.get('code_challenge');
  const codeChallengeMethod = searchParams.get('code_challenge_method');

  // Get client branding information (in production, fetch from server)
  const clientBranding = React.useMemo<ClientBranding>(() => {
    // Mock client data with branding - replace with actual client registry lookup
    const clients: Record<string, ClientBranding> = {
      branded_app: {
        id: 'branded_app',
        name: 'Branded Application',
        logo: '/logos/branded-app.svg',
        verified: true,
        theme: {
          primaryColor: '#4F46E5',
          accentColor: '#7C3AED',
          backgroundColor: '#F9FAFB',
          cardBackground: '#FFFFFF',
          borderRadius: '12px',
          fontFamily: '"Inter", system-ui, sans-serif',
        },
        customCSS: `
          .authorize-card {
            box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
          }
          .scope-item {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
          }
        `,
      },
      minimal_app: {
        id: 'minimal_app',
        name: 'Minimal Application',
        verified: false,
        theme: {
          primaryColor: '#000000',
          accentColor: '#666666',
          backgroundColor: '#FFFFFF',
          cardBackground: '#FAFAFA',
          borderRadius: '4px',
          fontFamily: 'monospace',
        },
      },
      default: {
        id: clientId || 'unknown',
        name: 'Unknown Application',
        verified: false,
      },
    };

    return clients[clientId || ''] || clients.default;
  }, [clientId]);

  // Apply custom theme
  React.useEffect(() => {
    if (clientBranding.theme) {
      const theme = clientBranding.theme;
      const root = document.documentElement;

      if (theme.primaryColor) {
        root.style.setProperty('--brand-primary', theme.primaryColor);
      }
      if (theme.accentColor) {
        root.style.setProperty('--brand-accent', theme.accentColor);
      }
      if (theme.backgroundColor) {
        root.style.setProperty('--brand-background', theme.backgroundColor);
      }
      if (theme.cardBackground) {
        root.style.setProperty('--brand-card', theme.cardBackground);
      }
      if (theme.borderRadius) {
        root.style.setProperty('--brand-radius', theme.borderRadius);
      }
      if (theme.fontFamily) {
        root.style.setProperty('--brand-font', theme.fontFamily);
      }
    }

    // Apply custom CSS if provided
    if (clientBranding.customCSS) {
      const styleElement = document.createElement('style');
      styleElement.textContent = clientBranding.customCSS;
      document.head.appendChild(styleElement);

      return () => {
        document.head.removeChild(styleElement);
      };
    }
  }, [clientBranding]);

  // Check if user is authenticated
  React.useEffect(() => {
    const checkAuth = async () => {
      try {
        const token = localStorage.getItem('sonr_auth_token');
        if (token) {
          setIsAuthenticated(true);
        }
      } catch (err) {
        console.error('Auth check failed:', err);
      }
    };

    checkAuth();
  }, []);

  // Validate request parameters
  React.useEffect(() => {
    if (!clientId || !redirectUri || !responseType) {
      setError('Missing required OAuth parameters');
      return;
    }

    if (responseType !== 'code' && responseType !== 'token') {
      setError('Invalid response type. Only "code" and "token" are supported.');
      return;
    }

    if (!codeChallenge && responseType === 'code') {
      setError('PKCE code challenge is required for public clients');
      return;
    }

    if (codeChallengeMethod && codeChallengeMethod !== 'S256') {
      setError('Only S256 code challenge method is supported');
      return;
    }
  }, [clientId, redirectUri, responseType, codeChallenge, codeChallengeMethod]);

  // Handle authorization approval
  const handleApprove = React.useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      if (!isAuthenticated) {
        const returnUrl = `/oauth/authorize?${searchParams.toString()}`;
        router.push(`/login?return_url=${encodeURIComponent(returnUrl)}`);
        return;
      }

      const response = await fetch('/api/oauth/authorize', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('sonr_auth_token')}`,
        },
        body: JSON.stringify({
          client_id: clientId,
          redirect_uri: redirectUri,
          response_type: responseType,
          scope,
          state,
          code_challenge: codeChallenge,
          code_challenge_method: codeChallengeMethod,
          approved: true,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error_description || 'Authorization failed');
      }

      const result = await response.json();

      const redirectUrl = new URL(redirectUri!);
      if (responseType === 'code') {
        redirectUrl.searchParams.set('code', result.code);
      } else {
        const fragment = new URLSearchParams({
          access_token: result.access_token,
          token_type: 'Bearer',
          expires_in: result.expires_in.toString(),
          scope: scope || '',
        });
        if (state) fragment.set('state', state);
        redirectUrl.hash = fragment.toString();
      }

      if (state) {
        redirectUrl.searchParams.set('state', state);
      }

      window.location.href = redirectUrl.toString();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Authorization failed');
    } finally {
      setIsLoading(false);
    }
  }, [
    isAuthenticated,
    clientId,
    redirectUri,
    responseType,
    scope,
    state,
    codeChallenge,
    codeChallengeMethod,
    searchParams,
    router,
  ]);

  // Handle denial
  const handleDeny = React.useCallback(() => {
    const redirectUrl = new URL(redirectUri!);
    redirectUrl.searchParams.set('error', 'access_denied');
    redirectUrl.searchParams.set('error_description', 'User denied authorization');
    if (state) {
      redirectUrl.searchParams.set('state', state);
    }
    window.location.href = redirectUrl.toString();
  }, [redirectUri, state]);

  // Parse requested scopes with custom icons
  const requestedScopes = React.useMemo(() => {
    if (!scope) return [];

    const scopeDescriptions: Record<
      string,
      { title: string; description: string; icon: React.ReactNode }
    > = {
      openid: {
        title: 'Basic Profile',
        description: 'Your Sonr ID and basic profile information',
        icon: <Shield className="h-4 w-4" />,
      },
      profile: {
        title: 'Profile Information',
        description: 'Your name, picture, and other profile details',
        icon: <Shield className="h-4 w-4" />,
      },
      'vault:read': {
        title: 'Read Vault Data',
        description: 'Read access to your encrypted vault',
        icon: <Shield className="h-4 w-4" />,
      },
      'vault:write': {
        title: 'Write Vault Data',
        description: 'Create and modify data in your vault',
        icon: <Shield className="h-4 w-4" />,
      },
      'vault:sign': {
        title: 'Sign with Vault Keys',
        description: 'Sign transactions and messages with your vault keys',
        icon: <Shield className="h-4 w-4" />,
      },
      'service:manage': {
        title: 'Manage Services',
        description: 'Register and manage services on your behalf',
        icon: <Shield className="h-4 w-4" />,
      },
    };

    return scope.split(' ').map((s) => ({
      scope: s,
      ...(scopeDescriptions[s] || {
        title: s,
        description: `Access to ${s}`,
        icon: <Shield className="h-4 w-4" />,
      }),
    }));
  }, [scope]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-destructive" />
              Authorization Error
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter>
            <Button variant="outline" onClick={() => window.history.back()}>
              Go Back
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div
      className="flex min-h-screen items-center justify-center p-4"
      style={{
        background: clientBranding.theme?.backgroundColor || 'var(--background)',
        fontFamily: clientBranding.theme?.fontFamily || 'inherit',
      }}
    >
      <Card
        className="authorize-card w-full max-w-md"
        style={{
          background: clientBranding.theme?.cardBackground || 'var(--card)',
          borderRadius: clientBranding.theme?.borderRadius || 'var(--radius)',
        }}
      >
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {clientBranding.logo && (
                <img
                  src={clientBranding.logo}
                  alt={clientBranding.name}
                  className="h-12 w-12 rounded"
                  style={{
                    borderRadius: clientBranding.theme?.borderRadius || 'var(--radius)',
                  }}
                />
              )}
              <div>
                <CardTitle
                  style={{
                    color: clientBranding.theme?.primaryColor || 'inherit',
                  }}
                >
                  {clientBranding.name}
                </CardTitle>
                <CardDescription>wants to access your Sonr account</CardDescription>
              </div>
            </div>
            {clientBranding.verified && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <Check
                  className="h-4 w-4"
                  style={{ color: clientBranding.theme?.accentColor || 'rgb(34 197 94)' }}
                />
                Verified
              </div>
            )}
          </div>
        </CardHeader>

        <CardContent>
          <div className="space-y-4">
            {!isAuthenticated && (
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  You need to sign in to continue with authorization
                </AlertDescription>
              </Alert>
            )}

            <div className="space-y-2">
              <p className="text-sm font-medium">This application will be able to:</p>
              <div className="space-y-2">
                {requestedScopes.map(({ scope, title, description, icon }) => (
                  <div
                    key={scope}
                    className="scope-item flex items-start gap-3 p-3 rounded-lg"
                    style={{
                      borderRadius: `calc(${clientBranding.theme?.borderRadius || 'var(--radius)'} * 0.5)`,
                    }}
                  >
                    <div className="mt-0.5">{icon}</div>
                    <div className="flex-1 space-y-1">
                      <p className="text-sm font-medium">{title}</p>
                      <p className="text-xs opacity-90">{description}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div
              className="rounded-lg p-3"
              style={{
                background: clientBranding.theme?.cardBackground
                  ? `color-mix(in srgb, ${clientBranding.theme.cardBackground} 95%, black)`
                  : 'var(--muted)',
                borderRadius: `calc(${clientBranding.theme?.borderRadius || 'var(--radius)'} * 0.5)`,
              }}
            >
              <p className="text-xs text-muted-foreground">
                By authorizing, you allow this application to access your information in accordance
                with its terms of service and privacy policy.
              </p>
            </div>
          </div>
        </CardContent>

        <CardFooter className="flex gap-2">
          <Button variant="outline" onClick={handleDeny} disabled={isLoading} className="flex-1">
            Deny
          </Button>
          <Button
            onClick={handleApprove}
            disabled={isLoading}
            className="flex-1"
            style={{
              background: clientBranding.theme?.primaryColor || 'var(--primary)',
              borderRadius: `calc(${clientBranding.theme?.borderRadius || 'var(--radius)'} * 0.5)`,
            }}
          >
            {isLoading ? 'Authorizing...' : isAuthenticated ? 'Authorize' : 'Sign In & Authorize'}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
