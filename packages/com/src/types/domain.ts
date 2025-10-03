/**
 * Domain & DNS Types
 * Types for domain verification and DNS management
 */

import type { ID, Timestamp } from './common';

/**
 * Domain verification status
 */
export type DomainStatus = 'unverified' | 'pending' | 'verified' | 'failed' | 'expired';

/**
 * Domain entity
 */
export interface Domain {
  id: ID;
  domain: string;
  status: DomainStatus;
  owner: string;
  verificationMethod: 'dns' | 'http';
  verificationRecord?: DNSRecord;
  verifiedAt?: Timestamp;
  expiresAt?: Timestamp;
  services?: ID[];
  createdAt: Timestamp;
  updatedAt: Timestamp;
}

/**
 * DNS record for domain verification
 */
export interface DNSRecord {
  type: 'TXT' | 'CNAME' | 'A' | 'AAAA';
  name: string;
  value: string;
  ttl?: number;
}

/**
 * Domain verification request
 */
export interface DomainVerificationRequest {
  domain: string;
  method: 'dns' | 'http';
  autoVerify?: boolean;
}

/**
 * Verification step
 */
export interface VerificationStep {
  id: string;
  label: string;
  description?: string;
  status: 'pending' | 'in-progress' | 'completed' | 'error';
  errorMessage?: string;
}

/**
 * Verification check result
 */
export interface VerificationCheck {
  type: 'dns' | 'http';
  success: boolean;
  message: string;
  details?: Record<string, any>;
  timestamp: Timestamp;
}
