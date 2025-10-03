import { type NextRequest, NextResponse } from 'next/server';

const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface OIDCUserInfo {
  sub: string;
  name?: string;
  preferred_username?: string;
  email?: string;
  email_verified?: boolean;
  did?: string;
  vault_id?: string;
  updated_at?: number;
  claims?: Record<string, unknown>;
}

/**
 * GET /api/oidc/userinfo
 *
 * OIDC UserInfo endpoint that returns user information for a valid access token.
 * This endpoint proxies the request to the Go bridge handler at /oidc/userinfo.
 * Requires a valid Bearer token in the Authorization header.
 */
export async function GET(request: NextRequest): Promise<NextResponse> {
  try {
    // Extract access token from Authorization header
    const authHeader = request.headers.get('Authorization');
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return NextResponse.json(
        {
          error: 'invalid_token',
          error_description:
            'Missing or invalid Authorization header. Expected format: Bearer <access_token>',
        },
        {
          status: 401,
          headers: {
            'WWW-Authenticate': 'Bearer realm="OIDC UserInfo", error="invalid_token"',
          },
        }
      );
    }

    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/userinfo`, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        // Forward the Authorization header with the access token
        Authorization: authHeader,
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
      },
    });

    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        // Handle different error statuses
        switch (response.status) {
          case 401:
            errorData = {
              error: 'invalid_token',
              error_description: 'Access token is invalid or expired',
            };
            break;
          case 403:
            errorData = {
              error: 'insufficient_scope',
              error_description: 'Access token does not have sufficient scope',
            };
            break;
          default:
            errorData = {
              error: 'server_error',
              error_description: 'UserInfo request failed',
            };
        }
      }

      console.error('OIDC UserInfo error:', response.status, errorData);

      const responseHeaders: HeadersInit = {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-store, no-cache, must-revalidate',
        Pragma: 'no-cache',
      };

      // Add WWW-Authenticate header for 401 responses
      if (response.status === 401) {
        responseHeaders['WWW-Authenticate'] = 'Bearer realm="OIDC UserInfo", error="invalid_token"';
      }

      return NextResponse.json(errorData, {
        status: response.status,
        headers: responseHeaders,
      });
    }

    const userInfo: OIDCUserInfo = await response.json();

    // Return the user information with appropriate headers
    return NextResponse.json(userInfo, {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-store, no-cache, must-revalidate',
        Pragma: 'no-cache',
        'Access-Control-Allow-Origin': request.headers.get('Origin') || '*',
        'Access-Control-Allow-Credentials': 'true',
      },
    });
  } catch (error) {
    console.error('OIDC UserInfo proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during userinfo retrieval',
      },
      { status: 500 }
    );
  }
}

/**
 * POST /api/oidc/userinfo
 *
 * OIDC UserInfo endpoint that accepts POST requests with access token in form data.
 * This is an alternative method for clients that prefer POST over GET.
 */
export async function POST(request: NextRequest): Promise<NextResponse> {
  try {
    const contentType = request.headers.get('Content-Type') || '';
    let accessToken: string | null = null;

    // Extract access token from different sources
    const authHeader = request.headers.get('Authorization');
    if (authHeader?.startsWith('Bearer ')) {
      accessToken = authHeader.substring(7);
    } else if (contentType.includes('application/x-www-form-urlencoded')) {
      const formData = await request.formData();
      accessToken = formData.get('access_token')?.toString() || null;
    } else if (contentType.includes('application/json')) {
      const body = await request.json();
      accessToken = body.access_token || null;
    }

    if (!accessToken) {
      return NextResponse.json(
        {
          error: 'invalid_token',
          error_description: 'Access token required in Authorization header or request body',
        },
        {
          status: 401,
          headers: {
            'WWW-Authenticate': 'Bearer realm="OIDC UserInfo", error="invalid_token"',
          },
        }
      );
    }

    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/userinfo`, {
      method: 'GET', // Bridge handler expects GET
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${accessToken}`,
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
      },
    });

    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'server_error',
          error_description: 'UserInfo request failed',
        };
      }

      return NextResponse.json(errorData, { status: response.status });
    }

    const userInfo: OIDCUserInfo = await response.json();
    return NextResponse.json(userInfo, {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'no-store, no-cache, must-revalidate',
        Pragma: 'no-cache',
      },
    });
  } catch (error) {
    console.error('OIDC UserInfo POST proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during userinfo retrieval',
      },
      { status: 500 }
    );
  }
}

/**
 * OPTIONS /api/oidc/userinfo
 *
 * Handle preflight CORS requests for the userinfo endpoint.
 */
export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Max-Age': '86400',
    },
  });
}
