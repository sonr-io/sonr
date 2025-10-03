/**
 * UCAN delegation utilities for OAuth2 integration
 */

import { JWK, SignJWT, importJWK } from 'jose';

/**
 * UCAN token structure
 */
export interface UCANToken {
  iss: string; // Issuer DID
  aud: string; // Audience DID
  exp: number; // Expiration timestamp
  nbf?: number; // Not before timestamp
  att: UCANCapability[]; // Attenuations/capabilities
  prf?: string[]; // Proof chain
  fct?: any[]; // Facts
}

/**
 * UCAN capability/attenuation
 */
export interface UCANCapability {
  with: string; // Resource URI
  can: string; // Action/capability
  nb?: Record<string, any>; // Caveats/constraints
}

/**
 * OAuth scope to UCAN capability mapping
 */
export interface ScopeMapping {
  scope: string;
  capabilities: UCANCapability[];
}

/**
 * Default scope mappings
 */
export const DEFAULT_SCOPE_MAPPINGS: ScopeMapping[] = [
  {
    scope: 'openid',
    capabilities: [{ with: 'did:*', can: 'identify' }],
  },
  {
    scope: 'profile',
    capabilities: [
      { with: 'profile:*', can: 'read' },
      { with: 'did:*', can: 'resolve' },
    ],
  },
  {
    scope: 'vault:read',
    capabilities: [
      { with: 'vault:*', can: 'read' },
      { with: 'vault:*', can: 'list' },
    ],
  },
  {
    scope: 'vault:write',
    capabilities: [
      { with: 'vault:*', can: 'write' },
      { with: 'vault:*', can: 'create' },
      { with: 'vault:*', can: 'update' },
      { with: 'vault:*', can: 'delete' },
    ],
  },
  {
    scope: 'vault:sign',
    capabilities: [
      { with: 'vault:*', can: 'sign' },
      { with: 'tx:*', can: 'sign' },
    ],
  },
  {
    scope: 'service:manage',
    capabilities: [
      { with: 'service:*', can: 'create' },
      { with: 'service:*', can: 'update' },
      { with: 'service:*', can: 'delete' },
      { with: 'service:*', can: 'list' },
    ],
  },
];

/**
 * Map OAuth scopes to UCAN capabilities
 */
export function scopesToCapabilities(scopes: string[]): UCANCapability[] {
  const capabilities: UCANCapability[] = [];

  for (const scope of scopes) {
    const mapping = DEFAULT_SCOPE_MAPPINGS.find((m) => m.scope === scope);
    if (mapping) {
      capabilities.push(...mapping.capabilities);
    } else {
      // Handle custom scopes
      capabilities.push({
        with: `custom:${scope}`,
        can: '*',
      });
    }
  }

  // Deduplicate capabilities
  const unique = new Map<string, UCANCapability>();
  for (const cap of capabilities) {
    const key = `${cap.with}:${cap.can}`;
    if (!unique.has(key)) {
      unique.set(key, cap);
    }
  }

  return Array.from(unique.values());
}

/**
 * Parse UCAN token from JWT string
 */
export async function parseUCAN(token: string): Promise<UCANToken> {
  // Decode JWT without verification (client-side)
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid UCAN token format');
  }

  const payload = JSON.parse(atob(parts[1]));

  return {
    iss: payload.iss,
    aud: payload.aud,
    exp: payload.exp,
    nbf: payload.nbf,
    att: payload.att || [],
    prf: payload.prf || [],
    fct: payload.fct || [],
  };
}

/**
 * Validate UCAN token capabilities
 */
export function validateCapabilities(
  token: UCANToken,
  requiredCapabilities: UCANCapability[]
): boolean {
  for (const required of requiredCapabilities) {
    const hasCapability = token.att.some((cap) => {
      // Check if the capability matches
      if (!matchResource(cap.with, required.with)) {
        return false;
      }

      if (!matchAction(cap.can, required.can)) {
        return false;
      }

      // Check caveats if present
      if (required.nb) {
        if (!cap.nb) return false;

        for (const [key, value] of Object.entries(required.nb)) {
          if (cap.nb[key] !== value) {
            return false;
          }
        }
      }

      return true;
    });

    if (!hasCapability) {
      return false;
    }
  }

  return true;
}

/**
 * Match resource patterns
 */
function matchResource(pattern: string, resource: string): boolean {
  // Handle wildcards
  if (pattern === '*' || resource === '*') {
    return true;
  }

  // Handle prefix wildcards
  if (pattern.endsWith('*')) {
    const prefix = pattern.slice(0, -1);
    return resource.startsWith(prefix);
  }

  // Exact match
  return pattern === resource;
}

/**
 * Match action patterns
 */
function matchAction(pattern: string, action: string): boolean {
  // Handle wildcards
  if (pattern === '*' || action === '*') {
    return true;
  }

  // Exact match
  return pattern === action;
}

/**
 * Create a delegation chain
 */
export async function createDelegation(
  issuerDID: string,
  audienceDID: string,
  capabilities: UCANCapability[],
  expiresIn = 3600,
  proofs?: string[]
): Promise<string> {
  const now = Math.floor(Date.now() / 1000);

  const payload = {
    iss: issuerDID,
    aud: audienceDID,
    exp: now + expiresIn,
    nbf: now,
    att: capabilities,
    prf: proofs || [],
  };

  // In production, this would be signed with the issuer's private key
  // For client-side, we'll create an unsigned token
  const header = { alg: 'none', typ: 'JWT', ucv: '0.9.0' };

  const encodedHeader = btoa(JSON.stringify(header))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');

  const encodedPayload = btoa(JSON.stringify(payload))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');

  return `${encodedHeader}.${encodedPayload}.`;
}

/**
 * Verify UCAN delegation chain
 */
export async function verifyDelegationChain(token: string, rootDID: string): Promise<boolean> {
  try {
    const ucan = await parseUCAN(token);

    // Check expiration
    const now = Math.floor(Date.now() / 1000);
    if (ucan.exp && ucan.exp < now) {
      return false;
    }

    // Check not before
    if (ucan.nbf && ucan.nbf > now) {
      return false;
    }

    // Check if this is the root delegation
    if (ucan.iss === rootDID) {
      return true;
    }

    // Verify proof chain
    if (!ucan.prf || ucan.prf.length === 0) {
      return false;
    }

    // Recursively verify each proof
    for (const proof of ucan.prf) {
      const isValid = await verifyDelegationChain(proof, rootDID);
      if (!isValid) {
        return false;
      }

      // Verify that the proof delegates to this token's issuer
      const proofUCAN = await parseUCAN(proof);
      if (proofUCAN.aud !== ucan.iss) {
        return false;
      }

      // Verify capabilities are attenuated
      if (!isAttenuated(proofUCAN.att, ucan.att)) {
        return false;
      }
    }

    return true;
  } catch {
    return false;
  }
}

/**
 * Check if capabilities are properly attenuated
 */
function isAttenuated(parentCaps: UCANCapability[], childCaps: UCANCapability[]): boolean {
  // Each child capability must be allowed by at least one parent capability
  for (const childCap of childCaps) {
    const isAllowed = parentCaps.some((parentCap) => {
      // Check resource is same or more specific
      if (!matchResource(parentCap.with, childCap.with)) {
        return false;
      }

      // Check action is same or more specific
      if (!matchAction(parentCap.can, childCap.can)) {
        return false;
      }

      // Check caveats are same or more restrictive
      if (childCap.nb) {
        if (!parentCap.nb) {
          // Child has caveats but parent doesn't - this is more restrictive, so OK
          return true;
        }

        // All parent caveats must be present in child
        for (const [key, value] of Object.entries(parentCap.nb)) {
          if (childCap.nb[key] !== value) {
            return false;
          }
        }
      }

      return true;
    });

    if (!isAllowed) {
      return false;
    }
  }

  return true;
}

/**
 * Human-readable capability description
 */
export function describeCapability(cap: UCANCapability): string {
  const resource = cap.with.replace(':', ' ').replace('*', 'all');
  const action = cap.can.replace('*', 'all actions');

  let description = `Can ${action} on ${resource}`;

  if (cap.nb && Object.keys(cap.nb).length > 0) {
    const caveats = Object.entries(cap.nb)
      .map(([k, v]) => `${k}=${v}`)
      .join(', ');
    description += ` (with constraints: ${caveats})`;
  }

  return description;
}

/**
 * Export utilities
 */
export const UCANUtils = {
  scopesToCapabilities,
  parseUCAN,
  validateCapabilities,
  createDelegation,
  verifyDelegationChain,
  describeCapability,
};
