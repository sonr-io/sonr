'use client';

import { AlertCircle, ArrowRight, CheckCircle, Copy, XCircle } from 'lucide-react';
import { useState } from 'react';
import { Alert, AlertDescription, AlertTitle } from '../../ui/alert';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Progress } from '../../ui/progress';

export interface VerificationStep {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
}

interface DomainVerificationFlowProps {
  domain: string;
  verificationCode: string;
  steps: VerificationStep[];
  currentStep: number;
  onVerify: () => void;
  onRetry: () => void;
  onCopyRecord: (record: string) => void;
}

export function DomainVerificationFlow({
  domain,
  verificationCode,
  steps,
  currentStep,
  onVerify,
  onRetry,
  onCopyRecord,
}: DomainVerificationFlowProps) {
  const [copiedRecord, setCopiedRecord] = useState<string | null>(null);
  const progress = (currentStep / steps.length) * 100;

  const handleCopyRecord = (record: string) => {
    onCopyRecord(record);
    setCopiedRecord(record);
    setTimeout(() => setCopiedRecord(null), 2000);
  };

  const dnsRecord = `_sonr.${domain} TXT "${verificationCode}"`;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Verify Domain Ownership</CardTitle>
          <CardDescription>Follow the steps below to verify ownership of {domain}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            {/* Progress Bar */}
            <div>
              <div className="flex justify-between text-sm text-muted-foreground mb-2">
                <span>Progress</span>
                <span>{Math.round(progress)}%</span>
              </div>
              <Progress value={progress} className="h-2" />
            </div>

            {/* DNS Record Information */}
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Add TXT Record</AlertTitle>
              <AlertDescription>
                <p className="mb-2">Add the following TXT record to your domain's DNS settings:</p>
                <div className="bg-muted p-3 rounded-md font-mono text-sm relative">
                  <code>{dnsRecord}</code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="absolute right-2 top-2"
                    onClick={() => handleCopyRecord(dnsRecord)}
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                  {copiedRecord === dnsRecord && (
                    <span className="absolute right-12 top-3 text-xs text-green-600">Copied!</span>
                  )}
                </div>
              </AlertDescription>
            </Alert>

            {/* Verification Steps */}
            <div className="space-y-4">
              {steps.map((step) => (
                <div key={step.id} className="flex items-start space-x-3">
                  <div className="flex-shrink-0 mt-0.5">
                    {step.status === 'completed' ? (
                      <CheckCircle className="h-5 w-5 text-green-500" />
                    ) : step.status === 'failed' ? (
                      <XCircle className="h-5 w-5 text-red-500" />
                    ) : step.status === 'in_progress' ? (
                      <div className="h-5 w-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                    ) : (
                      <div className="h-5 w-5 border-2 border-muted rounded-full" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p
                      className={`text-sm font-medium ${
                        step.status === 'completed'
                          ? 'text-green-600'
                          : step.status === 'failed'
                            ? 'text-red-600'
                            : step.status === 'in_progress'
                              ? 'text-primary'
                              : 'text-muted-foreground'
                      }`}
                    >
                      {step.title}
                    </p>
                    <p className="text-sm text-muted-foreground">{step.description}</p>
                  </div>
                </div>
              ))}
            </div>

            {/* Action Buttons */}
            <div className="flex space-x-3">
              <Button
                onClick={onVerify}
                disabled={steps.some((s) => s.status === 'in_progress')}
                className="flex-1"
              >
                Start Verification
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              {steps.some((s) => s.status === 'failed') && (
                <Button onClick={onRetry} variant="outline">
                  Retry
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
