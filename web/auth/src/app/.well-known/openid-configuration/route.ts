import { type NextRequest, NextResponse } from 'next/server';

const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

/**
 * GET /.well-known/openid-configuration
 *
 * OpenID Connect Discovery endpoint that returns the OIDC provider configuration.
 * This endpoint proxies the request to the Go bridge handler at /oidc/discovery.
 */
export async function GET(request: NextRequest): Promise<NextResponse> {
  try {
    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/discovery`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        // Forward relevant headers
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
      },
    });

    if (!response.ok) {
      console.error('OIDC Discovery error:', response.status, response.statusText);
      return NextResponse.json(
        {
          error: 'discovery_error',
          error_description: 'Failed to retrieve OIDC discovery configuration',
        },
        { status: response.status }
      );
    }

    const discoveryConfig = await response.json();

    // Return the discovery configuration with CORS headers
    return NextResponse.json(discoveryConfig, {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'public, max-age=3600',
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    });
  } catch (error) {
    console.error('OIDC Discovery proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during discovery configuration retrieval',
      },
      { status: 500 }
    );
  }
}

/**
 * OPTIONS /.well-known/openid-configuration
 *
 * Handle preflight CORS requests for the discovery endpoint.
 */
export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400',
    },
  });
}
