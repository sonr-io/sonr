import { type NextRequest, NextResponse } from 'next/server';

const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface OIDCAuthorizationParams {
  response_type: string;
  client_id: string;
  redirect_uri: string;
  scope: string;
  state?: string;
  nonce?: string;
  code_challenge?: string;
  code_challenge_method?: string;
}

/**
 * GET /api/oidc/authorize
 *
 * OIDC Authorization endpoint that initiates the authorization code flow.
 * This endpoint proxies the request to the Go bridge handler at /oidc/authorize.
 */
export async function GET(request: NextRequest): Promise<NextResponse> {
  try {
    const { searchParams } = new URL(request.url);

    // Extract and validate required OIDC parameters
    const authParams: OIDCAuthorizationParams = {
      response_type: searchParams.get('response_type') || '',
      client_id: searchParams.get('client_id') || '',
      redirect_uri: searchParams.get('redirect_uri') || '',
      scope: searchParams.get('scope') || '',
      state: searchParams.get('state') || undefined,
      nonce: searchParams.get('nonce') || undefined,
      code_challenge: searchParams.get('code_challenge') || undefined,
      code_challenge_method: searchParams.get('code_challenge_method') || undefined,
    };

    // Validate required parameters
    if (
      !authParams.response_type ||
      !authParams.client_id ||
      !authParams.redirect_uri ||
      !authParams.scope
    ) {
      return NextResponse.json(
        {
          error: 'invalid_request',
          error_description:
            'Missing required parameters: response_type, client_id, redirect_uri, or scope',
        },
        { status: 400 }
      );
    }

    // Build query string for the bridge handler
    const queryParams = new URLSearchParams();
    Object.entries(authParams).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        queryParams.append(key, value);
      }
    });

    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/authorize?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        // Forward authentication and session headers
        Authorization: request.headers.get('Authorization') || '',
        Cookie: request.headers.get('Cookie') || '',
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
        Referer: request.headers.get('Referer') || '',
      },
    });

    // Handle different response types from the bridge
    if (response.status === 302) {
      // Bridge handler returned a redirect - follow it
      const location = response.headers.get('Location');
      if (location) {
        return NextResponse.redirect(location);
      }
    }

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({
        error: 'authorization_error',
        error_description: 'Authorization request failed',
      }));

      console.error('OIDC Authorization error:', response.status, errorData);
      return NextResponse.json(errorData, { status: response.status });
    }

    const responseData = await response.json();

    // Return the authorization response with appropriate headers
    return NextResponse.json(responseData, {
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
    console.error('OIDC Authorization proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during authorization',
      },
      { status: 500 }
    );
  }
}

/**
 * POST /api/oidc/authorize
 *
 * Handle authorization requests sent via POST (form-encoded).
 */
export async function POST(request: NextRequest): Promise<NextResponse> {
  try {
    const formData = await request.formData();

    // Extract parameters from form data
    const authParams: OIDCAuthorizationParams = {
      response_type: formData.get('response_type')?.toString() || '',
      client_id: formData.get('client_id')?.toString() || '',
      redirect_uri: formData.get('redirect_uri')?.toString() || '',
      scope: formData.get('scope')?.toString() || '',
      state: formData.get('state')?.toString() || undefined,
      nonce: formData.get('nonce')?.toString() || undefined,
      code_challenge: formData.get('code_challenge')?.toString() || undefined,
      code_challenge_method: formData.get('code_challenge_method')?.toString() || undefined,
    };

    // Validate required parameters
    if (
      !authParams.response_type ||
      !authParams.client_id ||
      !authParams.redirect_uri ||
      !authParams.scope
    ) {
      return NextResponse.json(
        {
          error: 'invalid_request',
          error_description:
            'Missing required parameters: response_type, client_id, redirect_uri, or scope',
        },
        { status: 400 }
      );
    }

    // Forward as form data to the bridge handler
    const bridgeFormData = new FormData();
    Object.entries(authParams).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        bridgeFormData.append(key, value);
      }
    });

    const response = await fetch(`${BRIDGE_URL}/oidc/authorize`, {
      method: 'POST',
      body: bridgeFormData,
      headers: {
        // Forward authentication and session headers
        Authorization: request.headers.get('Authorization') || '',
        Cookie: request.headers.get('Cookie') || '',
        'User-Agent': request.headers.get('User-Agent') || '',
        Origin: request.headers.get('Origin') || '',
        Referer: request.headers.get('Referer') || '',
      },
    });

    // Handle redirect response
    if (response.status === 302) {
      const location = response.headers.get('Location');
      if (location) {
        return NextResponse.redirect(location);
      }
    }

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({
        error: 'authorization_error',
        error_description: 'Authorization request failed',
      }));

      return NextResponse.json(errorData, { status: response.status });
    }

    const responseData = await response.json();
    return NextResponse.json(responseData);
  } catch (error) {
    console.error('OIDC Authorization POST proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during authorization',
      },
      { status: 500 }
    );
  }
}

/**
 * OPTIONS /api/oidc/authorize
 *
 * Handle preflight CORS requests for the authorization endpoint.
 */
export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization, Cookie',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Max-Age': '86400',
    },
  });
}
