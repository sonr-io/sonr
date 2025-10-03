-- +goose Up
-- +goose StatementBegin
-- Credentials store WebAuthn credentials
CREATE TABLE credentials (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    handle TEXT NOT NULL,
    raw_id TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'public-key',
    authenticator_attachment TEXT,
    transports TEXT[], -- Array of transport strings
    client_extension_results JSONB,
    attestation_object TEXT NOT NULL, -- Base64URL encoded
    client_data_json TEXT NOT NULL, -- Base64URL encoded
    public_key_alg INTEGER,
    public_key BYTEA
);

-- Unique constraints
CREATE UNIQUE INDEX idx_credentials_id_unique ON credentials(id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_credentials_raw_id_unique ON credentials(raw_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_credentials_handle_unique ON credentials(handle) WHERE deleted_at IS NULL;

-- Regular indexes for query performance
CREATE INDEX idx_credentials_handle ON credentials(handle);
CREATE INDEX idx_credentials_type ON credentials(type);
CREATE INDEX idx_credentials_authenticator_attachment ON credentials(authenticator_attachment);
CREATE INDEX idx_credentials_deleted_at ON credentials(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE credentials;
-- +goose StatementEnd
