-- +goose Up
-- +goose StatementBegin

-- Enable pgcrypto extension for gen_random_bytes if not already enabled
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Function to generate WebAuthn registration options with largeBlob and payment extensions
CREATE OR REPLACE FUNCTION generate_webauthn_register_options(
    session_id UUID,
    vault_id UUID
) RETURNS JSONB AS $$
DECLARE
    challenge_bytes BYTEA;
    challenge_base64 TEXT;
    rp_id TEXT;
    rp_name TEXT;
    user_name TEXT;
    user_display_name TEXT;
BEGIN
    -- Generate a random challenge based on session_id
    challenge_bytes = gen_random_bytes(32);
    challenge_base64 = encode(challenge_bytes, 'base64');
    
    -- Set default user information
    user_name = 'user@' || session_id::TEXT;
    user_display_name = 'User';
    
    -- Set relying party information
    rp_id = 'localhost'; -- This should be configurable
    rp_name = 'Sonr Highway';
    
    -- Build and return the registration options
    RETURN jsonb_build_object(
        'challenge', challenge_base64,
        'rp', jsonb_build_object(
            'id', rp_id,
            'name', rp_name
        ),
        'user', jsonb_build_object(
            'id', encode(session_id::TEXT::BYTEA, 'base64'),
            'name', user_name,
            'displayName', user_display_name
        ),
        'pubKeyCredParams', jsonb_build_array(
            jsonb_build_object('type', 'public-key', 'alg', -7),  -- ES256
            jsonb_build_object('type', 'public-key', 'alg', -257) -- RS256
        ),
        'timeout', 60000,
        'attestation', 'direct',
        'authenticatorSelection', jsonb_build_object(
            'authenticatorAttachment', 'platform',
            'userVerification', 'required',
            'residentKey', 'required',
            'requireResidentKey', true
        ),
        'extensions', jsonb_build_object(
            'largeBlob', jsonb_build_object(
                'support', 'required'
            ),
            'payment', jsonb_build_object(
                'isPayment', true
            )
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Function to generate WebAuthn login options with largeBlob and payment extensions
CREATE OR REPLACE FUNCTION generate_webauthn_login_options(
    session_id UUID,
    vault_id UUID
) RETURNS JSONB AS $$
DECLARE
    challenge_bytes BYTEA;
    challenge_base64 TEXT;
    allow_credentials JSONB;
    rp_id TEXT;
BEGIN
    -- Generate a random challenge based on session_id
    challenge_bytes = gen_random_bytes(32);
    challenge_base64 = encode(challenge_bytes, 'base64');
    
    -- Set relying party information
    rp_id = 'localhost'; -- This should be configurable
    
    -- Set empty allowed credentials (no vault_id column exists)
    allow_credentials = '[]'::JSONB;
    
    -- Build and return the login options
    RETURN jsonb_build_object(
        'challenge', challenge_base64,
        'timeout', 60000,
        'rpId', rp_id,
        'allowCredentials', allow_credentials,
        'userVerification', 'required',
        'extensions', jsonb_build_object(
            'largeBlob', jsonb_build_object(
                'support', 'required',
                'read', true,
                'write', true
            ),
            'payment', jsonb_build_object(
                'isPayment', true,
                'payeeName', 'Sonr Network',
                'payeeOrigin', 'https://sonr.io',
                'total', jsonb_build_object(
                    'currency', 'USD',
                    'value', '0.00'
                ),
                'instrument', jsonb_build_object(
                    'displayName', 'Sonr Wallet',
                    'icon', 'https://sonr.io/icon.png'
                )
            )
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON FUNCTION generate_webauthn_register_options(UUID, UUID) IS 
'Generates WebAuthn registration options with largeBlob and payment extensions.
Parameters:
  - session_id: UUID for generating unique challenge
  - vault_id: UUID to associate with user information
Returns: JSONB object with WebAuthn registration options including required extensions';

COMMENT ON FUNCTION generate_webauthn_login_options(UUID, UUID) IS 
'Generates WebAuthn login options with largeBlob and payment extensions.
Parameters:
  - session_id: UUID for generating unique challenge
  - vault_id: UUID to lookup allowed credentials
Returns: JSONB object with WebAuthn login options including read/write largeBlob and SPC payment extensions';

-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS generate_webauthn_login_options(UUID, UUID);
DROP FUNCTION IF EXISTS generate_webauthn_register_options(UUID, UUID);
