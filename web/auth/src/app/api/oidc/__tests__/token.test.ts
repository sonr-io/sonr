import { NextRequest } from 'next/server';
import { POST } from '../token/route';

describe('OIDC Token Endpoint', () => {
  it('handles authorization code token exchange', async () => {
    const mockRequest = new NextRequest(new URL('https://example.com/token'), {
      method: 'POST',
      body: JSON.stringify({
        grant_type: 'authorization_code',
        code: 'valid_code',
        client_id: 'test_client',
        redirect_uri: 'https://example.com/callback',
        code_verifier: 'test_verifier',
      }),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    });

    const response = await POST(mockRequest);

    expect(response.status).toBe(200);

    const tokens = await response.json();
    expect(tokens).toHaveProperty('access_token');
    expect(tokens).toHaveProperty('token_type', 'Bearer');
    expect(tokens).toHaveProperty('expires_in');
  });

  it('handles refresh token grant', async () => {
    const mockRequest = new NextRequest(new URL('https://example.com/token'), {
      method: 'POST',
      body: JSON.stringify({
        grant_type: 'refresh_token',
        refresh_token: 'valid_refresh_token',
        client_id: 'test_client',
      }),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    });

    const response = await POST(mockRequest);

    expect(response.status).toBe(200);

    const tokens = await response.json();
    expect(tokens).toHaveProperty('access_token');
    expect(tokens).toHaveProperty('token_type', 'Bearer');
  });

  it('rejects invalid grant types', async () => {
    const mockRequest = new NextRequest(new URL('https://example.com/token'), {
      method: 'POST',
      body: JSON.stringify({
        grant_type: 'invalid_grant',
        client_id: 'test_client',
      }),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    });

    const response = await POST(mockRequest);

    expect(response.status).toBe(400);

    const error = await response.json();
    expect(error).toHaveProperty('error', 'unsupported_grant_type');
  });
});
