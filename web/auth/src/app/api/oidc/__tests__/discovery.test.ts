import { GET } from '../.well-known/openid-configuration/route';

describe('OIDC Discovery Endpoint', () => {
  it('returns a valid OpenID Connect configuration', async () => {
    const response = await GET(new Request('https://example.com/.well-known/openid-configuration'));

    expect(response.status).toBe(200);

    const config = await response.json();

    // Basic structure validation
    expect(config).toHaveProperty('issuer');
    expect(config).toHaveProperty('authorization_endpoint');
    expect(config).toHaveProperty('token_endpoint');
    expect(config).toHaveProperty('userinfo_endpoint');
    expect(config).toHaveProperty('jwks_uri');

    // Validate specific SIOP requirements
    expect(config.subject_syntax_types_supported).toContain('did');
    expect(config.id_token_types_supported).toContain('subject-signed_id_token');
  });

  it('responds with correct CORS headers', async () => {
    const response = await GET(new Request('https://example.com/.well-known/openid-configuration'));

    expect(response.headers.get('Access-Control-Allow-Origin')).toBe('*');
    expect(response.headers.get('Access-Control-Allow-Methods')).toBe('GET');
    expect(response.headers.get('Content-Type')).toBe('application/json');
  });
});
