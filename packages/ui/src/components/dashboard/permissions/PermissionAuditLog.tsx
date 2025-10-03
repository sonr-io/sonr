'use client';

import {
  AlertTriangle,
  CheckCircle,
  Clock,
  Download,
  RefreshCw,
  Search,
  Shield,
  User,
  XCircle,
} from 'lucide-react';
import { useState } from 'react';
import { cn } from '../../../lib/utils';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Input } from '../../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../ui/table';

/**
 * Audit log entry structure
 */
export interface AuditLogEntry {
  id: string;
  timestamp: Date;
  actor: {
    id: string;
    name: string;
    type: 'user' | 'service' | 'system';
  };
  action: 'grant' | 'revoke' | 'request' | 'deny' | 'expire' | 'attest';
  resource: string;
  permission: string;
  status: 'success' | 'failed' | 'pending';
  reason?: string;
  metadata?: Record<string, any>;
}

/**
 * Props for PermissionAuditLog component
 */
export interface PermissionAuditLogProps {
  entries: AuditLogEntry[];
  loading?: boolean;
  onRefresh?: () => void;
  onExport?: () => void;
  showFilters?: boolean;
  className?: string;
}

/**
 * Table display for permission audit trail
 */
export function PermissionAuditLog({
  entries,
  loading = false,
  onRefresh,
  onExport,
  showFilters = true,
  className,
}: PermissionAuditLogProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterAction, setFilterAction] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  const filteredEntries = entries.filter((entry) => {
    const matchesSearch =
      searchQuery === '' ||
      entry.actor.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.resource.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.permission.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesAction = filterAction === 'all' || entry.action === filterAction;
    const matchesStatus = filterStatus === 'all' || entry.status === filterStatus;

    return matchesSearch && matchesAction && matchesStatus;
  });

  const getActionIcon = (action: AuditLogEntry['action']) => {
    switch (action) {
      case 'grant':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'revoke':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'request':
        return <Clock className="h-4 w-4 text-blue-500" />;
      case 'deny':
        return <XCircle className="h-4 w-4 text-orange-500" />;
      case 'expire':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case 'attest':
        return <Shield className="h-4 w-4 text-purple-500" />;
      default:
        return null;
    }
  };

  const getStatusBadge = (status: AuditLogEntry['status']) => {
    switch (status) {
      case 'success':
        return (
          <Badge variant="default" className="text-xs">
            Success
          </Badge>
        );
      case 'failed':
        return (
          <Badge variant="destructive" className="text-xs">
            Failed
          </Badge>
        );
      case 'pending':
        return (
          <Badge variant="secondary" className="text-xs">
            Pending
          </Badge>
        );
      default:
        return null;
    }
  };

  const getActorIcon = (type: AuditLogEntry['actor']['type']) => {
    switch (type) {
      case 'user':
        return <User className="h-3 w-3" />;
      case 'service':
        return <Shield className="h-3 w-3" />;
      case 'system':
        return <Clock className="h-3 w-3" />;
      default:
        return null;
    }
  };

  return (
    <Card className={cn('', className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Permission Audit Log</CardTitle>
            <CardDescription>Track all permission changes and access requests</CardDescription>
          </div>
          <div className="flex gap-2">
            {onRefresh && (
              <Button size="sm" variant="outline" onClick={onRefresh} disabled={loading}>
                <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              </Button>
            )}
            {onExport && (
              <Button size="sm" variant="outline" onClick={onExport}>
                <Download className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {showFilters && (
          <div className="flex flex-col sm:flex-row gap-4 mb-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search by actor, resource, or permission..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="flex gap-2">
              <Select value={filterAction} onValueChange={setFilterAction}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Actions</SelectItem>
                  <SelectItem value="grant">Grant</SelectItem>
                  <SelectItem value="revoke">Revoke</SelectItem>
                  <SelectItem value="request">Request</SelectItem>
                  <SelectItem value="deny">Deny</SelectItem>
                  <SelectItem value="expire">Expire</SelectItem>
                  <SelectItem value="attest">Attest</SelectItem>
                </SelectContent>
              </Select>

              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="success">Success</SelectItem>
                  <SelectItem value="failed">Failed</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        )}

        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-40">Timestamp</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Permission</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Reason</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8">
                    <div className="flex items-center justify-center gap-2 text-muted-foreground">
                      <RefreshCw className="h-4 w-4 animate-spin" />
                      Loading audit logs...
                    </div>
                  </TableCell>
                </TableRow>
              ) : filteredEntries.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    No audit log entries found
                  </TableCell>
                </TableRow>
              ) : (
                filteredEntries.map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell className="font-mono text-xs">
                      {entry.timestamp.toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getActorIcon(entry.actor.type)}
                        <div>
                          <p className="text-sm font-medium">{entry.actor.name}</p>
                          <p className="text-xs text-muted-foreground">{entry.actor.type}</p>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getActionIcon(entry.action)}
                        <span className="text-sm capitalize">{entry.action}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{entry.resource}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {entry.permission}
                      </Badge>
                    </TableCell>
                    <TableCell>{getStatusBadge(entry.status)}</TableCell>
                    <TableCell className="text-right text-xs text-muted-foreground">
                      {entry.reason || '-'}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
