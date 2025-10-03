-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION convert_webauthn_to_vc(
    webauthn_credential JSONB
) RETURNS JSONB AS $$
DECLARE
    vc_result JSONB;
BEGIN
    -- Create the Verifiable Credential structure based on W3C standard
    vc_result = jsonb_build_object(
        '@context', jsonb_build_array('https://www.w3.org/2018/credentials/v1'),
        'type', jsonb_build_array('VerifiableCredential', 'WebAuthnCredential'),
        'issuer', jsonb_build_object(
            'id', 'did:example:' || encode(digest(webauthn_credential->>'rpId', 'sha256'), 'hex')
        ),
        'issuanceDate', to_char(now() at time zone 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
        'credentialSubject', jsonb_build_object(
            'id', 'did:key:' || encode(digest(webauthn_credential->>'id', 'sha256'), 'hex'),
            'webauthn', webauthn_credential
        )
    );
    
    RETURN vc_result;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION convert_webauthn_to_vc;
