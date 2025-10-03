import { type NextRequest, NextResponse } from 'next/server';

export const dynamic = 'force-static';
export const revalidate = 3600; // Revalidate every hour

const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface OIDCTokenRequest {
  grant_type: string;
  code?: string;
  redirect_uri?: string;
  client_id: string;
  client_secret?: string;
  code_verifier?: string;
  refresh_token?: string;
}

interface OIDCTokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
}

/**
 * POST /api/oidc/token
 *
 * OIDC Token endpoint that exchanges authorization codes for access tokens.
 * This endpoint proxies the request to the Go bridge handler at /oidc/token.
 * Supports authorization_code, refresh_token, and client_credentials grant types.
 */
export async function POST(request: NextRequest): Promise<NextResponse> {
  try {
    const contentType = request.headers.get('Content-Type') || '';
    let tokenRequest: OIDCTokenRequest;

    // Parse request based on content type
    if (contentType.includes('application/x-www-form-urlencoded')) {
      const formData = await request.formData();
      tokenRequest = {
        grant_type: formData.get('grant_type')?.toString() || '',
        code: formData.get('code')?.toString() || undefined,
        redirect_uri: formData.get('redirect_uri')?.toString() || undefined,
        client_id: formData.get('client_id')?.toString() || '',
        client_secret: formData.get('client_secret')?.toString() || undefined,
        code_verifier: formData.get('code_verifier')?.toString() || undefined,
        refresh_token: formData.get('refresh_token')?.toString() || undefined,
      };
    } else if (contentType.includes('application/json')) {
      tokenRequest = await request.json();
    } else {
      return NextResponse.json(
        {
          error: 'invalid_request',
          error_description:
            'Content-Type must be application/x-www-form-urlencoded or application/json',
        },
        { status: 400 }
      );
    }

    // Validate required parameters
    if (!tokenRequest.grant_type || !tokenRequest.client_id) {
      return NextResponse.json(
        {
          error: 'invalid_request',
          error_description: 'Missing required parameters: grant_type and client_id',
        },
        { status: 400 }
      );
    }

    // Validate grant type specific parameters
    if (tokenRequest.grant_type === 'authorization_code') {
      if (!tokenRequest.code || !tokenRequest.redirect_uri) {
        return NextResponse.json(
          {
            error: 'invalid_request',
            error_description:
              'Missing required parameters for authorization_code grant: code and redirect_uri',
          },
          { status: 400 }
        );
      }
    } else if (tokenRequest.grant_type === 'refresh_token') {
      if (!tokenRequest.refresh_token) {
        return NextResponse.json(
          {
            error: 'invalid_request',
            error_description: 'Missing required parameter for refresh_token grant: refresh_token',
          },
          { status: 400 }
        );
      }
    }

    // Prepare form data for the bridge handler (OIDC typically uses form encoding)
    const bridgeFormData = new FormData();
    Object.entries(tokenRequest).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        bridgeFormData.append(key, value);
      }
    });

    // Forward the request to the Go bridge handler
    const response = await fetch(`${BRIDGE_URL}/oidc/token`, {
      method: 'POST',
      body: bridgeFormData,
      headers: {
        Accept: 'application/json',
        // Forward client authentication headers
        Authorization: request.headers.get('Authorization') || '',
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
          error_description: 'Token request failed',
        };
      }

      console.error('OIDC Token error:', response.status, errorData);
      return NextResponse.json(errorData, { status: response.status });
    }

    const tokenResponse: OIDCTokenResponse = await response.json();

    // Return the token response with appropriate headers
    return NextResponse.json(tokenResponse, {
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
    console.error('OIDC Token proxy error:', error);
    return NextResponse.json(
      {
        error: 'server_error',
        error_description: 'Internal server error during token exchange',
      },
      { status: 500 }
    );
  }
}

/**
 * OPTIONS /api/oidc/token
 *
 * Handle preflight CORS requests for the token endpoint.
 */
export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, {
    status: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Max-Age': '86400',
    },
  });
}

/**
 * GET /api/oidc/token
 *
 * Return method not allowed for GET requests to token endpoint.
 */
export async function GET(): Promise<NextResponse> {
  return NextResponse.json(
    {
      error: 'invalid_request',
      error_description: 'GET method not allowed for token endpoint. Use POST.',
    },
    {
      status: 405,
      headers: {
        Allow: 'POST, OPTIONS',
      },
    }
  );
}
