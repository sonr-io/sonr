'use client';

import { CheckCircle, Clock, XCircle } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../ui/select';
import type { Domain } from './DomainList';

interface DomainSelectorProps {
  domains: Domain[];
  value?: string;
  onValueChange?: (value: string) => void;
  placeholder?: string;
  showStatus?: boolean;
}

export function DomainSelector({
  domains,
  value,
  onValueChange,
  placeholder = 'Select a domain',
  showStatus = true,
}: DomainSelectorProps) {
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

  const verifiedDomains = domains.filter((domain) => domain.status === 'verified');

  if (domains.length === 0) {
    return (
      <Select disabled>
        <SelectTrigger>
          <SelectValue placeholder="No domains available" />
        </SelectTrigger>
      </Select>
    );
  }

  return (
    <Select value={value} onValueChange={onValueChange}>
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {verifiedDomains.length > 0 &&
          verifiedDomains.map((domain) => (
            <SelectItem key={domain.id} value={domain.domain}>
              <div className="flex items-center gap-2">
                {showStatus && getStatusIcon(domain.status)}
                <span>{domain.domain}</span>
              </div>
            </SelectItem>
          ))}

        {domains.filter((d) => d.status !== 'verified').length > 0 &&
          domains
            .filter((domain) => domain.status !== 'verified')
            .map((domain) => (
              <SelectItem
                key={domain.id}
                value={domain.domain}
                disabled={domain.status !== 'verified'}
              >
                <div className="flex items-center gap-2">
                  {showStatus && getStatusIcon(domain.status)}
                  <span className="text-muted-foreground">{domain.domain}</span>
                  <span className="text-xs text-muted-foreground">({domain.status})</span>
                </div>
              </SelectItem>
            ))}
      </SelectContent>
    </Select>
  );
}
