-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    number INTEGER NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0,
    address TEXT NOT NULL,
    public_key JSONB NOT NULL,
    chain_id TEXT NOT NULL,
    block_created INTEGER NOT NULL,
    controller TEXT NOT NULL,
    label TEXT NOT NULL,
    handle TEXT NOT NULL,
    is_subsidiary BOOLEAN NOT NULL DEFAULT FALSE,
    is_validator BOOLEAN NOT NULL DEFAULT FALSE,
    is_delegator BOOLEAN NOT NULL DEFAULT FALSE,
    is_accountable BOOLEAN NOT NULL DEFAULT TRUE
);


-- Unique constraints
CREATE UNIQUE INDEX idx_accounts_address_unique ON accounts(address) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_accounts_handle_controller_unique ON accounts(handle, controller) WHERE deleted_at IS NULL;

-- Regular indexes for query performance
CREATE INDEX idx_accounts_chain_id ON accounts(chain_id);
CREATE INDEX idx_accounts_block_created ON accounts(block_created);
CREATE INDEX idx_accounts_label ON accounts(label);
CREATE INDEX idx_accounts_controller ON accounts(controller);
CREATE INDEX idx_accounts_number ON accounts(number);
CREATE INDEX idx_accounts_is_subsidiary ON accounts(is_subsidiary) WHERE is_subsidiary = TRUE;
CREATE INDEX idx_accounts_is_validator ON accounts(is_validator) WHERE is_validator = TRUE;
CREATE INDEX idx_accounts_is_delegator ON accounts(is_delegator) WHERE is_delegator = TRUE;
CREATE INDEX idx_accounts_deleted_at ON accounts(deleted_at) WHERE deleted_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE accounts;
-- +goose StatementEnd

