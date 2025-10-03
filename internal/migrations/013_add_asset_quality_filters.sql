-- +goose Up
-- +goose StatementBegin

-- Add status and quality fields to chain_assets table
ALTER TABLE chain_assets
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS verification_score INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Create trigger to update updated_at
CREATE TRIGGER update_chain_assets_updated_at
    BEFORE UPDATE ON chain_assets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for efficient filtering
CREATE INDEX IF NOT EXISTS idx_chain_assets_active ON chain_assets(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_chain_assets_verified ON chain_assets(is_verified) WHERE is_verified = true;
CREATE INDEX IF NOT EXISTS idx_chain_assets_score ON chain_assets(verification_score) WHERE verification_score > 0;

-- Function to calculate verification score for assets
CREATE OR REPLACE FUNCTION calculate_asset_verification_score(p_asset_id INTEGER)
RETURNS INTEGER AS $$
DECLARE
    v_score INTEGER := 0;
    v_chain_status VARCHAR(50);
    v_has_symbol BOOLEAN;
    v_has_logo BOOLEAN;
    v_has_description BOOLEAN;
    v_linked_coin_rank INTEGER;
    v_linked_coin_active BOOLEAN;
    v_match_confidence VARCHAR(20);
BEGIN
    -- Get asset details
    SELECT 
        c.status,
        ca.symbol IS NOT NULL AND ca.symbol != '',
        ca.logo_uri IS NOT NULL AND ca.logo_uri != '',
        ca.description IS NOT NULL AND ca.description != ''
    INTO v_chain_status, v_has_symbol, v_has_logo, v_has_description
    FROM chain_assets ca
    JOIN chains c ON ca.chain_id = c.chain_id
    WHERE ca.id = p_asset_id;
    
    -- Base scoring
    IF v_chain_status = 'live' THEN
        v_score := v_score + 30;
    ELSIF v_chain_status = 'upcoming' THEN
        v_score := v_score + 10;
    END IF;
    
    IF v_has_symbol THEN
        v_score := v_score + 20;
    END IF;
    
    IF v_has_logo THEN
        v_score := v_score + 10;
    END IF;
    
    IF v_has_description THEN
        v_score := v_score + 10;
    END IF;
    
    -- Check for linked market data
    SELECT 
        cc.rank,
        cc.is_active,
        asl.match_confidence
    INTO v_linked_coin_rank, v_linked_coin_active, v_match_confidence
    FROM asset_symbol_links asl
    JOIN crypto_coins cc ON asl.crypto_coin_id = cc.id
    WHERE asl.chain_asset_id = p_asset_id
    LIMIT 1;
    
    IF v_linked_coin_rank IS NOT NULL THEN
        -- Has market data
        v_score := v_score + 30;
        
        -- Bonus for high rank coins
        IF v_linked_coin_rank <= 100 THEN
            v_score := v_score + 20;
        ELSIF v_linked_coin_rank <= 500 THEN
            v_score := v_score + 10;
        END IF;
        
        -- Match confidence bonus
        IF v_match_confidence = 'exact' THEN
            v_score := v_score + 10;
        ELSIF v_match_confidence = 'manual' THEN
            v_score := v_score + 15; -- Manual verification is most trusted
        END IF;
        
        IF v_linked_coin_active THEN
            v_score := v_score + 10;
        END IF;
    END IF;
    
    RETURN v_score;
END;
$$ LANGUAGE plpgsql;

-- Function to update asset quality flags
CREATE OR REPLACE FUNCTION update_asset_quality_flags()
RETURNS TABLE(
    total_assets INTEGER,
    deactivated_count INTEGER,
    verified_count INTEGER,
    high_quality_count INTEGER
) AS $$
DECLARE
    v_total INTEGER;
    v_deactivated INTEGER := 0;
    v_verified INTEGER := 0;
    v_high_quality INTEGER := 0;
BEGIN
    -- Count total assets
    SELECT COUNT(*) INTO v_total FROM chain_assets;
    
    -- Update verification scores
    UPDATE chain_assets
    SET verification_score = calculate_asset_verification_score(id);
    
    -- Mark as inactive: assets with very low scores or from non-live chains
    UPDATE chain_assets ca
    SET is_active = false
    WHERE ca.verification_score < 20
    OR EXISTS (
        SELECT 1 FROM chains c 
        WHERE c.chain_id = ca.chain_id 
        AND c.status NOT IN ('live', 'upcoming')
    );
    
    GET DIAGNOSTICS v_deactivated = ROW_COUNT;
    
    -- Mark as verified: high quality assets with market data
    UPDATE chain_assets ca
    SET is_verified = true
    WHERE ca.verification_score >= 70
    AND EXISTS (
        SELECT 1 FROM asset_symbol_links asl
        JOIN crypto_coins cc ON asl.crypto_coin_id = cc.id
        WHERE asl.chain_asset_id = ca.id
        AND cc.is_active = true
        AND cc.rank IS NOT NULL
    );
    
    GET DIAGNOSTICS v_verified = ROW_COUNT;
    
    -- Count high quality assets
    SELECT COUNT(*) INTO v_high_quality 
    FROM chain_assets 
    WHERE verification_score >= 50;
    
    RETURN QUERY SELECT v_total, v_deactivated, v_verified, v_high_quality;
END;
$$ LANGUAGE plpgsql;

-- Update the populate_asset_symbol_links function to respect filters
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
    
    -- First, update asset quality flags
    PERFORM update_asset_quality_flags();
    
    -- Clear existing automatic matches (preserve manual matches)
    DELETE FROM asset_symbol_links WHERE match_confidence IN ('exact', 'case_insensitive');
    
    -- Insert exact symbol matches (only for active assets)
    INSERT INTO asset_symbol_links (chain_asset_id, crypto_coin_id, symbol, match_confidence)
    SELECT DISTINCT
        ca.id,
        cc.id,
        ca.symbol,
        'exact'
    FROM chain_assets ca
    INNER JOIN crypto_coins cc ON ca.symbol = cc.symbol
    INNER JOIN chains c ON ca.chain_id = c.chain_id
    WHERE ca.symbol IS NOT NULL 
        AND cc.symbol IS NOT NULL
        AND cc.is_active = TRUE
        AND ca.is_active = TRUE
        AND c.status IN ('live', 'upcoming')
        AND ca.verification_score >= 30
    ON CONFLICT (chain_asset_id, crypto_coin_id) DO NOTHING;
    
    GET DIAGNOSTICS exact_count = ROW_COUNT;
    
    -- Insert case-insensitive matches (excluding already matched, only for quality assets)
    INSERT INTO asset_symbol_links (chain_asset_id, crypto_coin_id, symbol, match_confidence)
    SELECT DISTINCT
        ca.id,
        cc.id,
        UPPER(ca.symbol),
        'case_insensitive'
    FROM chain_assets ca
    INNER JOIN crypto_coins cc ON LOWER(ca.symbol) = LOWER(cc.symbol)
    INNER JOIN chains c ON ca.chain_id = c.chain_id
    WHERE ca.symbol IS NOT NULL 
        AND cc.symbol IS NOT NULL
        AND cc.is_active = TRUE
        AND ca.is_active = TRUE
        AND c.status IN ('live', 'upcoming')
        AND ca.verification_score >= 30
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

-- Create view for quality assets only
CREATE OR REPLACE VIEW v_quality_linked_assets AS
SELECT 
    ca.chain_id,
    ca.asset_name,
    ca.symbol,
    ca.description,
    ca.logo_uri,
    ca.is_verified,
    ca.verification_score,
    cc.id as crypto_coin_id,
    cc.name as crypto_coin_name,
    cc.rank as market_rank,
    cc.type as coin_type,
    asl.match_confidence,
    cp.close as current_price,
    cp.volume as volume_24h,
    cp.market_cap,
    CASE 
        WHEN cp.open > 0 THEN ((cp.close - cp.open) / cp.open * 100)
        ELSE 0
    END as change_24h
FROM chain_assets ca
INNER JOIN asset_symbol_links asl ON ca.id = asl.chain_asset_id
INNER JOIN crypto_coins cc ON asl.crypto_coin_id = cc.id
LEFT JOIN LATERAL (
    SELECT * FROM crypto_coin_price 
    WHERE coin_id = cc.id 
    ORDER BY time_close DESC 
    LIMIT 1
) cp ON true
WHERE ca.is_active = true
    AND ca.verification_score >= 50
    AND cc.is_active = true
ORDER BY cc.rank NULLS LAST;

-- Function to manually mark an asset as junk/inactive
CREATE OR REPLACE FUNCTION mark_asset_as_junk(p_chain_id TEXT, p_base_denom TEXT)
RETURNS VOID AS $$
BEGIN
    UPDATE chain_assets
    SET 
        is_active = false,
        is_verified = false,
        verification_score = 0,
        updated_at = NOW()
    WHERE chain_id = p_chain_id AND base = p_base_denom;
    
    -- Remove any symbol links for this asset
    DELETE FROM asset_symbol_links
    WHERE chain_asset_id IN (
        SELECT id FROM chain_assets 
        WHERE chain_id = p_chain_id AND base = p_base_denom
    );
END;
$$ LANGUAGE plpgsql;

-- Function to manually verify an asset
CREATE OR REPLACE FUNCTION verify_asset(p_chain_id TEXT, p_base_denom TEXT)
RETURNS VOID AS $$
BEGIN
    UPDATE chain_assets
    SET 
        is_active = true,
        is_verified = true,
        verification_score = GREATEST(verification_score, 80),
        updated_at = NOW()
    WHERE chain_id = p_chain_id AND base = p_base_denom;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop views
DROP VIEW IF EXISTS v_quality_linked_assets;

-- Drop functions
DROP FUNCTION IF EXISTS verify_asset(TEXT, TEXT);
DROP FUNCTION IF EXISTS mark_asset_as_junk(TEXT, TEXT);
DROP FUNCTION IF EXISTS update_asset_quality_flags();
DROP FUNCTION IF EXISTS calculate_asset_verification_score(INTEGER);

-- Restore original populate_asset_symbol_links function
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

-- Drop triggers
DROP TRIGGER IF EXISTS update_chain_assets_updated_at ON chain_assets;

-- Drop indexes
DROP INDEX IF EXISTS idx_chain_assets_score;
DROP INDEX IF EXISTS idx_chain_assets_verified;
DROP INDEX IF EXISTS idx_chain_assets_active;

-- Drop columns
ALTER TABLE chain_assets
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS verification_score,
DROP COLUMN IF EXISTS is_verified,
DROP COLUMN IF EXISTS is_active;

-- +goose StatementEnd