'use client';

import { CheckCircle, Clock, ExternalLink, XCircle } from 'lucide-react';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../ui/table';

export interface Domain {
  id: string;
  domain: string;
  status: 'verified' | 'pending' | 'failed';
  verifiedAt?: string;
  createdAt: string;
}

interface DomainListProps {
  domains: Domain[];
  onVerifyDomain?: (domainId: string) => void;
  onDeleteDomain?: (domainId: string) => void;
}

export function DomainList({ domains, onVerifyDomain, onDeleteDomain }: DomainListProps) {
  const getStatusIcon = (status: Domain['status']) => {
    switch (status) {
      case 'verified':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'pending':
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
    }
  };

  const getStatusBadge = (status: Domain['status']) => {
    const variants = {
      verified: 'default' as const,
      pending: 'secondary' as const,
      failed: 'destructive' as const,
    };

    return (
      <Badge variant={variants[status]} className="capitalize">
        {status}
      </Badge>
    );
  };

  if (domains.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="text-center py-8">
            <p className="text-muted-foreground">No domains registered yet.</p>
            <p className="text-sm text-muted-foreground mt-2">
              Add your first domain to get started with service registration.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Registered Domains</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Domain</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Verified</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {domains.map((domain) => (
              <TableRow key={domain.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(domain.status)}
                    {domain.domain}
                  </div>
                </TableCell>
                <TableCell>{getStatusBadge(domain.status)}</TableCell>
                <TableCell>
                  {domain.verifiedAt ? new Date(domain.verifiedAt).toLocaleDateString() : '-'}
                </TableCell>
                <TableCell>{new Date(domain.createdAt).toLocaleDateString()}</TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    {domain.status === 'failed' && onVerifyDomain && (
                      <Button variant="outline" size="sm" onClick={() => onVerifyDomain(domain.id)}>
                        Retry
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => window.open(`https://${domain.domain}`, '_blank')}
                    >
                      <ExternalLink className="h-4 w-4" />
                    </Button>
                    {onDeleteDomain && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDeleteDomain(domain.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        Delete
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
