-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS crypto_global_market_data (
    id SERIAL PRIMARY KEY,
    market_cap_usd NUMERIC(20, 2),
    volume_24h_usd NUMERIC(20, 2),
    bitcoin_dominance_percentage NUMERIC(5, 2),
    cryptocurrencies_number INTEGER,
    market_cap_ath_value NUMERIC(20, 2),
    market_cap_ath_date TIMESTAMP WITH TIME ZONE,
    volume_24h_ath_value NUMERIC(20, 2),
    volume_24h_ath_date TIMESTAMP WITH TIME ZONE,
    market_cap_change_24h NUMERIC(10, 2),
    volume_24h_change_24h NUMERIC(10, 2),
    last_updated INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_crypto_global_market_data_last_updated ON crypto_global_market_data(last_updated DESC);
CREATE INDEX IF NOT EXISTS idx_crypto_global_market_data_created_at ON crypto_global_market_data(created_at DESC);

CREATE TABLE IF NOT EXISTS crypto_coins (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    rank INTEGER,
    is_new BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    type VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_crypto_coins_symbol ON crypto_coins(symbol);
CREATE INDEX IF NOT EXISTS idx_crypto_coins_rank ON crypto_coins(rank);
CREATE INDEX IF NOT EXISTS idx_crypto_coins_is_active ON crypto_coins(is_active);
CREATE INDEX IF NOT EXISTS idx_crypto_coins_type ON crypto_coins(type);

CREATE TABLE IF NOT EXISTS crypto_coin_details (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    parent_id VARCHAR(100),
    parent_name VARCHAR(255),
    parent_symbol VARCHAR(20),
    rank INTEGER,
    is_new BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    type VARCHAR(50),
    logo VARCHAR(500),
    tags JSONB,
    team JSONB,
    description TEXT,
    message TEXT,
    open_source BOOLEAN,
    hardware_wallet BOOLEAN,
    started_at TIMESTAMP WITH TIME ZONE,
    development_status VARCHAR(100),
    proof_type VARCHAR(100),
    org_structure VARCHAR(100),
    hash_algorithm VARCHAR(100),
    contract VARCHAR(255),
    platform VARCHAR(255),
    contracts JSONB,
    links JSONB,
    links_extended JSONB,
    whitepaper JSONB,
    first_data_at TIMESTAMP WITH TIME ZONE,
    last_data_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_crypto_coin_details_symbol ON crypto_coin_details(symbol);
CREATE INDEX IF NOT EXISTS idx_crypto_coin_details_rank ON crypto_coin_details(rank);
CREATE INDEX IF NOT EXISTS idx_crypto_coin_details_type ON crypto_coin_details(type);
CREATE INDEX IF NOT EXISTS idx_crypto_coin_details_started_at ON crypto_coin_details(started_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fetch_coinpaprika_global_market_data()
RETURNS TABLE(
    operation VARCHAR,
    records_inserted INTEGER,
    last_market_cap NUMERIC,
    last_volume NUMERIC,
    processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    market_data JSONB;
    insert_count INTEGER := 0;
    last_cap NUMERIC;
    last_vol NUMERIC;
BEGIN
    start_time := clock_timestamp();
    SELECT content::JSONB INTO market_data
    FROM http_get('https://api.coinpaprika.com/v1/global');
    INSERT INTO crypto_global_market_data (
        market_cap_usd,
        volume_24h_usd,
        bitcoin_dominance_percentage,
        cryptocurrencies_number,
        market_cap_ath_value,
        market_cap_ath_date,
        volume_24h_ath_value,
        volume_24h_ath_date,
        market_cap_change_24h,
        volume_24h_change_24h,
        last_updated
    )
    SELECT
        (market_data->>'market_cap_usd')::NUMERIC,
        (market_data->>'volume_24h_usd')::NUMERIC,
        (market_data->>'bitcoin_dominance_percentage')::NUMERIC,
        (market_data->>'cryptocurrencies_number')::INTEGER,
        (market_data->>'market_cap_ath_value')::NUMERIC,
        (market_data->>'market_cap_ath_date')::TIMESTAMP WITH TIME ZONE,
        (market_data->>'volume_24h_ath_value')::NUMERIC,
        (market_data->>'volume_24h_ath_date')::TIMESTAMP WITH TIME ZONE,
        (market_data->>'market_cap_change_24h')::NUMERIC,
        (market_data->>'volume_24h_change_24h')::NUMERIC,
        (market_data->>'last_updated')::INTEGER
    RETURNING market_cap_usd, volume_24h_usd INTO last_cap, last_vol;
    
    GET DIAGNOSTICS insert_count = ROW_COUNT;
    
    RETURN QUERY SELECT 
        'SUCCESS'::VARCHAR as operation,
        insert_count as records_inserted,
        last_cap as last_market_cap,
        last_vol as last_volume,
        clock_timestamp() - start_time as processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'ERROR: ' || SQLERRM::VARCHAR as operation,
            0::INTEGER as records_inserted,
            0::NUMERIC as last_market_cap,
            0::NUMERIC as last_volume,
            clock_timestamp() - start_time as processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION refresh_market_data()
RETURNS VOID AS $$
BEGIN
    PERFORM fetch_coinpaprika_global_market_data();
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fetch_coinpaprika_coins()
RETURNS TABLE(
    operation VARCHAR,
    records_upserted INTEGER,
    new_coins INTEGER,
    updated_coins INTEGER,
    processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    coins_data JSONB;
    coin_record JSONB;
    upsert_count INTEGER := 0;
    new_count INTEGER := 0;
    update_count INTEGER := 0;
    existing_id VARCHAR;
BEGIN
    start_time := clock_timestamp();
    SELECT content::JSONB INTO coins_data
    FROM http_get('https://api.coinpaprika.com/v1/coins');
    FOR coin_record IN SELECT * FROM jsonb_array_elements(coins_data)
    LOOP
        SELECT id INTO existing_id FROM crypto_coins WHERE id = coin_record->>'id';
        INSERT INTO crypto_coins (
            id,
            name,
            symbol,
            rank,
            is_new,
            is_active,
            type,
            updated_at
        )
        VALUES (
            coin_record->>'id',
            coin_record->>'name',
            coin_record->>'symbol',
            (coin_record->>'rank')::INTEGER,
            (coin_record->>'is_new')::BOOLEAN,
            (coin_record->>'is_active')::BOOLEAN,
            coin_record->>'type',
            CURRENT_TIMESTAMP
        )
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            symbol = EXCLUDED.symbol,
            rank = EXCLUDED.rank,
            is_new = EXCLUDED.is_new,
            is_active = EXCLUDED.is_active,
            type = EXCLUDED.type,
            updated_at = CURRENT_TIMESTAMP;
        
        upsert_count := upsert_count + 1;
        
        IF existing_id IS NULL THEN
            new_count := new_count + 1;
        ELSE
            update_count := update_count + 1;
        END IF;
    END LOOP;
    RETURN QUERY SELECT 
        'SUCCESS'::VARCHAR as operation,
        upsert_count as records_upserted,
        new_count as new_coins,
        update_count as updated_coins,
        clock_timestamp() - start_time as processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'ERROR: ' || SQLERRM::VARCHAR as operation,
            0::INTEGER as records_upserted,
            0::INTEGER as new_coins,
            0::INTEGER as updated_coins,
            clock_timestamp() - start_time as processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION refresh_all_coinpaprika_data()
RETURNS TABLE(
    operation VARCHAR,
    market_data_result VARCHAR,
    coins_result VARCHAR,
    total_processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    market_result RECORD;
    coins_result RECORD;
BEGIN
    start_time := clock_timestamp();
    
    -- Fetch market data
    SELECT * INTO market_result FROM fetch_coinpaprika_global_market_data();
    
    -- Fetch coins data
    SELECT * INTO coins_result FROM fetch_coinpaprika_coins();
    
    -- Return combined results
    RETURN QUERY SELECT 
        'COMPLETE'::VARCHAR as operation,
        market_result.operation as market_data_result,
        coins_result.operation as coins_result,
        clock_timestamp() - start_time as total_processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT 
            'ERROR'::VARCHAR as operation,
            'ERROR: ' || SQLERRM::VARCHAR as market_data_result,
            'ERROR: ' || SQLERRM::VARCHAR as coins_result,
            clock_timestamp() - start_time as total_processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Function to fetch and store detailed coin information
CREATE OR REPLACE FUNCTION fetch_coinpaprika_coin_details(coin_id VARCHAR)
RETURNS TABLE(
    operation VARCHAR,
    coin_name VARCHAR,
    coin_symbol VARCHAR,
    processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    coin_data JSONB;
    parent_data JSONB;
BEGIN
    start_time := clock_timestamp();
    
    -- Fetch data from CoinPaprika coin details endpoint
    SELECT content::JSONB INTO coin_data
    FROM http_get('https://api.coinpaprika.com/v1/coins/' || coin_id);
    
    -- Extract parent data if exists
    parent_data := coin_data->'parent';
    
    -- Insert or update the coin details
    INSERT INTO crypto_coin_details (
        id,
        name,
        symbol,
        parent_id,
        parent_name,
        parent_symbol,
        rank,
        is_new,
        is_active,
        type,
        logo,
        tags,
        team,
        description,
        message,
        open_source,
        hardware_wallet,
        started_at,
        development_status,
        proof_type,
        org_structure,
        hash_algorithm,
        contract,
        platform,
        contracts,
        links,
        links_extended,
        whitepaper,
        first_data_at,
        last_data_at,
        updated_at
    )
    VALUES (
        coin_data->>'id',
        coin_data->>'name',
        coin_data->>'symbol',
        CASE WHEN parent_data IS NOT NULL THEN parent_data->>'id' ELSE NULL END,
        CASE WHEN parent_data IS NOT NULL THEN parent_data->>'name' ELSE NULL END,
        CASE WHEN parent_data IS NOT NULL THEN parent_data->>'symbol' ELSE NULL END,
        (coin_data->>'rank')::INTEGER,
        (coin_data->>'is_new')::BOOLEAN,
        (coin_data->>'is_active')::BOOLEAN,
        coin_data->>'type',
        coin_data->>'logo',
        coin_data->'tags',
        coin_data->'team',
        coin_data->>'description',
        coin_data->>'message',
        (coin_data->>'open_source')::BOOLEAN,
        (coin_data->>'hardware_wallet')::BOOLEAN,
        (coin_data->>'started_at')::TIMESTAMP WITH TIME ZONE,
        coin_data->>'development_status',
        coin_data->>'proof_type',
        coin_data->>'org_structure',
        coin_data->>'hash_algorithm',
        coin_data->>'contract',
        coin_data->>'platform',
        coin_data->'contracts',
        coin_data->'links',
        coin_data->'links_extended',
        coin_data->'whitepaper',
        (coin_data->>'first_data_at')::TIMESTAMP WITH TIME ZONE,
        (coin_data->>'last_data_at')::TIMESTAMP WITH TIME ZONE,
        CURRENT_TIMESTAMP
    )
    ON CONFLICT (id) DO UPDATE SET
        name = EXCLUDED.name,
        symbol = EXCLUDED.symbol,
        parent_id = EXCLUDED.parent_id,
        parent_name = EXCLUDED.parent_name,
        parent_symbol = EXCLUDED.parent_symbol,
        rank = EXCLUDED.rank,
        is_new = EXCLUDED.is_new,
        is_active = EXCLUDED.is_active,
        type = EXCLUDED.type,
        logo = EXCLUDED.logo,
        tags = EXCLUDED.tags,
        team = EXCLUDED.team,
        description = EXCLUDED.description,
        message = EXCLUDED.message,
        open_source = EXCLUDED.open_source,
        hardware_wallet = EXCLUDED.hardware_wallet,
        started_at = EXCLUDED.started_at,
        development_status = EXCLUDED.development_status,
        proof_type = EXCLUDED.proof_type,
        org_structure = EXCLUDED.org_structure,
        hash_algorithm = EXCLUDED.hash_algorithm,
        contract = EXCLUDED.contract,
        platform = EXCLUDED.platform,
        contracts = EXCLUDED.contracts,
        links = EXCLUDED.links,
        links_extended = EXCLUDED.links_extended,
        whitepaper = EXCLUDED.whitepaper,
        first_data_at = EXCLUDED.first_data_at,
        last_data_at = EXCLUDED.last_data_at,
        updated_at = CURRENT_TIMESTAMP;
    
    -- Return summary
    RETURN QUERY SELECT 
        'SUCCESS'::VARCHAR as operation,
        coin_data->>'name' as coin_name,
        coin_data->>'symbol' as coin_symbol,
        clock_timestamp() - start_time as processing_time;
        
EXCEPTION
    WHEN OTHERS THEN
        -- Return error information
        RETURN QUERY SELECT 
            'ERROR: ' || SQLERRM::VARCHAR as operation,
            NULL::VARCHAR as coin_name,
            NULL::VARCHAR as coin_symbol,
            clock_timestamp() - start_time as processing_time;
        RAISE;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Function to fetch details for top N coins
CREATE OR REPLACE FUNCTION fetch_top_coins_details(limit_count INTEGER DEFAULT 10)
RETURNS TABLE(
    operation VARCHAR,
    coins_processed INTEGER,
    successful_fetches INTEGER,
    failed_fetches INTEGER,
    processing_time INTERVAL
) AS $$
DECLARE
    start_time TIMESTAMP;
    coin_rec RECORD;
    success_count INTEGER := 0;
    fail_count INTEGER := 0;
    total_count INTEGER := 0;
    fetch_result RECORD;
BEGIN
    start_time := clock_timestamp();
    
    -- Fetch details for each top coin
    FOR coin_rec IN 
        SELECT id FROM crypto_coins 
        WHERE is_active = TRUE AND rank IS NOT NULL 
        ORDER BY rank 
        LIMIT limit_count
    LOOP
        total_count := total_count + 1;
        
        -- Try to fetch coin details
        BEGIN
            SELECT * INTO fetch_result FROM fetch_coinpaprika_coin_details(coin_rec.id);
            IF fetch_result.operation = 'SUCCESS' THEN
                success_count := success_count + 1;
            ELSE
                fail_count := fail_count + 1;
            END IF;
        EXCEPTION
            WHEN OTHERS THEN
                fail_count := fail_count + 1;
        END;
    END LOOP;
    
    -- Return summary
    RETURN QUERY SELECT 
        'COMPLETE'::VARCHAR as operation,
        total_count as coins_processed,
        success_count as successful_fetches,
        fail_count as failed_fetches,
        clock_timestamp() - start_time as processing_time;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS fetch_top_coins_details(INTEGER);
DROP FUNCTION IF EXISTS fetch_coinpaprika_coin_details(VARCHAR);
DROP FUNCTION IF EXISTS refresh_all_coinpaprika_data();
DROP FUNCTION IF EXISTS fetch_coinpaprika_coins();
DROP FUNCTION IF EXISTS refresh_market_data();
DROP FUNCTION IF EXISTS fetch_coinpaprika_global_market_data();
DROP TABLE IF EXISTS crypto_coin_details CASCADE;
DROP TABLE IF EXISTS crypto_coins CASCADE;
DROP TABLE IF EXISTS crypto_global_market_data CASCADE;
-- +goose StatementEnd
