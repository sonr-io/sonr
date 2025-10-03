/**
 * Self-Issued OpenID Provider (SIOP) v2 Utilities
 *
 * This module provides comprehensive SIOP v2 utilities for decentralized identity
 * authentication using DIDs (Decentralized Identifiers) and Verifiable Presentations.
 *
 * Features:
 * - SIOP request parsing and validation
 * - Self-issued ID token generation
 * - DID-based authentication
 * - Verifiable Presentation (VP) token utilities
 * - SIOP response builder
 * - Dynamic client registration
 * - DID document resolution
 */

import { type JWTHeader, type JWTPayload, decodeJWT, generateRandomString } from './oidc';

// ============================================================================
// Types and Interfaces
// ============================================================================

/**
 * SIOP Request Parameters
 */
export interface SIOPRequest {
  response_type: 'id_token';
  client_id: string;
  redirect_uri: string;
  scope: string;
  state?: string;
  nonce?: string;
  response_mode?: 'form_post' | 'fragment' | 'query';
  claims?: string | SIOPClaims;
  registration?: string | SIOPClientMetadata;
  request?: string; // JWT request object
  request_uri?: string;
  // SIOP specific parameters
  subject_syntax_types_supported?: string[];
  id_token_types_supported?: string[];
}

/**
 * SIOP Claims Request
 */
export interface SIOPClaims {
  id_token?: {
    [key: string]: {
      essential?: boolean;
      value?: string;
      values?: string[];
    } | null;
  };
  userinfo?: {
    [key: string]: {
      essential?: boolean;
      value?: string;
      values?: string[];
    } | null;
  };
}

/**
 * SIOP Client Metadata for Dynamic Registration
 */
export interface SIOPClientMetadata {
  application_type?: string;
  client_name?: string;
  client_uri?: string;
  logo_uri?: string;
  contacts?: string[];
  tos_uri?: string;
  policy_uri?: string;
  redirect_uris?: string[];
  response_types?: string[];
  grant_types?: string[];
  subject_type?: 'public' | 'pairwise';
  id_token_signed_response_alg?: string;
  id_token_encrypted_response_alg?: string;
  id_token_encrypted_response_enc?: string;
  userinfo_signed_response_alg?: string;
  userinfo_encrypted_response_alg?: string;
  userinfo_encrypted_response_enc?: string;
  request_object_signing_alg?: string;
  request_object_encryption_alg?: string;
  request_object_encryption_enc?: string;
  // SIOP specific metadata
  subject_syntax_types_supported?: string[];
  id_token_types_supported?: string[];
  vp_formats?: VPFormats;
}

/**
 * Verifiable Presentation Formats
 */
export interface VPFormats {
  jwt_vp?: {
    alg?: string[];
  };
  jwt_vc?: {
    alg?: string[];
  };
  ldp_vp?: {
    proof_type?: string[];
  };
  ldp_vc?: {
    proof_type?: string[];
  };
}

/**
 * SIOP Response
 */
export interface SIOPResponse {
  id_token: string;
  state?: string;
  vp_token?: string; // Verifiable Presentation token
}

/**
 * Self-Issued ID Token Payload
 */
export interface SelfIssuedIDToken extends JWTPayload {
  iss: string; // 'https://self-issued.me/v2' or DID
  sub: string; // DID or subject identifier
  aud: string; // Client ID
  sub_jwk?: JWK; // Subject's public key
  did?: string; // DID of the subject
  did_doc?: DIDDocument; // DID Document
  // VP token reference
  _vp_token?: {
    presentation_submission?: PresentationSubmission;
  };
}

/**
 * JSON Web Key (JWK)
 */
export interface JWK {
  kty: string; // Key type
  use?: string; // Key use
  key_ops?: string[]; // Key operations
  alg?: string; // Algorithm
  kid?: string; // Key ID
  x5u?: string; // X.509 URL
  x5c?: string[]; // X.509 certificate chain
  x5t?: string; // X.509 thumbprint
  'x5t#S256'?: string; // X.509 thumbprint (SHA-256)
  // RSA keys
  n?: string; // Modulus
  e?: string; // Exponent
  d?: string; // Private exponent
  p?: string; // First prime factor
  q?: string; // Second prime factor
  dp?: string; // First factor CRT exponent
  dq?: string; // Second factor CRT exponent
  qi?: string; // First CRT coefficient
  // EC keys
  crv?: string; // Curve
  x?: string; // X coordinate
  y?: string; // Y coordinate
  // Symmetric keys
  k?: string; // Key value
}

/**
 * DID Document (simplified)
 */
export interface DIDDocument {
  '@context'?: string | string[];
  id: string; // DID
  alsoKnownAs?: string[];
  controller?: string | string[];
  verificationMethod?: VerificationMethod[];
  authentication?: (string | VerificationMethod)[];
  assertionMethod?: (string | VerificationMethod)[];
  keyAgreement?: (string | VerificationMethod)[];
  capabilityInvocation?: (string | VerificationMethod)[];
  capabilityDelegation?: (string | VerificationMethod)[];
  service?: ServiceEndpoint[];
}

/**
 * Verification Method
 */
export interface VerificationMethod {
  id: string;
  type: string;
  controller: string;
  publicKeyJwk?: JWK;
  publicKeyMultibase?: string;
  publicKeyBase58?: string;
  blockchainAccountId?: string;
}

/**
 * Service Endpoint
 */
export interface ServiceEndpoint {
  id: string;
  type: string;
  serviceEndpoint: string | object;
}

/**
 * Verifiable Presentation
 */
export interface VerifiablePresentation {
  '@context': string | string[];
  type: string | string[];
  verifiableCredential?: VerifiableCredential[];
  holder?: string; // DID of the holder
  proof?: Proof | Proof[];
  // Presentation submission
  presentation_submission?: PresentationSubmission;
}

/**
 * Verifiable Credential
 */
export interface VerifiableCredential {
  '@context': string | string[];
  type: string | string[];
  issuer: string | { id: string; [key: string]: unknown };
  issuanceDate: string;
  expirationDate?: string;
  credentialSubject: CredentialSubject;
  proof?: Proof | Proof[];
}

/**
 * Credential Subject
 */
export interface CredentialSubject {
  id?: string; // DID of the subject
  [key: string]: unknown;
}

/**
 * Cryptographic Proof
 */
export interface Proof {
  type: string;
  created: string;
  verificationMethod: string;
  proofPurpose: string;
  challenge?: string;
  domain?: string;
  proofValue?: string;
  jws?: string;
  [key: string]: unknown;
}

/**
 * Presentation Submission
 */
export interface PresentationSubmission {
  id: string;
  definition_id: string;
  descriptor_map: InputDescriptor[];
}

/**
 * Input Descriptor
 */
export interface InputDescriptor {
  id: string;
  format: string;
  path: string;
  path_nested?: InputDescriptor[];
}

/**
 * VP Token Payload
 */
export interface VPTokenPayload extends JWTPayload {
  iss: string; // DID of the presenter
  vp: VerifiablePresentation;
}

// ============================================================================
// Constants
// ============================================================================

const SIOP_CONSTANTS = {
  SELF_ISSUED_ISSUER_V2: 'https://self-issued.me/v2',
  DEFAULT_RESPONSE_TYPE: 'id_token',
  DEFAULT_SCOPE: 'openid',
  SUBJECT_SYNTAX_TYPES: {
    JWK_THUMBPRINT: 'urn:ietf:params:oauth:jwk-thumbprint',
    DID: 'did',
  },
  ID_TOKEN_TYPES: {
    SUBJECT_SIGNED: 'subject-signed_id_token',
    ATTESTER_SIGNED: 'attester-signed_id_token',
  },
} as const;

// ============================================================================
// SIOP Request Parsing and Validation
// ============================================================================

/**
 * Parse SIOP request from URL parameters
 */
export function parseSIOPRequest(searchParams: URLSearchParams): SIOPRequest {
  const request: SIOPRequest = {
    response_type: (searchParams.get('response_type') as 'id_token') || 'id_token',
    client_id: searchParams.get('client_id') || '',
    redirect_uri: searchParams.get('redirect_uri') || '',
    scope: searchParams.get('scope') || SIOP_CONSTANTS.DEFAULT_SCOPE,
    state: searchParams.get('state') || undefined,
    nonce: searchParams.get('nonce') || undefined,
    response_mode:
      (searchParams.get('response_mode') as 'form_post' | 'fragment' | 'query' | null) || undefined,
    request: searchParams.get('request') || undefined,
    request_uri: searchParams.get('request_uri') || undefined,
  };

  // Parse claims parameter
  const claimsParam = searchParams.get('claims');
  if (claimsParam) {
    try {
      request.claims = JSON.parse(claimsParam) as SIOPClaims;
    } catch {
      request.claims = claimsParam;
    }
  }

  // Parse registration parameter
  const registrationParam = searchParams.get('registration');
  if (registrationParam) {
    try {
      request.registration = JSON.parse(registrationParam) as SIOPClientMetadata;
    } catch {
      request.registration = registrationParam;
    }
  }

  return request;
}

/**
 * Validate SIOP request parameters
 */
export function validateSIOPRequest(request: SIOPRequest): void {
  // Validate required parameters
  if (!request.client_id) {
    throw new Error('Missing required parameter: client_id');
  }

  if (!request.redirect_uri) {
    throw new Error('Missing required parameter: redirect_uri');
  }

  // Validate response_type
  if (request.response_type !== 'id_token') {
    throw new Error('Invalid response_type: only "id_token" is supported for SIOP');
  }

  // Validate redirect URI format
  try {
    new URL(request.redirect_uri);
  } catch {
    throw new Error('Invalid redirect_uri format');
  }

  // Validate scope contains 'openid'
  if (!request.scope.includes('openid')) {
    throw new Error('Scope must include "openid"');
  }

  // If request object is present, it should be validated separately
  if (request.request && !request.request_uri) {
    try {
      decodeJWT(request.request);
    } catch {
      throw new Error('Invalid request object: not a valid JWT');
    }
  }
}

/**
 * Resolve SIOP request object from JWT or URI
 */
export async function resolveSIOPRequestObject(request: SIOPRequest): Promise<SIOPRequest> {
  if (request.request_uri) {
    // Fetch request object from URI
    try {
      const response = await fetch(request.request_uri);
      if (!response.ok) {
        throw new Error(`Failed to fetch request object: ${response.status}`);
      }
      const requestObject = await response.text();
      return parseJWTRequestObject(requestObject);
    } catch (error) {
      throw new Error(
        `Failed to resolve request_uri: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  } else if (request.request) {
    // Parse inline request object
    return parseJWTRequestObject(request.request);
  }

  return request;
}

/**
 * Parse JWT request object
 */
function parseJWTRequestObject(jwt: string): SIOPRequest {
  const { payload } = decodeJWT(jwt);

  return {
    response_type: (payload.response_type as 'id_token') || 'id_token',
    client_id: (payload.client_id as string) || '',
    redirect_uri: (payload.redirect_uri as string) || '',
    scope: (payload.scope as string) || SIOP_CONSTANTS.DEFAULT_SCOPE,
    state: (payload.state as string) || undefined,
    nonce: (payload.nonce as string) || undefined,
    response_mode:
      (payload.response_mode as 'form_post' | 'fragment' | 'query' | null) || undefined,
    claims: (payload.claims as SIOPClaims) || undefined,
    registration: (payload.registration as SIOPClientMetadata) || undefined,
  };
}

// ============================================================================
// Self-Issued ID Token Generation
// ============================================================================

/**
 * Generate self-issued ID token
 */
export async function generateSelfIssuedIDToken(
  request: SIOPRequest,
  did: string,
  privateKey: CryptoKey,
  additionalClaims?: Record<string, unknown>
): Promise<string> {
  const now = Math.floor(Date.now() / 1000);

  // Create ID token payload
  const payload: SelfIssuedIDToken = {
    iss: SIOP_CONSTANTS.SELF_ISSUED_ISSUER_V2,
    sub: did,
    aud: request.client_id,
    exp: now + 3600, // 1 hour
    iat: now,
    nbf: now,
    jti: generateRandomString(16),
    did,
    ...additionalClaims,
  };

  // Add nonce if present
  if (request.nonce) {
    payload.nonce = request.nonce;
  }

  // Create header
  const header: JWTHeader = {
    typ: 'JWT',
    alg: 'ES256', // Assuming ES256, adjust based on key type
  };

  // Sign the token
  return signJWT(header, payload, privateKey);
}

/**
 * Generate subject JWK thumbprint
 */
export async function generateSubjectJWKThumbprint(publicKey: CryptoKey): Promise<string> {
  // Export public key as JWK
  const jwk = await crypto.subtle.exportKey('jwk', publicKey);

  // Create canonical JWK for thumbprint
  const canonicalJWK = {
    kty: jwk.kty,
    ...(jwk.crv && { crv: jwk.crv }),
    ...(jwk.x && { x: jwk.x }),
    ...(jwk.y && { y: jwk.y }),
    ...(jwk.n && { n: jwk.n }),
    ...(jwk.e && { e: jwk.e }),
  };

  // Calculate SHA-256 thumbprint
  const canonicalString = JSON.stringify(canonicalJWK);
  const encoder = new TextEncoder();
  const data = encoder.encode(canonicalString);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));

  // Base64url encode
  return btoa(String.fromCharCode(...hashArray))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

// ============================================================================
// DID Document Resolution
// ============================================================================

/**
 * Resolve DID document (simplified implementation)
 * In production, use a proper DID resolver library
 */
export async function resolveDIDDocument(did: string): Promise<DIDDocument> {
  // This is a simplified implementation
  // In practice, you would use a DID resolver that supports multiple DID methods

  if (did.startsWith('did:web:')) {
    return resolveWebDID(did);
  }
  if (did.startsWith('did:key:')) {
    return resolveKeyDID(did);
  }
  if (did.startsWith('did:sonr:')) {
    return resolveSonrDID(did);
  }
  throw new Error(`Unsupported DID method: ${did}`);
}

/**
 * Resolve did:web DID
 */
async function resolveWebDID(did: string): Promise<DIDDocument> {
  // Extract domain from did:web
  const domain = did.replace('did:web:', '').replace(/:/g, '/');
  const url = `https://${domain}/.well-known/did.json`;

  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`Failed to resolve DID document: ${response.status}`);
    }
    return response.json() as Promise<DIDDocument>;
  } catch (error) {
    throw new Error(
      `Failed to resolve did:web: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
  }
}

/**
 * Resolve did:key DID (simplified)
 */
async function resolveKeyDID(did: string): Promise<DIDDocument> {
  // This is a very simplified implementation
  // In practice, you would properly decode the multibase key

  const verificationMethod: VerificationMethod = {
    id: `${did}#keys-1`,
    type: 'Ed25519VerificationKey2020',
    controller: did,
    // In practice, decode the key from the DID
    publicKeyMultibase: did.split(':')[2],
  };

  return {
    id: did,
    verificationMethod: [verificationMethod],
    authentication: [verificationMethod.id],
    assertionMethod: [verificationMethod.id],
  };
}

/**
 * Resolve did:sonr DID from blockchain
 */
async function resolveSonrDID(did: string): Promise<DIDDocument> {
  // This would query the Sonr blockchain for the DID document
  // For now, return a mock structure

  const identifier = did.split(':')[2];
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  try {
    const response = await fetch(`${apiUrl}/did/resolve/${identifier}`);
    if (!response.ok) {
      throw new Error(`Failed to resolve Sonr DID: ${response.status}`);
    }
    return response.json() as Promise<DIDDocument>;
  } catch (error) {
    throw new Error(
      `Failed to resolve did:sonr: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
  }
}

// ============================================================================
// Verifiable Presentation Utilities
// ============================================================================

/**
 * Create Verifiable Presentation token
 */
export async function createVPToken(
  presentation: VerifiablePresentation,
  holderDID: string,
  privateKey: CryptoKey,
  audience: string,
  nonce?: string
): Promise<string> {
  const now = Math.floor(Date.now() / 1000);

  const payload: VPTokenPayload = {
    iss: holderDID,
    sub: holderDID,
    aud: audience,
    exp: now + 3600,
    iat: now,
    nbf: now,
    jti: generateRandomString(16),
    vp: presentation,
  };

  if (nonce) {
    payload.nonce = nonce;
  }

  const header: JWTHeader = {
    typ: 'JWT',
    alg: 'ES256',
  };

  return signJWT(header, payload, privateKey);
}

/**
 * Validate Verifiable Presentation
 */
export function validateVP(vp: VerifiablePresentation): void {
  // Basic validation
  if (!vp['@context'] || !vp.type) {
    throw new Error('VP must have @context and type');
  }

  // Validate context
  const contexts = Array.isArray(vp['@context']) ? vp['@context'] : [vp['@context']];
  if (!contexts.includes('https://www.w3.org/2018/credentials/v1')) {
    throw new Error('VP must include credentials v1 context');
  }

  // Validate type
  const types = Array.isArray(vp.type) ? vp.type : [vp.type];
  if (!types.includes('VerifiablePresentation')) {
    throw new Error('VP must include VerifiablePresentation type');
  }

  // Validate credentials if present
  if (vp.verifiableCredential) {
    vp.verifiableCredential.forEach((vc, index) => {
      try {
        validateVC(vc);
      } catch (error) {
        throw new Error(
          `Invalid credential at index ${index}: ${error instanceof Error ? error.message : 'Unknown error'}`
        );
      }
    });
  }
}

/**
 * Validate Verifiable Credential
 */
export function validateVC(vc: VerifiableCredential): void {
  if (!vc['@context'] || !vc.type || !vc.issuer || !vc.issuanceDate || !vc.credentialSubject) {
    throw new Error(
      'VC must have required fields: @context, type, issuer, issuanceDate, credentialSubject'
    );
  }

  // Validate expiration
  if (vc.expirationDate) {
    const expirationDate = new Date(vc.expirationDate);
    if (expirationDate < new Date()) {
      throw new Error('Credential has expired');
    }
  }

  // Additional validation would include signature verification
}

// ============================================================================
// SIOP Response Builder
// ============================================================================

/**
 * Build SIOP response
 */
export async function buildSIOPResponse(
  request: SIOPRequest,
  did: string,
  privateKey: CryptoKey,
  vp?: VerifiablePresentation,
  additionalClaims?: Record<string, unknown>
): Promise<SIOPResponse> {
  // Generate self-issued ID token
  const idToken = await generateSelfIssuedIDToken(request, did, privateKey, additionalClaims);

  const response: SIOPResponse = {
    id_token: idToken,
  };

  // Add state if present in request
  if (request.state) {
    response.state = request.state;
  }

  // Add VP token if VP is provided
  if (vp) {
    validateVP(vp);
    response.vp_token = await createVPToken(vp, did, privateKey, request.client_id, request.nonce);
  }

  return response;
}

/**
 * Submit SIOP response to client
 */
export async function submitSIOPResponse(
  response: SIOPResponse,
  redirectUri: string,
  responseMode: 'form_post' | 'fragment' | 'query' = 'fragment'
): Promise<void> {
  if (responseMode === 'form_post') {
    // Submit via form post
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = redirectUri;

    Object.entries(response).forEach(([key, value]) => {
      const input = document.createElement('input');
      input.type = 'hidden';
      input.name = key;
      input.value = value;
      form.appendChild(input);
    });

    document.body.appendChild(form);
    form.submit();
  } else {
    // Submit via redirect with parameters
    const url = new URL(redirectUri);
    const params = responseMode === 'fragment' ? new URLSearchParams() : url.searchParams;

    Object.entries(response).forEach(([key, value]) => {
      params.append(key, value);
    });

    const finalUrl =
      responseMode === 'fragment'
        ? `${url.origin}${url.pathname}#${params.toString()}`
        : url.toString();

    window.location.href = finalUrl;
  }
}

// ============================================================================
// Dynamic Client Registration
// ============================================================================

/**
 * Validate client metadata for dynamic registration
 */
export function validateClientMetadata(metadata: SIOPClientMetadata): void {
  // Validate required fields for SIOP
  if (!metadata.subject_type) {
    metadata.subject_type = 'public';
  }

  // Validate redirect URIs
  if (metadata.redirect_uris) {
    metadata.redirect_uris.forEach((uri) => {
      try {
        new URL(uri);
      } catch {
        throw new Error(`Invalid redirect URI: ${uri}`);
      }
    });
  }

  // Validate response types
  if (metadata.response_types && !metadata.response_types.includes('id_token')) {
    throw new Error('SIOP clients must support "id_token" response type');
  }

  // Validate subject syntax types
  if (metadata.subject_syntax_types_supported) {
    const validTypes = Object.values(SIOP_CONSTANTS.SUBJECT_SYNTAX_TYPES);
    const hasValidType = metadata.subject_syntax_types_supported.some((type) =>
      validTypes.includes(
        type as (typeof SIOP_CONSTANTS.SUBJECT_SYNTAX_TYPES)[keyof typeof SIOP_CONSTANTS.SUBJECT_SYNTAX_TYPES]
      )
    );

    if (!hasValidType) {
      throw new Error('Client must support at least one valid subject syntax type');
    }
  }
}

// ============================================================================
// JWT Signing Utilities
// ============================================================================

/**
 * Sign JWT with private key
 */
async function signJWT(
  header: JWTHeader,
  payload: Record<string, unknown>,
  privateKey: CryptoKey
): Promise<string> {
  // Encode header and payload
  const encodedHeader = base64URLEncode(new TextEncoder().encode(JSON.stringify(header)));
  const encodedPayload = base64URLEncode(new TextEncoder().encode(JSON.stringify(payload)));

  // Create signing input
  const signingInput = `${encodedHeader}.${encodedPayload}`;

  // Sign
  const signature = await crypto.subtle.sign(
    'ECDSA',
    privateKey,
    new TextEncoder().encode(signingInput)
  );

  // Encode signature
  const encodedSignature = base64URLEncode(new Uint8Array(signature));

  return `${signingInput}.${encodedSignature}`;
}

/**
 * Base64URL encode utility
 */
function base64URLEncode(buffer: Uint8Array): string {
  return btoa(String.fromCharCode(...Array.from(buffer)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

// ============================================================================
// High-Level SIOP Client
// ============================================================================

/**
 * SIOP Client for handling SIOP requests and responses
 */
export class SIOPClient {
  private did: string;
  private privateKey?: CryptoKey;
  private didDocument?: DIDDocument;

  constructor(did: string, privateKey?: CryptoKey) {
    this.did = did;
    this.privateKey = privateKey;
  }

  /**
   * Initialize the client by resolving DID document
   */
  async initialize(): Promise<void> {
    this.didDocument = await resolveDIDDocument(this.did);
  }

  /**
   * Handle incoming SIOP request
   */
  async handleSIOPRequest(
    requestUrl: string,
    credentials?: VerifiableCredential[]
  ): Promise<SIOPResponse> {
    const url = new URL(requestUrl);
    let request = parseSIOPRequest(url.searchParams);

    // Resolve request object if present
    request = await resolveSIOPRequestObject(request);

    // Validate request
    validateSIOPRequest(request);

    if (!this.privateKey) {
      throw new Error('Private key required for signing');
    }

    // Create VP if credentials are provided
    let vp: VerifiablePresentation | undefined;
    if (credentials && credentials.length > 0) {
      vp = {
        '@context': ['https://www.w3.org/2018/credentials/v1'],
        type: ['VerifiablePresentation'],
        verifiableCredential: credentials,
        holder: this.did,
      };
    }

    // Build and return response
    return buildSIOPResponse(request, this.did, this.privateKey, vp);
  }

  /**
   * Submit SIOP response
   */
  async submitResponse(
    response: SIOPResponse,
    redirectUri: string,
    responseMode?: 'form_post' | 'fragment' | 'query'
  ): Promise<void> {
    await submitSIOPResponse(response, redirectUri, responseMode);
  }

  /**
   * Get DID document
   */
  getDIDDocument(): DIDDocument | undefined {
    return this.didDocument;
  }

  /**
   * Set private key
   */
  setPrivateKey(privateKey: CryptoKey): void {
    this.privateKey = privateKey;
  }
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Check if a request is a SIOP request
 */
export function isSIOPRequest(searchParams: URLSearchParams): boolean {
  const responseType = searchParams.get('response_type');
  const scope = searchParams.get('scope');

  return responseType === 'id_token' && scope?.includes('openid') === true;
}

/**
 * Extract DID from self-issued ID token
 */
export function extractDIDFromSelfIssuedToken(idToken: string): string | null {
  try {
    const { payload } = decodeJWT(idToken);
    return (payload as SelfIssuedIDToken).did || null;
  } catch {
    return null;
  }
}

/**
 * Validate self-issued ID token structure
 */
export function validateSelfIssuedIDToken(
  idToken: string,
  expectedAudience: string
): SelfIssuedIDToken {
  const { payload } = decodeJWT(idToken);
  const token = payload as SelfIssuedIDToken;

  // Validate issuer
  if (token.iss !== SIOP_CONSTANTS.SELF_ISSUED_ISSUER_V2 && !token.did?.startsWith('did:')) {
    throw new Error('Invalid issuer for self-issued ID token');
  }

  // Validate audience
  if (token.aud !== expectedAudience) {
    throw new Error('Invalid audience');
  }

  // Validate expiration
  const now = Math.floor(Date.now() / 1000);
  if (token.exp && token.exp < now) {
    throw new Error('Token has expired');
  }

  // Validate not before
  if (token.nbf && token.nbf > now) {
    throw new Error('Token not yet valid');
  }

  return token;
}
