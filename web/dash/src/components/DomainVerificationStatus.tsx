'use client';

import { Alert, AlertDescription, Badge, Button, Card, CardContent, Progress } from '@sonr.io/ui';
import { CheckCircle, Clock, Copy, ExternalLink, RefreshCw, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';

interface DomainVerificationStatusProps {
  domain: string;
  status?: 'verified' | 'pending' | 'failed' | 'unverified';
}

export function DomainVerificationStatus({
  domain,
  status = 'unverified',
}: DomainVerificationStatusProps) {
  const [verificationStatus, setVerificationStatus] = useState(status);
  const [isVerifying, setIsVerifying] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  const txtRecord = `sonr-verify=${domain.replace(/^https?:\/\//, '')}-${Date.now()}`;

  const handleVerify = async () => {
    setIsVerifying(true);
    try {
      // TODO: Implement actual domain verification API call
      await new Promise((resolve) => setTimeout(resolve, 2000));
      setVerificationStatus('verified');
    } catch (_error) {
      setVerificationStatus('failed');
    } finally {
      setIsVerifying(false);
    }
  };

  const handleCopyTxtRecord = () => {
    navigator.clipboard.writeText(txtRecord);
    setCopySuccess(true);
    setTimeout(() => setCopySuccess(false), 2000);
  };

  const getStatusIcon = () => {
    switch (verificationStatus) {
      case 'verified':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'pending':
        return <Clock className="h-5 w-5 text-yellow-500" />;
      case 'failed':
        return <XCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Clock className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusBadge = () => {
    switch (verificationStatus) {
      case 'verified':
        return <Badge variant="success">Verified</Badge>;
      case 'pending':
        return <Badge variant="warning">Pending</Badge>;
      case 'failed':
        return <Badge variant="destructive">Failed</Badge>;
      default:
        return <Badge variant="secondary">Unverified</Badge>;
    }
  };

  return (
    <div className="space-y-4">
      {/* Status Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {getStatusIcon()}
          <div>
            <p className="font-medium">{domain}</p>
            <p className="text-sm text-muted-foreground">Domain ownership verification</p>
          </div>
        </div>
        {getStatusBadge()}
      </div>

      {/* Verification Steps */}
      {verificationStatus !== 'verified' && (
        <Card>
          <CardContent className="pt-6">
            <h4 className="font-medium mb-4">Verification Steps</h4>

            <div className="space-y-4">
              <div className="space-y-2">
                <p className="text-sm font-medium">1. Add TXT Record to DNS</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 p-2 text-xs bg-muted rounded font-mono">{txtRecord}</code>
                  <Button size="sm" variant="outline" onClick={handleCopyTxtRecord}>
                    {copySuccess ? (
                      <CheckCircle className="h-4 w-4" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium">2. Wait for DNS Propagation</p>
                <p className="text-xs text-muted-foreground">
                  This typically takes 5-30 minutes but can take up to 48 hours
                </p>
                <Progress value={33} className="h-2" />
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium">3. Verify Domain Ownership</p>
                <Button onClick={handleVerify} disabled={isVerifying} className="w-full">
                  {isVerifying ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Verifying...
                    </>
                  ) : (
                    'Verify Now'
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Success Message */}
      {verificationStatus === 'verified' && (
        <Alert>
          <CheckCircle className="h-4 w-4" />
          <AlertDescription>
            Domain verification successful! Your service is now linked to {domain}.
          </AlertDescription>
        </Alert>
      )}

      {/* Failed Message */}
      {verificationStatus === 'failed' && (
        <Alert variant="destructive">
          <XCircle className="h-4 w-4" />
          <AlertDescription>
            Domain verification failed. Please check your DNS records and try again.
          </AlertDescription>
        </Alert>
      )}

      {/* Help Link */}
      <div className="flex items-center justify-between text-sm">
        <a
          href="/docs/domain-verification"
          className="text-primary hover:underline flex items-center gap-1"
        >
          Domain verification help
          <ExternalLink className="h-3 w-3" />
        </a>

        {verificationStatus === 'pending' && (
          <Button variant="ghost" size="sm" onClick={handleVerify}>
            <RefreshCw className="mr-2 h-3 w-3" />
            Check Status
          </Button>
        )}
      </div>
    </div>
  );
}
