'use client';

import { CheckCircle, Copy } from 'lucide-react';
import { useState } from 'react';
import { Alert, AlertDescription } from '../../ui/alert';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';

interface DNSInstructionsProps {
  domain: string;
  verificationCode: string;
  onCopyCode?: () => void;
}

export function DNSInstructions({ domain, verificationCode, onCopyCode }: DNSInstructionsProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(verificationCode);
    setCopied(true);
    onCopyCode?.();
    setTimeout(() => setCopied(false), 2000);
  };

  const txtRecord = `sonr-domain-verification=${verificationCode}`;

  return (
    <Card>
      <CardHeader>
        <CardTitle>DNS Verification Instructions</CardTitle>
        <CardDescription>
          Add the following TXT record to your DNS settings to verify domain ownership.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Domain</label>
          <div className="p-3 bg-muted rounded-md font-mono text-sm">{domain}</div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Record Type</label>
          <div className="p-3 bg-muted rounded-md font-mono text-sm">TXT</div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Record Name/Host</label>
          <div className="p-3 bg-muted rounded-md font-mono text-sm">
            @ (or leave empty for root domain)
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Record Value</label>
          <div className="flex items-center gap-2">
            <div className="flex-1 p-3 bg-muted rounded-md font-mono text-sm break-all">
              {txtRecord}
            </div>
            <Button variant="outline" size="sm" onClick={handleCopy} className="shrink-0">
              {copied ? (
                <CheckCircle className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>

        <Alert>
          <AlertDescription>
            <strong>Note:</strong> DNS propagation can take up to 24 hours, but typically takes 5-15
            minutes. Once you've added the TXT record, click "Verify Domain" to complete the
            verification process.
          </AlertDescription>
        </Alert>

        <div className="pt-4 space-y-2">
          <h4 className="font-medium">Common DNS Providers</h4>
          <ul className="text-sm text-muted-foreground space-y-1">
            <li>
              • <strong>Cloudflare:</strong> DNS → Records → Add record
            </li>
            <li>
              • <strong>GoDaddy:</strong> DNS Management → TXT records
            </li>
            <li>
              • <strong>Namecheap:</strong> Domain List → Manage → Advanced DNS
            </li>
            <li>
              • <strong>Google Domains:</strong> DNS → Custom records
            </li>
          </ul>
        </div>
      </CardContent>
    </Card>
  );
}
