'use client';

import {
  AlertCircle,
  CheckCircle,
  ChevronDown,
  ChevronRight,
  Clock,
  Copy,
  ExternalLink,
  Key,
  Shield,
} from 'lucide-react';
import { useState } from 'react';
import { cn } from '../../../lib/utils';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../ui/tabs';

/**
 * UCAN capability structure
 */
export interface UCANCapability {
  with: string; // Resource identifier
  can: string; // Action/permission
  nb?: Record<string, any>; // Additional constraints
}

/**
 * UCAN token structure
 */
export interface UCANToken {
  iss: string; // Issuer DID
  aud: string; // Audience DID
  exp?: number; // Expiration timestamp
  nbf?: number; // Not before timestamp
  att: UCANCapability[]; // Attenuations/capabilities
  prf?: string[]; // Proof chain
  fct?: Record<string, any>; // Facts
}

/**
 * Props for UCANViewer component
 */
export interface UCANViewerProps {
  token: UCANToken;
  showRaw?: boolean;
  showProofChain?: boolean;
  onCopyToken?: () => void;
  onVerify?: () => Promise<boolean>;
  className?: string;
}

/**
 * Viewer for UCAN token visualization and capability display
 */
export function UCANViewer({
  token,
  showRaw = true,
  showProofChain = true,
  onCopyToken,
  onVerify,
  className,
}: UCANViewerProps) {
  const [expandedCapabilities, setExpandedCapabilities] = useState<Set<number>>(new Set());
  const [verificationStatus, setVerificationStatus] = useState<
    'idle' | 'verifying' | 'valid' | 'invalid'
  >('idle');

  const toggleCapability = (index: number) => {
    const newExpanded = new Set(expandedCapabilities);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedCapabilities(newExpanded);
  };

  const handleVerify = async () => {
    if (!onVerify) return;
    setVerificationStatus('verifying');
    try {
      const isValid = await onVerify();
      setVerificationStatus(isValid ? 'valid' : 'invalid');
    } catch {
      setVerificationStatus('invalid');
    }
  };

  const formatDate = (timestamp?: number) => {
    if (!timestamp) return 'Never';
    return new Date(timestamp * 1000).toLocaleString();
  };

  const isExpired = token.exp && token.exp * 1000 < Date.now();
  const isNotYetValid = token.nbf && token.nbf * 1000 > Date.now();

  return (
    <div className={cn('space-y-4', className)}>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              <CardTitle>UCAN Token</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              {isExpired && (
                <Badge variant="destructive" className="text-xs">
                  <AlertCircle className="h-3 w-3 mr-1" />
                  Expired
                </Badge>
              )}
              {isNotYetValid && (
                <Badge variant="secondary" className="text-xs">
                  <Clock className="h-3 w-3 mr-1" />
                  Not Yet Valid
                </Badge>
              )}
              {verificationStatus === 'valid' && (
                <Badge variant="default" className="text-xs">
                  <CheckCircle className="h-3 w-3 mr-1" />
                  Verified
                </Badge>
              )}
              {verificationStatus === 'invalid' && (
                <Badge variant="destructive" className="text-xs">
                  <AlertCircle className="h-3 w-3 mr-1" />
                  Invalid
                </Badge>
              )}
            </div>
          </div>
          <CardDescription>
            User-Controlled Authorization Network token with delegated capabilities
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="details" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="details">Details</TabsTrigger>
              <TabsTrigger value="capabilities">Capabilities ({token.att.length})</TabsTrigger>
              {showRaw && <TabsTrigger value="raw">Raw Token</TabsTrigger>}
            </TabsList>

            <TabsContent value="details" className="space-y-4">
              <div className="space-y-3">
                <div className="flex items-start gap-2">
                  <Key className="h-4 w-4 mt-0.5 text-muted-foreground" />
                  <div className="flex-1">
                    <p className="text-sm font-medium">Issuer</p>
                    <code className="text-xs text-muted-foreground break-all">{token.iss}</code>
                  </div>
                </div>

                <div className="flex items-start gap-2">
                  <Key className="h-4 w-4 mt-0.5 text-muted-foreground" />
                  <div className="flex-1">
                    <p className="text-sm font-medium">Audience</p>
                    <code className="text-xs text-muted-foreground break-all">{token.aud}</code>
                  </div>
                </div>

                {token.exp && (
                  <div className="flex items-start gap-2">
                    <Clock className="h-4 w-4 mt-0.5 text-muted-foreground" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">Expires</p>
                      <p className="text-xs text-muted-foreground">{formatDate(token.exp)}</p>
                    </div>
                  </div>
                )}

                {token.nbf && (
                  <div className="flex items-start gap-2">
                    <Clock className="h-4 w-4 mt-0.5 text-muted-foreground" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">Not Before</p>
                      <p className="text-xs text-muted-foreground">{formatDate(token.nbf)}</p>
                    </div>
                  </div>
                )}

                {showProofChain && token.prf && token.prf.length > 0 && (
                  <div className="flex items-start gap-2">
                    <ExternalLink className="h-4 w-4 mt-0.5 text-muted-foreground" />
                    <div className="flex-1">
                      <p className="text-sm font-medium">Proof Chain</p>
                      <div className="space-y-1 mt-1">
                        {token.prf.map((proof, index) => (
                          <code
                            key={index}
                            className="block text-xs text-muted-foreground break-all"
                          >
                            {proof}
                          </code>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="flex gap-2">
                {onVerify && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={handleVerify}
                    disabled={verificationStatus === 'verifying'}
                  >
                    {verificationStatus === 'verifying' ? 'Verifying...' : 'Verify Token'}
                  </Button>
                )}
                {onCopyToken && (
                  <Button size="sm" variant="outline" onClick={onCopyToken}>
                    <Copy className="h-3 w-3 mr-1" />
                    Copy Token
                  </Button>
                )}
              </div>
            </TabsContent>

            <TabsContent value="capabilities" className="space-y-3">
              {token.att.map((capability, index) => (
                <Card key={index} className="border-muted">
                  <CardHeader
                    className="cursor-pointer py-3"
                    onClick={() => toggleCapability(index)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {expandedCapabilities.has(index) ? (
                          <ChevronDown className="h-4 w-4" />
                        ) : (
                          <ChevronRight className="h-4 w-4" />
                        )}
                        <Badge variant="outline">{capability.can}</Badge>
                        <code className="text-xs text-muted-foreground">{capability.with}</code>
                      </div>
                    </div>
                  </CardHeader>
                  {expandedCapabilities.has(index) && capability.nb && (
                    <CardContent className="pt-0">
                      <div className="space-y-2">
                        <p className="text-xs font-medium text-muted-foreground">Constraints:</p>
                        <pre className="text-xs bg-muted p-2 rounded overflow-auto">
                          {JSON.stringify(capability.nb, null, 2)}
                        </pre>
                      </div>
                    </CardContent>
                  )}
                </Card>
              ))}
              {token.att.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">
                  No capabilities defined
                </div>
              )}
            </TabsContent>

            {showRaw && (
              <TabsContent value="raw" className="space-y-3">
                <pre className="text-xs bg-muted p-4 rounded overflow-auto">
                  {JSON.stringify(token, null, 2)}
                </pre>
              </TabsContent>
            )}
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
