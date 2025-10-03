-- +goose Up
-- +goose StatementBegin
-- Profiles represent user identities
CREATE TABLE profiles (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    address TEXT NOT NULL,
    handle TEXT NOT NULL,
    origin TEXT NOT NULL,
    name TEXT NOT NULL
);

-- Unique constraints
CREATE UNIQUE INDEX idx_profiles_handle_unique ON profiles(handle) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_profiles_address_origin_unique ON profiles(address, origin) WHERE deleted_at IS NULL;

-- Regular indexes for query performance
CREATE INDEX idx_profiles_address ON profiles(address);
CREATE INDEX idx_profiles_origin ON profiles(origin);
CREATE INDEX idx_profiles_name ON profiles(name);
CREATE INDEX idx_profiles_deleted_at ON profiles(deleted_at) WHERE deleted_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE profiles;
-- +goose StatementEnd

