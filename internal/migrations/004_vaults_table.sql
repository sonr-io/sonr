-- +goose Up
-- +goose StatementBegin
-- Vaults store encrypted data
CREATE TABLE vaults (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    handle TEXT NOT NULL,
    origin TEXT NOT NULL,
    address TEXT NOT NULL,
    cid TEXT NOT NULL,
    config JSONB NOT NULL,
    session_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL
);

-- Unique constraints
CREATE UNIQUE INDEX idx_vaults_cid_unique ON vaults(cid) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_vaults_handle_origin_address_unique ON vaults(handle, origin, address) WHERE deleted_at IS NULL;

-- Regular indexes for query performance
CREATE INDEX idx_vaults_handle ON vaults(handle);
CREATE INDEX idx_vaults_origin ON vaults(origin);
CREATE INDEX idx_vaults_address ON vaults(address);
CREATE INDEX idx_vaults_session_id ON vaults(session_id);
CREATE INDEX idx_vaults_deleted_at ON vaults(deleted_at) WHERE deleted_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE vaults;
-- +goose StatementEnd
