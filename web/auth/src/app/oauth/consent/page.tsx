'use client';

import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@sonr.io/ui';
import { Button } from '@sonr.io/ui';
import { Alert, AlertDescription } from '@sonr.io/ui';
import { Checkbox } from '@sonr.io/ui';
import { Label } from '@sonr.io/ui';
import { AlertCircle, Database, Info, Key, Shield, Wrench } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import React, { Suspense } from 'react';

interface ConsentScope {
  scope: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  required: boolean;
  capabilities?: string[];
}

function ConsentContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isLoading, setIsLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [selectedScopes, setSelectedScopes] = React.useState<Set<string>>(new Set());
  const [rememberChoice, setRememberChoice] = React.useState(false);

  // Extract OAuth parameters
  const clientId = searchParams.get('client_id');
  const redirectUri = searchParams.get('redirect_uri');
  const requestedScopes = searchParams.get('scope')?.split(' ') || [];
  const state = searchParams.get('state');
  const authCode = searchParams.get('auth_code'); // Internal auth code for consent

  // Define available scopes with UCAN capability mappings
  const scopeDefinitions: ConsentScope[] = React.useMemo(
    () => [
      {
        scope: 'openid',
        title: 'Basic Identity',
        description: 'Access to your Sonr DID and basic authentication info',
        icon: <Shield className="h-5 w-5" />,
        required: true,
        capabilities: ['did:read'],
      },
      {
        scope: 'profile',
        title: 'Profile Information',
        description: 'Read your name, picture, and public profile details',
        icon: <Shield className="h-5 w-5" />,
        required: false,
        capabilities: ['profile:read'],
      },
      {
        scope: 'vault:read',
        title: 'Read Vault Data',
        description: 'Read encrypted data stored in your personal vault',
        icon: <Database className="h-5 w-5" />,
        required: false,
        capabilities: ['vault:read', 'vault:list'],
      },
      {
        scope: 'vault:write',
        title: 'Modify Vault Data',
        description: 'Create, update, and organize data in your vault',
        icon: <Database className="h-5 w-5" />,
        required: false,
        capabilities: ['vault:write', 'vault:create', 'vault:update'],
      },
      {
        scope: 'vault:sign',
        title: 'Sign Transactions',
        description: 'Sign messages and transactions using your vault keys',
        icon: <Key className="h-5 w-5" />,
        required: false,
        capabilities: ['vault:sign', 'tx:sign'],
      },
      {
        scope: 'service:manage',
        title: 'Manage Services',
        description: 'Register and manage decentralized services on your behalf',
        icon: <Wrench className="h-5 w-5" />,
        required: false,
        capabilities: ['service:create', 'service:update', 'service:delete'],
      },
    ],
    []
  );

  // Filter to only requested scopes
  const availableScopes = React.useMemo(() => {
    return scopeDefinitions.filter((def) => requestedScopes.includes(def.scope));
  }, [scopeDefinitions, requestedScopes]);

  // Initialize selected scopes with required ones
  React.useEffect(() => {
    const required = new Set(availableScopes.filter((s) => s.required).map((s) => s.scope));
    setSelectedScopes(required);
  }, [availableScopes]);

  // Validate request
  React.useEffect(() => {
    if (!clientId || !redirectUri || !authCode) {
      setError('Invalid consent request. Missing required parameters.');
    }
  }, [clientId, redirectUri, authCode]);

  // Handle scope toggle
  const toggleScope = (scope: string, required: boolean) => {
    if (required) return; // Can't toggle required scopes

    setSelectedScopes((prev) => {
      const next = new Set(prev);
      if (next.has(scope)) {
        next.delete(scope);
      } else {
        next.add(scope);
      }
      return next;
    });
  };

  // Handle consent approval
  const handleApprove = async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Submit consent decision
      const response = await fetch('/api/oauth/consent', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('sonr_auth_token')}`,
        },
        body: JSON.stringify({
          auth_code: authCode,
          client_id: clientId,
          approved_scopes: Array.from(selectedScopes),
          remember: rememberChoice,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error_description || 'Consent submission failed');
      }

      const result = await response.json();

      // Redirect back to authorization endpoint with consent token
      const authUrl = new URL('/oauth/authorize', window.location.origin);
      authUrl.searchParams.set('client_id', clientId!);
      authUrl.searchParams.set('redirect_uri', redirectUri!);
      authUrl.searchParams.set('scope', Array.from(selectedScopes).join(' '));
      authUrl.searchParams.set('consent_token', result.consent_token);
      if (state) {
        authUrl.searchParams.set('state', state);
      }

      router.push(authUrl.pathname + authUrl.search);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Consent submission failed');
    } finally {
      setIsLoading(false);
    }
  };

  // Handle consent denial
  const handleDeny = () => {
    const redirectUrl = new URL(redirectUri!);
    redirectUrl.searchParams.set('error', 'consent_required');
    redirectUrl.searchParams.set('error_description', 'User denied consent for requested scopes');
    if (state) {
      redirectUrl.searchParams.set('state', state);
    }
    window.location.href = redirectUrl.toString();
  };

  // Calculate total capabilities being granted
  const totalCapabilities = React.useMemo(() => {
    const caps = new Set<string>();
    availableScopes.forEach((scope) => {
      if (selectedScopes.has(scope.scope) && scope.capabilities) {
        scope.capabilities.forEach((cap) => caps.add(cap));
      }
    });
    return Array.from(caps);
  }, [availableScopes, selectedScopes]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-destructive" />
              Consent Error
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
    <div className="flex min-h-screen items-center justify-center p-4 bg-gradient-to-br from-background to-muted">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Review Permissions</CardTitle>
          <CardDescription>Choose which permissions to grant this application</CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Scope selection */}
          <div className="space-y-3">
            {availableScopes.map((scope) => (
              <div
                key={scope.scope}
                className={`border rounded-lg p-4 transition-colors ${
                  selectedScopes.has(scope.scope)
                    ? 'bg-primary/5 border-primary/20'
                    : 'bg-muted/30 border-muted-foreground/10'
                }`}
              >
                <div className="flex items-start gap-3">
                  <Checkbox
                    id={scope.scope}
                    checked={selectedScopes.has(scope.scope)}
                    disabled={scope.required}
                    onCheckedChange={() => toggleScope(scope.scope, scope.required)}
                    className="mt-1"
                  />
                  <div className="flex-1 space-y-1">
                    <Label htmlFor={scope.scope} className="flex items-center gap-2 cursor-pointer">
                      {scope.icon}
                      <span className="font-medium">{scope.title}</span>
                      {scope.required && (
                        <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded">
                          Required
                        </span>
                      )}
                    </Label>
                    <p className="text-sm text-muted-foreground">{scope.description}</p>
                    {scope.capabilities && scope.capabilities.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {scope.capabilities.map((cap) => (
                          <span key={cap} className="text-xs bg-muted px-2 py-0.5 rounded">
                            {cap}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* UCAN capabilities summary */}
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              <strong>UCAN Capabilities:</strong> This will create a delegation chain granting{' '}
              {totalCapabilities.length} capabilities to the application.
            </AlertDescription>
          </Alert>

          {/* Remember choice option */}
          <div className="flex items-center gap-2 p-3 rounded-lg bg-muted/50">
            <Checkbox
              id="remember"
              checked={rememberChoice}
              onCheckedChange={(checked) => setRememberChoice(checked as boolean)}
            />
            <Label htmlFor="remember" className="text-sm cursor-pointer">
              Remember my choice for this application
            </Label>
          </div>

          {/* Privacy notice */}
          <div className="text-xs text-muted-foreground p-3 rounded-lg bg-muted/30">
            Your data remains encrypted in your vault. Applications can only access what you
            explicitly permit through these capabilities.
          </div>
        </CardContent>

        <CardFooter className="flex gap-2">
          <Button variant="outline" onClick={handleDeny} disabled={isLoading} className="flex-1">
            Deny
          </Button>
          <Button
            onClick={handleApprove}
            disabled={isLoading || selectedScopes.size === 0}
            className="flex-1"
          >
            {isLoading
              ? 'Processing...'
              : `Grant ${selectedScopes.size} Permission${selectedScopes.size !== 1 ? 's' : ''}`}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

export default function ConsentPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <ConsentContent />
    </Suspense>
  );
}
