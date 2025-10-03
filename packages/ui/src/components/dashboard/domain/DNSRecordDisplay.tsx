'use client';

import { Copy, ExternalLink, RefreshCw } from 'lucide-react';
import { useState } from 'react';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../ui/table';

export interface DNSRecord {
  type: 'TXT' | 'A' | 'AAAA' | 'CNAME' | 'MX';
  name: string;
  value: string;
  ttl: number;
  priority?: number;
  verified: boolean;
}

interface DNSRecordDisplayProps {
  domain: string;
  records: DNSRecord[];
  onRefresh: () => void;
  onCopy: (record: DNSRecord) => void;
  isRefreshing?: boolean;
  providerInstructions?: {
    name: string;
    url: string;
  };
}

export function DNSRecordDisplay({
  domain,
  records,
  onRefresh,
  onCopy,
  isRefreshing = false,
  providerInstructions,
}: DNSRecordDisplayProps) {
  const [copiedRecord, setCopiedRecord] = useState<string | null>(null);

  const handleCopy = (record: DNSRecord) => {
    const recordString = `${record.name} ${record.type} ${record.value}`;
    onCopy(record);
    setCopiedRecord(recordString);
    setTimeout(() => setCopiedRecord(null), 2000);
  };

  const formatTTL = (ttl: number) => {
    if (ttl < 60) return `${ttl}s`;
    if (ttl < 3600) return `${Math.floor(ttl / 60)}m`;
    return `${Math.floor(ttl / 3600)}h`;
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>DNS Records</CardTitle>
            <CardDescription>
              Configure these records in your DNS provider for {domain}
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={onRefresh} disabled={isRefreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {providerInstructions && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <div className="flex items-center justify-between">
                <p className="text-sm text-blue-900">
                  Need help? View instructions for {providerInstructions.name}
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => window.open(providerInstructions.url, '_blank')}
                >
                  <ExternalLink className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}

          <div className="border rounded-lg overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead>TTL</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-[100px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {records.map((record, index) => (
                  <TableRow key={index}>
                    <TableCell>
                      <Badge variant="outline">{record.type}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{record.name}</TableCell>
                    <TableCell className="font-mono text-sm max-w-[300px] truncate">
                      <div className="flex items-center space-x-2">
                        <span className="truncate">{record.value}</span>
                        {record.priority && (
                          <Badge variant="secondary" className="ml-2">
                            Priority: {record.priority}
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{formatTTL(record.ttl)}</TableCell>
                    <TableCell>
                      {record.verified ? (
                        <Badge variant="default" className="bg-green-500">
                          Verified
                        </Badge>
                      ) : (
                        <Badge variant="secondary">Pending</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <Button variant="ghost" size="sm" onClick={() => handleCopy(record)}>
                        <Copy className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {copiedRecord && (
            <div className="fixed bottom-4 right-4 bg-green-600 text-white px-4 py-2 rounded-lg shadow-lg">
              Record copied to clipboard!
            </div>
          )}

          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="text-sm font-medium mb-2">Quick Setup Guide</h4>
            <ol className="text-sm text-muted-foreground space-y-1">
              <li>1. Log in to your DNS provider's control panel</li>
              <li>2. Navigate to DNS management for {domain}</li>
              <li>3. Add each record shown above</li>
              <li>4. Save changes and wait for propagation (up to 48 hours)</li>
              <li>5. Click "Refresh" to check verification status</li>
            </ol>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
