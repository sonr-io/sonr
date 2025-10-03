-- +goose Up
-- +goose StatementBegin
-- Step 1: Execute the main function to fetch and parse data
SELECT * FROM fetch_and_parse_cosmos_chains();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Clean up data from all tables
DELETE FROM chain_params;
DELETE FROM chain_assets;
DELETE FROM chain_versions;
DELETE FROM chain_explorers;
DELETE FROM chain_apis;
DELETE FROM chains;
-- +goose StatementEnd