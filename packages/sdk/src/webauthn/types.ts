/**
 * WebAuthn types for Sonr SDK
 */

export interface WebAuthnRegistrationOptions {
  username: string;
  displayName?: string;
  email?: string;
  tel?: string;
  createVault?: boolean;
  origin?: string;
}

export interface WebAuthnAuthenticationOptions {
  username: string;
  origin?: string;
}

export interface WebAuthnCredential {
  id: string;
  rawId: string;
  type: string;
  publicKey: string;
  counter: number;
  createdAt: string;
}

export interface RegisterStartRequest {
  assertion_value: string;
  assertion_type: 'email' | 'tel' | 'username';
  service_origin: string;
}

export interface RegisterStartResponse {
  challenge: string;
  rp: {
    id: string;
    name: string;
  };
  user: {
    id: string;
    name: string;
    displayName: string;
  };
  pubKeyCredParams: Array<{
    type: string;
    alg: number;
  }>;
  timeout?: number;
  attestation?: AttestationConveyancePreference;
  authenticatorSelection?: AuthenticatorSelectionCriteria;
}

export interface AuthenticatorSelectionCriteria {
  authenticatorAttachment?: AuthenticatorAttachment;
  requireResidentKey?: boolean;
  residentKey?: ResidentKeyRequirement;
  userVerification?: UserVerificationRequirement;
}

export type AuthenticatorAttachment = 'platform' | 'cross-platform';
export type ResidentKeyRequirement = 'discouraged' | 'preferred' | 'required';
export type UserVerificationRequirement = 'required' | 'preferred' | 'discouraged';
export type AttestationConveyancePreference = 'none' | 'indirect' | 'direct' | 'enterprise';

export interface WebAuthnRegistrationResult {
  success: boolean;
  did?: string;
  vaultId?: string;
  credential?: WebAuthnCredential;
  ucanToken?: string;
  error?: string;
}

export interface WebAuthnAuthenticationResult {
  success: boolean;
  did?: string;
  vaultId?: string;
  sessionToken?: string;
  error?: string;
}

export interface DIDDocument {
  id: string;
  controller: string[];
  verificationMethod?: VerificationMethod[];
  authentication?: Array<string | VerificationMethod>;
  assertionMethod?: Array<string | VerificationMethod>;
  keyAgreement?: Array<string | VerificationMethod>;
  capabilityInvocation?: Array<string | VerificationMethod>;
  capabilityDelegation?: Array<string | VerificationMethod>;
  service?: ServiceEndpoint[];
}

export interface VerificationMethod {
  id: string;
  type: string;
  controller: string;
  publicKeyBase58?: string;
  publicKeyBase64?: string;
  publicKeyJwk?: any; // JsonWebKey interface from Web Crypto API
  publicKeyMultibase?: string;
}

export interface ServiceEndpoint {
  id: string;
  type: string | string[];
  serviceEndpoint: string | string[] | Record<string, any>;
}

export interface DIDDocumentMetadata {
  created?: string;
  updated?: string;
  deactivated?: boolean;
  versionId?: string;
  nextUpdate?: string;
  nextVersionId?: string;
  equivalentId?: string[];
  canonicalId?: string;
  ucanDelegationChain?: string;
}