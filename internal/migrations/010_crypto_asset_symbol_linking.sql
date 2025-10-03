-- +goose Up
-- +goose StatementBegin
-- Create a junction table to link chain assets with crypto coins by symbol
CREATE TABLE IF NOT EXISTS asset_symbol_links (
    id SERIAL PRIMARY KEY,
    chain_asset_id INTEGER REFERENCES chain_assets(id) ON DELETE CASCADE,
    crypto_coin_id VARCHAR(100) REFERENCES crypto_coins(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    match_confidence VARCHAR(20) DEFAULT 'exact', -- 'exact', 'case_insensitive', 'manual'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain_asset_id, crypto_coin_id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_asset_symbol_links_symbol ON asset_symbol_links(symbol);
CREATE INDEX IF NOT EXISTS idx_asset_symbol_links_chain_asset ON asset_symbol_links(chain_asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_symbol_links_crypto_coin ON asset_symbol_links(crypto_coin_id);
CREATE INDEX IF NOT EXISTS idx_asset_symbol_links_confidence ON asset_symbol_links(match_confidence);

-- Create a composite index on both tables' symbol columns for faster joins
CREATE INDEX IF NOT EXISTS idx_chain_assets_symbol_lower ON chain_assets(LOWER(symbol));
CREATE INDEX IF NOT EXISTS idx_crypto_coins_symbol_lower ON crypto_coins(LOWER(symbol));

-- Function to populate asset symbol links based on matching symbols
CREATE OR REPLACE FUNCTION populate_asset_symbol_links()
RETURNS TABLE(
    operation VARCHAR,
    exact_matches INTEGER,
    case_insensitive_matches INTEGER,
    total_links_created INTEGER,
    processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    exact_count INTEGER := 0;
    case_count INTEGER := 0;
BEGIN
    start_time := clock_timestamp();
    
    -- First, clear existing automatic matches (preserve manual matches)
    DELETE FROM asset_symbol_links WHERE match_confidence IN ('exact', 'case_insensitive');
    
    -- Insert exact symbol matches
    INSERT INTO asset_symbol_links (chain_asset_id, crypto_coin_id, symbol, match_confidence)
    SELECT DISTINCT
        ca.id,
        cc.id,
        ca.symbol,
        'exact'
    FROM chain_assets ca
    INNER JOIN crypto_coins cc ON ca.symbol = cc.symbol
    WHERE ca.symbol IS NOT NULL 
        AND cc.symbol IS NOT NULL
        AND cc.is_active = TRUE
    ON CONFLICT (chain_asset_id, crypto_coin_id) DO NOTHING;
    
    GET DIAGNOSTICS exact_count = ROW_COUNT;
    
    -- Insert case-insensitive matches (excluding already matched)
    INSERT INTO asset_symbol_links (chain_asset_id, crypto_coin_id, symbol, match_confidence)
    SELECT DISTINCT
        ca.id,
        cc.id,
        UPPER(ca.symbol),
        'case_insensitive'
    FROM chain_assets ca
    INNER JOIN crypto_coins cc ON LOWER(ca.symbol) = LOWER(cc.symbol)
    WHERE ca.symbol IS NOT NULL 
        AND cc.symbol IS NOT NULL
        AND cc.is_active = TRUE
        AND NOT EXISTS (
            SELECT 1 FROM asset_symbol_links asl 
            WHERE asl.chain_asset_id = ca.id 
                AND asl.crypto_coin_id = cc.id
        )
    ON CONFLICT (chain_asset_id, crypto_coin_id) DO NOTHING;
    
    GET DIAGNOSTICS case_count = ROW_COUNT;
    
    RETURN QUERY SELECT 
        'SUCCESS'::VARCHAR as operation,
        exact_count as exact_matches,
        case_count as case_insensitive_matches,
        exact_count + case_count as total_links_created,
        clock_timestamp() - start_time as processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'ERROR: ' || SQLERRM::VARCHAR as operation,
            0::INTEGER as exact_matches,
            0::INTEGER as case_insensitive_matches,
            0::INTEGER as total_links_created,
            clock_timestamp() - start_time as processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;

-- View to easily query linked assets with their market data
CREATE OR REPLACE VIEW v_linked_crypto_assets AS
SELECT 
    ca.chain_id,
    ca.asset_name,
    ca.symbol AS chain_symbol,
    ca.description AS chain_description,
    ca.logo_uri AS chain_logo,
    cc.id AS crypto_coin_id,
    cc.name AS crypto_coin_name,
    cc.symbol AS crypto_symbol,
    cc.rank AS market_rank,
    cc.type AS crypto_type,
    asl.match_confidence,
    asl.created_at AS link_created_at
FROM asset_symbol_links asl
INNER JOIN chain_assets ca ON asl.chain_asset_id = ca.id
INNER JOIN crypto_coins cc ON asl.crypto_coin_id = cc.id
ORDER BY cc.rank;

-- Function to get market data for a specific chain asset
CREATE OR REPLACE FUNCTION get_chain_asset_market_data(p_chain_id VARCHAR, p_symbol VARCHAR)
RETURNS TABLE(
    chain_id VARCHAR,
    asset_name VARCHAR,
    symbol VARCHAR,
    crypto_coin_id VARCHAR,
    crypto_coin_name VARCHAR,
    market_rank INTEGER,
    coin_type VARCHAR,
    match_confidence VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ca.chain_id,
        ca.asset_name,
        ca.symbol,
        cc.id AS crypto_coin_id,
        cc.name AS crypto_coin_name,
        cc.rank AS market_rank,
        cc.type AS coin_type,
        asl.match_confidence
    FROM chain_assets ca
    INNER JOIN asset_symbol_links asl ON ca.id = asl.chain_asset_id
    INNER JOIN crypto_coins cc ON asl.crypto_coin_id = cc.id
    WHERE ca.chain_id = p_chain_id 
        AND ca.symbol = p_symbol;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_asset_symbol_links_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_asset_symbol_links_updated_at
    BEFORE UPDATE ON asset_symbol_links
    FOR EACH ROW
    EXECUTE FUNCTION update_asset_symbol_links_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_update_asset_symbol_links_updated_at ON asset_symbol_links;
DROP FUNCTION IF EXISTS update_asset_symbol_links_updated_at();
DROP FUNCTION IF EXISTS get_chain_asset_market_data(VARCHAR, VARCHAR);
DROP VIEW IF EXISTS v_linked_crypto_assets;
DROP FUNCTION IF EXISTS populate_asset_symbol_links();
DROP INDEX IF EXISTS idx_crypto_coins_symbol_lower;
DROP INDEX IF EXISTS idx_chain_assets_symbol_lower;
DROP TABLE IF EXISTS asset_symbol_links CASCADE;
-- +goose StatementEnd