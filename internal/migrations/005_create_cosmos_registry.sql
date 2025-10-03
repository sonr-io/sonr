-- +goose Up
-- +goose StatementBegin
-- Create the http extension required for making HTTP requests
CREATE EXTENSION IF NOT EXISTS http;

CREATE TABLE IF NOT EXISTS chains (
    chain_id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(100),
    path VARCHAR(100),
    chain_name VARCHAR(100),
    network_type VARCHAR(50),
    pretty_name VARCHAR(200),
    status VARCHAR(50),
    bech32_prefix VARCHAR(20),
    slip44 INTEGER,
    symbol VARCHAR(20),
    display VARCHAR(20),
    denom VARCHAR(50),
    decimals INTEGER,
    image VARCHAR(500),
    website VARCHAR(500),
    height BIGINT,
    proxy_status JSONB,
    keywords TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chain_apis (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(100) REFERENCES chains(chain_id) ON DELETE CASCADE,
    api_type VARCHAR(20), -- 'rest', 'rpc', 'grpc'
    address VARCHAR(500),
    provider VARCHAR(200),
    archive BOOLEAN DEFAULT FALSE,
    is_best_api BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chain_explorers (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(100) REFERENCES chains(chain_id) ON DELETE CASCADE,
    kind VARCHAR(50),
    url VARCHAR(500),
    tx_page VARCHAR(500),
    account_page VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create versions table
CREATE TABLE IF NOT EXISTS chain_versions (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(100) REFERENCES chains(chain_id) ON DELETE CASCADE,
    application_name VARCHAR(100),
    application_version VARCHAR(50),
    consensus_name VARCHAR(100),
    consensus_version VARCHAR(50),
    cosmos_sdk_version VARCHAR(50),
    tendermint_version VARCHAR(50),
    ibc_go_version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create assets table
CREATE TABLE IF NOT EXISTS chain_assets (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(100) REFERENCES chains(chain_id) ON DELETE CASCADE,
    asset_name VARCHAR(100),
    description VARCHAR(1000),
    denom_units JSONB,
    base VARCHAR(100),
    display VARCHAR(100),
    symbol VARCHAR(20),
    logo_uri VARCHAR(500),
    keywords TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Create parameters table
CREATE TABLE IF NOT EXISTS chain_params (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(100) REFERENCES chains(chain_id) ON DELETE CASCADE,
    actual_block_time NUMERIC,
    actual_blocks_per_year NUMERIC,
    bonded_ratio NUMERIC,
    inflation JSONB,
    staking_apr NUMERIC,
    unbonding_time INTEGER,
    max_validators INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_chains_status ON chains(status);
CREATE INDEX IF NOT EXISTS idx_chains_network_type ON chains(network_type);
CREATE INDEX IF NOT EXISTS idx_chain_apis_type ON chain_apis(api_type);
CREATE INDEX IF NOT EXISTS idx_chain_apis_chain_id ON chain_apis(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_explorers_chain_id ON chain_explorers(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_versions_chain_id ON chain_versions(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_assets_chain_id ON chain_assets(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_params_chain_id ON chain_params(chain_id);
-- +goose StatementEnd

-- +goose StatementBegin
-- 8. Main function to fetch and parse Cosmos chain data
CREATE OR REPLACE FUNCTION fetch_and_parse_cosmos_chains()
RETURNS TABLE(
    operation VARCHAR,
    chains_inserted INTEGER,
    apis_inserted INTEGER,
    explorers_inserted INTEGER,
    versions_inserted INTEGER,
    assets_inserted INTEGER,
    params_inserted INTEGER,
    total_processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    cosmos_data JSONB;
    chains_count INTEGER := 0;
    apis_count INTEGER := 0;
    explorers_count INTEGER := 0;
    versions_count INTEGER := 0;
    assets_count INTEGER := 0;
    params_count INTEGER := 0;
BEGIN
    start_time := clock_timestamp();
    
    -- Fetch data from the Cosmos chain directory
    SELECT content::JSONB INTO cosmos_data
    FROM http_get('https://chains.cosmos.directory');
    
    -- Clear existing data (optional - comment out if you want to preserve data)
    DELETE FROM chain_params;
    DELETE FROM chain_assets;
    DELETE FROM chain_versions;
    DELETE FROM chain_explorers;
    DELETE FROM chain_apis;
    DELETE FROM chains;
    
    -- Insert chains data
    INSERT INTO chains (
        chain_id, name, path, chain_name, network_type, pretty_name,
        status, bech32_prefix, slip44, symbol, display, denom,
        decimals, image, website, height, proxy_status, keywords
    )
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        (chain->>'name')::VARCHAR(100),
        (chain->>'path')::VARCHAR(100),
        (chain->>'chain_name')::VARCHAR(100),
        (chain->>'network_type')::VARCHAR(50),
        (chain->>'pretty_name')::VARCHAR(200),
        (chain->>'status')::VARCHAR(50),
        (chain->>'bech32_prefix')::VARCHAR(20),
        (chain->>'slip44')::INTEGER,
        (chain->>'symbol')::VARCHAR(20),
        (chain->>'display')::VARCHAR(20),
        (chain->>'denom')::VARCHAR(50),
        (chain->>'decimals')::INTEGER,
        (chain->>'image')::VARCHAR(500),
        (chain->>'website')::VARCHAR(500),
        (chain->>'height')::BIGINT,
        chain->'proxy_status',
        CASE 
            WHEN chain->'keywords' IS NOT NULL 
            THEN ARRAY(SELECT jsonb_array_elements_text(chain->'keywords'))
            ELSE NULL
        END
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain;
    
    GET DIAGNOSTICS chains_count = ROW_COUNT;
    
    -- Insert REST API endpoints
    INSERT INTO chain_apis (chain_id, api_type, address, provider, archive, is_best_api)
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        'rest'::VARCHAR(20),
        (api->>'address')::VARCHAR(500),
        (api->>'provider')::VARCHAR(200),
        COALESCE((api->>'archive')::BOOLEAN, false),
        true
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain,
         JSONB_ARRAY_ELEMENTS(chain->'best_apis'->'rest') AS api
    WHERE chain->'best_apis'->'rest' IS NOT NULL;
    
    -- Insert RPC API endpoints
    INSERT INTO chain_apis (chain_id, api_type, address, provider, archive, is_best_api)
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        'rpc'::VARCHAR(20),
        (api->>'address')::VARCHAR(500),
        (api->>'provider')::VARCHAR(200),
        COALESCE((api->>'archive')::BOOLEAN, false),
        true
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain,
         JSONB_ARRAY_ELEMENTS(chain->'best_apis'->'rpc') AS api
    WHERE chain->'best_apis'->'rpc' IS NOT NULL;
    
    -- Insert GRPC API endpoints
    INSERT INTO chain_apis (chain_id, api_type, address, provider, archive, is_best_api)
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        'grpc'::VARCHAR(20),
        (api->>'address')::VARCHAR(500),
        (api->>'provider')::VARCHAR(200),
        COALESCE((api->>'archive')::BOOLEAN, false),
        true
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain,
         JSONB_ARRAY_ELEMENTS(chain->'best_apis'->'grpc') AS api
    WHERE chain->'best_apis'->'grpc' IS NOT NULL;
    
    GET DIAGNOSTICS apis_count = ROW_COUNT;
    
    -- Insert explorers
    INSERT INTO chain_explorers (chain_id, kind, url, tx_page, account_page)
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        (explorer->>'kind')::VARCHAR(50),
        (explorer->>'url')::VARCHAR(500),
        (explorer->>'tx_page')::VARCHAR(500),
        (explorer->>'account_page')::VARCHAR(500)
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain,
         JSONB_ARRAY_ELEMENTS(chain->'explorers') AS explorer
    WHERE chain->'explorers' IS NOT NULL;
    
    GET DIAGNOSTICS explorers_count = ROW_COUNT;
    
    -- Insert version information
    INSERT INTO chain_versions (
        chain_id, application_name, application_version,
        consensus_name, consensus_version, cosmos_sdk_version,
        tendermint_version, ibc_go_version
    )
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        (chain->'versions'->>'name')::VARCHAR(100),
        (chain->'versions'->>'version')::VARCHAR(50),
        (chain->'versions'->'consensus'->>'name')::VARCHAR(100),
        (chain->'versions'->'consensus'->>'version')::VARCHAR(50),
        (chain->'versions'->>'cosmos_sdk_version')::VARCHAR(50),
        (chain->'versions'->>'tendermint_version')::VARCHAR(50),
        (chain->'versions'->>'ibc_go_version')::VARCHAR(50)
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain
    WHERE chain->'versions' IS NOT NULL;
    
    GET DIAGNOSTICS versions_count = ROW_COUNT;
    
    -- Insert assets
    INSERT INTO chain_assets (
        chain_id, asset_name, description, denom_units,
        base, display, symbol, logo_uri, keywords
    )
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        (asset->>'name')::VARCHAR(100),
        (asset->>'description')::VARCHAR(1000),
        asset->'denom_units',
        (asset->>'base')::VARCHAR(100),
        (asset->>'display')::VARCHAR(100),
        (asset->>'symbol')::VARCHAR(20),
        (asset->'logo_URIs'->>'png')::VARCHAR(500),
        CASE 
            WHEN asset->'keywords' IS NOT NULL 
            THEN ARRAY(SELECT jsonb_array_elements_text(asset->'keywords'))
            ELSE NULL
        END
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain,
         JSONB_ARRAY_ELEMENTS(chain->'assets') AS asset
    WHERE chain->'assets' IS NOT NULL;
    
    GET DIAGNOSTICS assets_count = ROW_COUNT;
    
    -- Insert parameters
    INSERT INTO chain_params (
        chain_id, actual_block_time, actual_blocks_per_year,
        bonded_ratio, inflation, staking_apr, unbonding_time, max_validators
    )
    SELECT 
        (chain->>'chain_id')::VARCHAR(100),
        (chain->'params'->>'actual_block_time')::NUMERIC,
        (chain->'params'->>'actual_blocks_per_year')::NUMERIC,
        (chain->'params'->>'bonded_ratio')::NUMERIC,
        chain->'params'->'inflation',
        (chain->'params'->>'staking_apr')::NUMERIC,
        (chain->'params'->>'unbonding_time')::INTEGER,
        (chain->'params'->>'max_validators')::INTEGER
    FROM JSONB_ARRAY_ELEMENTS(cosmos_data->'chains') AS chain
    WHERE chain->'params' IS NOT NULL;
    
    GET DIAGNOSTICS params_count = ROW_COUNT;
    
    -- Return summary
    RETURN QUERY SELECT 
        'SUCCESS'::VARCHAR as operation,
        chains_count,
        apis_count,
        explorers_count,
        versions_count,
        assets_count,
        params_count,
        clock_timestamp() - start_time as total_processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        -- Return error information
        RETURN QUERY SELECT 
            'ERROR'::VARCHAR as operation,
            0::INTEGER,
            0::INTEGER,
            0::INTEGER,
            0::INTEGER,
            0::INTEGER,
            0::INTEGER,
            clock_timestamp() - start_time as total_processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- 9. Convenience function to refresh data
CREATE OR REPLACE FUNCTION refresh_cosmos_data()
RETURNS VOID AS $$
BEGIN
    PERFORM fetch_and_parse_cosmos_chains();
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- 10. Function to get chain summary statistics
CREATE OR REPLACE FUNCTION get_cosmos_chain_stats()
RETURNS TABLE(
    total_chains INTEGER,
    live_chains INTEGER,
    testnet_chains INTEGER,
    total_apis INTEGER,
    total_explorers INTEGER,
    chains_with_staking_info INTEGER,
    avg_staking_apr NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*)::INTEGER FROM chains) as total_chains,
        (SELECT COUNT(*)::INTEGER FROM chains WHERE status = 'live') as live_chains,
        (SELECT COUNT(*)::INTEGER FROM chains WHERE network_type = 'testnet') as testnet_chains,
        (SELECT COUNT(*)::INTEGER FROM chain_apis) as total_apis,
        (SELECT COUNT(*)::INTEGER FROM chain_explorers) as total_explorers,
        (SELECT COUNT(*)::INTEGER FROM chain_params WHERE staking_apr IS NOT NULL) as chains_with_staking_info,
        (SELECT AVG(staking_apr) FROM chain_params WHERE staking_apr IS NOT NULL) as avg_staking_apr;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_cosmos_chain_stats();
DROP FUNCTION IF EXISTS refresh_cosmos_data();
DROP FUNCTION IF EXISTS fetch_and_parse_cosmos_chains();
DROP TABLE IF EXISTS chain_params CASCADE;
DROP TABLE IF EXISTS chain_assets CASCADE;
DROP TABLE IF EXISTS chain_versions CASCADE;
DROP TABLE IF EXISTS chain_explorers CASCADE;
DROP TABLE IF EXISTS chain_apis CASCADE;
DROP TABLE IF EXISTS chains CASCADE;
-- +goose StatementEnd
