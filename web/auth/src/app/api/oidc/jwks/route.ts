import { type NextRequest, NextResponse } from 'next/server';

const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface JWK {
  kty: string;
  use?: string;
  kid: string;
  alg?: string;
  n?: string; // RSA modulus
  e?: string; // RSA exponent
  x?: string; // EC x coordinate
  y?: string; // EC y coordinate
  crv?: string; // EC curve
}

interface JWKSet {
  keys: JWK[];
}

/**
 * GET /api/oidc/jwks
 *
 * OIDC JSON Web Key Set (JWKS) endpoint that returns public keys used for token verification.
 * This endpoint proxies the request to the Go bridge handler at /oidc/jwks.
 * The returned keys are used by relying parties to verify JWT tokens issued by this OIDC provider.
 */
export async function GET(request: NextRequest): Promise<NextResponse> {
  try {
    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/jwks`, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
      },
    });

    if (!response.ok) {
      console.error('OIDC JWKS error:', response.status, response.statusText);

      let errorData;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'server_error',
          error_description: 'Failed to retrieve JSON Web Key Set',
        };
      }

      return NextResponse.json(errorData, { status: response.status });
    }

    const jwks: JWKSet = await response.json();

    // Validate JWKS structure
    if (!jwks || !Array.isArray(jwks.keys)) {
      console.error('Invalid JWKS response structure:', jwks);
      return NextResponse.json(
        {
          error: 'server_error',
          error_description: 'Invalid JWKS response from server',
        },
        { status: 500 }
      );
    }

    // Validate each key in the set
    const validKeys = jwks.keys.filter((key: JWK) => {
      // Basic validation for required JWK fields
      if (!key.kty || !key.kid) {
        console.warn('Invalid JWK missing required fields:', key);
        return false;
      }

      // Validate key type specific fields
      if (key.kty === 'RSA' && (!key.n || !key.e)) {
        console.warn('Invalid RSA JWK missing n or e:', key);
        return false;
      }

      if (key.kty === 'EC' && (!key.x || !key.y || !key.crv)) {
        console.warn('Invalid EC JWK missing x, y, or crv:', key);
        return false;
      }

      return true;
    });

    const validatedJWKS: JWKSet = {
      keys: validKeys,
    };

    // Return the JWKS with appropriate headers
    return NextResponse.json(validatedJWKS, {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        // JWKS can be cached longer since keys don't change frequently
        'Cache-Control': 'public, max-age=86400, stale-while-revalidate=43200',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET',
        'Access-Control-Allow-Headers': 'Content-Type',
        // Security headers for key endpoint
        'X-Content-Type-Options': 'nosniff',
        'X-Frame-Options': 'DENY',
      },
    });
  } catch (error) {
    console.error('OIDC JWKS proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during JWKS retrieval',
      },
      { status: 500 }
    );
  }
}

/**
 * OPTIONS /api/oidc/jwks
 *
 * Handle preflight CORS requests for the JWKS endpoint.
 */
export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Max-Age': '86400',
    },
  });
}

/**
 * POST /api/oidc/jwks
 *
 * Return method not allowed for POST requests to JWKS endpoint.
 */
export async function POST(): Promise<NextResponse> {
  return NextResponse.json(
    {
      error: 'method_not_allowed',
      error_description: 'POST method not allowed for JWKS endpoint. Use GET.',
    },
    {
      status: 405,
      headers: {
        Allow: 'GET, OPTIONS',
      },
    }
  );
}
