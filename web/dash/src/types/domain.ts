import { z } from 'zod';

/**
 * Domain verification method
 */
export enum DomainVerificationMethod {
  DNS_TXT = 'dns_txt',
  DNS_CNAME = 'dns_cname',
  HTTP_FILE = 'http_file',
  META_TAG = 'meta_tag',
}

/**
 * Domain verification status
 */
export enum DomainVerificationStatus {
  UNVERIFIED = 'unverified',
  PENDING = 'pending',
  VERIFIED = 'verified',
  FAILED = 'failed',
  EXPIRED = 'expired',
}

/**
 * DNS record type for verification
 */
export enum DnsRecordType {
  TXT = 'TXT',
  CNAME = 'CNAME',
  A = 'A',
  AAAA = 'AAAA',
}

/**
 * Domain verification challenge
 */
export interface DomainChallenge {
  id: string;
  domain: string;
  method: DomainVerificationMethod;
  token: string;
  expiresAt: string;
  createdAt: string;
  attempts: number;
  maxAttempts: number;
}

/**
 * DNS record for verification
 */
export interface DnsRecord {
  type: DnsRecordType;
  name: string;
  value: string;
  ttl?: number;
  priority?: number;
}

/**
 * Domain verification instructions
 */
export interface DomainVerificationInstructions {
  method: DomainVerificationMethod;
  dnsRecords?: DnsRecord[];
  httpFilePath?: string;
  httpFileContent?: string;
  metaTagName?: string;
  metaTagContent?: string;
  instructions: string[];
  estimatedTime?: string;
}

/**
 * Domain verification attempt
 */
export interface DomainVerificationAttempt {
  id: string;
  domain: string;
  timestamp: string;
  success: boolean;
  method: DomainVerificationMethod;
  recordsFound?: DnsRecord[];
  expectedRecords?: DnsRecord[];
  error?: string;
  responseTime?: number;
}

/**
 * Domain ownership details
 */
export interface DomainOwnership {
  domain: string;
  owner: string;
  verifiedAt?: string;
  expiresAt?: string;
  autoRenew: boolean;
  delegatedTo?: string[];
}

/**
 * Domain configuration
 */
export interface DomainConfig {
  domain: string;
  subdomains: string[];
  wildcardEnabled: boolean;
  sslEnabled: boolean;
  cors?: {
    enabled: boolean;
    origins: string[];
    methods: string[];
    headers: string[];
  };
  rateLimit?: {
    enabled: boolean;
    requestsPerMinute: number;
    requestsPerHour: number;
  };
  redirects?: Array<{
    from: string;
    to: string;
    statusCode: number;
  }>;
}

/**
 * Domain verification state
 */
export interface DomainVerification {
  domain: string;
  status: DomainVerificationStatus;
  method?: DomainVerificationMethod;
  challenge?: DomainChallenge;
  instructions?: DomainVerificationInstructions;
  attempts?: DomainVerificationAttempt[];
  ownership?: DomainOwnership;
  config?: DomainConfig;
  createdAt: string;
  updatedAt: string;
  verifiedAt?: string;
  lastChecked?: string;
  nextCheckAt?: string;
}

/**
 * Domain health status
 */
export interface DomainHealth {
  domain: string;
  status: 'healthy' | 'degraded' | 'offline';
  sslValid: boolean;
  sslExpiresAt?: string;
  dnsResolvable: boolean;
  httpReachable: boolean;
  averageResponseTime?: number;
  lastChecked: string;
  issues?: string[];
}

/**
 * Domain analytics
 */
export interface DomainAnalytics {
  domain: string;
  timeRange: {
    start: string;
    end: string;
  };
  metrics: {
    totalRequests: number;
    uniqueVisitors: number;
    bandwidth: number;
    cacheHitRate: number;
    errorRate: number;
  };
  topPaths?: Array<{
    path: string;
    requests: number;
  }>;
  topReferers?: Array<{
    referer: string;
    requests: number;
  }>;
  geographic?: Array<{
    country: string;
    requests: number;
  }>;
}

/**
 * Validation schemas
 */
export const DomainChallengeSchema = z.object({
  id: z.string(),
  domain: z.string(),
  method: z.nativeEnum(DomainVerificationMethod),
  token: z.string().min(32),
  expiresAt: z.string(),
  createdAt: z.string(),
  attempts: z.number().min(0),
  maxAttempts: z.number().min(1).max(100),
});

export const DnsRecordSchema = z.object({
  type: z.nativeEnum(DnsRecordType),
  name: z.string(),
  value: z.string(),
  ttl: z.number().min(60).max(86400).optional(),
  priority: z.number().min(0).max(65535).optional(),
});

export const DomainVerificationInstructionsSchema = z.object({
  method: z.nativeEnum(DomainVerificationMethod),
  dnsRecords: z.array(DnsRecordSchema).optional(),
  httpFilePath: z.string().optional(),
  httpFileContent: z.string().optional(),
  metaTagName: z.string().optional(),
  metaTagContent: z.string().optional(),
  instructions: z.array(z.string()),
  estimatedTime: z.string().optional(),
});

export const DomainOwnershipSchema = z.object({
  domain: z.string(),
  owner: z.string(),
  verifiedAt: z.string().optional(),
  expiresAt: z.string().optional(),
  autoRenew: z.boolean(),
  delegatedTo: z.array(z.string()).optional(),
});

export const DomainConfigSchema = z.object({
  domain: z.string(),
  subdomains: z.array(z.string()),
  wildcardEnabled: z.boolean(),
  sslEnabled: z.boolean(),
  cors: z
    .object({
      enabled: z.boolean(),
      origins: z.array(z.string()),
      methods: z.array(z.string()),
      headers: z.array(z.string()),
    })
    .optional(),
  rateLimit: z
    .object({
      enabled: z.boolean(),
      requestsPerMinute: z.number().min(1),
      requestsPerHour: z.number().min(1),
    })
    .optional(),
  redirects: z
    .array(
      z.object({
        from: z.string(),
        to: z.string(),
        statusCode: z.number().min(300).max(399),
      })
    )
    .optional(),
});

export const DomainVerificationSchema = z.object({
  domain: z.string(),
  status: z.nativeEnum(DomainVerificationStatus),
  method: z.nativeEnum(DomainVerificationMethod).optional(),
  challenge: DomainChallengeSchema.optional(),
  instructions: DomainVerificationInstructionsSchema.optional(),
  attempts: z.array(z.any()).optional(), // Simplified for brevity
  ownership: DomainOwnershipSchema.optional(),
  config: DomainConfigSchema.optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
  verifiedAt: z.string().optional(),
  lastChecked: z.string().optional(),
  nextCheckAt: z.string().optional(),
});

/**
 * Type guards
 */
export function isDomainVerified(verification: DomainVerification): boolean {
  return verification.status === DomainVerificationStatus.VERIFIED;
}

export function isDomainPending(verification: DomainVerification): boolean {
  return verification.status === DomainVerificationStatus.PENDING;
}

export function isDomainExpired(verification: DomainVerification): boolean {
  if (verification.status === DomainVerificationStatus.EXPIRED) return true;
  if (verification.ownership?.expiresAt) {
    return new Date(verification.ownership.expiresAt) < new Date();
  }
  return false;
}

export function hasDnsMethod(verification: DomainVerification): boolean {
  return (
    verification.method === DomainVerificationMethod.DNS_TXT ||
    verification.method === DomainVerificationMethod.DNS_CNAME
  );
}

/**
 * Utility functions
 */
export function formatDomainStatus(status: DomainVerificationStatus): string {
  const statusMap: Record<DomainVerificationStatus, string> = {
    [DomainVerificationStatus.UNVERIFIED]: 'Unverified',
    [DomainVerificationStatus.PENDING]: 'Pending Verification',
    [DomainVerificationStatus.VERIFIED]: 'Verified',
    [DomainVerificationStatus.FAILED]: 'Verification Failed',
    [DomainVerificationStatus.EXPIRED]: 'Expired',
  };
  return statusMap[status] || status;
}

export function getDomainStatusColor(status: DomainVerificationStatus): string {
  const colorMap: Record<DomainVerificationStatus, string> = {
    [DomainVerificationStatus.VERIFIED]: 'green',
    [DomainVerificationStatus.PENDING]: 'yellow',
    [DomainVerificationStatus.FAILED]: 'red',
    [DomainVerificationStatus.EXPIRED]: 'orange',
    [DomainVerificationStatus.UNVERIFIED]: 'gray',
  };
  return colorMap[status] || 'gray';
}

export function formatDnsRecord(record: DnsRecord): string {
  return `${record.type} ${record.name} ${record.value}${record.ttl ? ` TTL:${record.ttl}` : ''}`;
}

export function generateTxtRecordValue(token: string, prefix = 'sonr-verification'): string {
  return `${prefix}=${token}`;
}

export function parseTxtRecordValue(value: string): { prefix: string; token: string } | null {
  const match = value.match(/^([^=]+)=(.+)$/);
  if (!match) return null;
  return { prefix: match[1], token: match[2] };
}

export function estimateVerificationTime(method: DomainVerificationMethod): string {
  const estimates: Record<DomainVerificationMethod, string> = {
    [DomainVerificationMethod.DNS_TXT]: '5-60 minutes (DNS propagation)',
    [DomainVerificationMethod.DNS_CNAME]: '5-60 minutes (DNS propagation)',
    [DomainVerificationMethod.HTTP_FILE]: '1-2 minutes',
    [DomainVerificationMethod.META_TAG]: '1-2 minutes',
  };
  return estimates[method] || 'Unknown';
}

export function getVerificationMethodName(method: DomainVerificationMethod): string {
  const names: Record<DomainVerificationMethod, string> = {
    [DomainVerificationMethod.DNS_TXT]: 'DNS TXT Record',
    [DomainVerificationMethod.DNS_CNAME]: 'DNS CNAME Record',
    [DomainVerificationMethod.HTTP_FILE]: 'HTTP File Upload',
    [DomainVerificationMethod.META_TAG]: 'HTML Meta Tag',
  };
  return names[method] || method;
}

/**
 * Domain validation helpers
 */
export function isValidDomain(domain: string): boolean {
  const domainRegex = /^[a-z0-9]+([-.]{1}[a-z0-9]+)*\.[a-z]{2,}$/i;
  return domainRegex.test(domain);
}

export function isSubdomain(domain: string): boolean {
  const parts = domain.split('.');
  return parts.length > 2;
}

export function getBaseDomain(domain: string): string {
  const parts = domain.split('.');
  if (parts.length <= 2) return domain;
  return parts.slice(-2).join('.');
}

export function getSubdomainPrefix(domain: string): string | null {
  const parts = domain.split('.');
  if (parts.length <= 2) return null;
  return parts.slice(0, -2).join('.');
}

export function normalizeDomain(domain: string): string {
  return domain
    .toLowerCase()
    .replace(/^https?:\/\//, '')
    .replace(/\/$/, '');
}
