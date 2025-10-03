-- +goose Up
-- +goose StatementBegin

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable pg_cron extension for scheduled jobs
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Grant usage on pg_cron schema to postgres user
GRANT USAGE ON SCHEMA cron TO postgres;

-- Create crypto_coin_price table for storing historical price data
CREATE TABLE IF NOT EXISTS crypto_coin_price (
    id SERIAL,
    coin_id VARCHAR(255) NOT NULL REFERENCES crypto_coins(id) ON DELETE CASCADE,
    time_open TIMESTAMPTZ NOT NULL,
    time_close TIMESTAMPTZ NOT NULL,
    open DECIMAL(20, 8),
    high DECIMAL(20, 8),
    low DECIMAL(20, 8),
    close DECIMAL(20, 8),
    volume DECIMAL(30, 2),
    market_cap DECIMAL(30, 2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(coin_id, time_open, time_close)
);

-- Convert to hypertable partitioned by time_close
SELECT create_hypertable('crypto_coin_price', 'time_close', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- Create indexes optimized for TimescaleDB
CREATE INDEX idx_crypto_coin_price_time_close_coin_id 
ON crypto_coin_price (time_close DESC, coin_id) 
WHERE time_close IS NOT NULL;

CREATE INDEX idx_crypto_coin_price_coin_id_time_close 
ON crypto_coin_price (coin_id, time_close DESC) 
WHERE coin_id IS NOT NULL;

-- Enable compression on older chunks
ALTER TABLE crypto_coin_price SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'coin_id',
    timescaledb.compress_orderby = 'time_close DESC'
);

-- Create a policy to automatically compress chunks older than specified interval
-- Default is 30 days, but can be configured via app.timescaledb_compress_after_days
DO $$
DECLARE
    compress_days INTEGER;
    compress_interval INTERVAL;
BEGIN
    -- Try to get compression days from config, default to 30 days
    compress_days := COALESCE(
        current_setting('app.timescaledb_compress_after_days', true)::INTEGER,
        30
    );
    
    compress_interval := compress_days || ' days';
    
    PERFORM add_compression_policy('crypto_coin_price', compress_interval);
    
    RAISE NOTICE 'Compression policy set to compress chunks older than % days', compress_days;
END $$;

-- Add retention policy to drop chunks older than specified interval
-- Default is 1 year, but can be configured via app.timescaledb_retention_days
DO $$
DECLARE
    retention_days INTEGER;
    retention_interval INTERVAL;
BEGIN
    -- Try to get retention days from config, default to 365 days (1 year)
    retention_days := COALESCE(
        current_setting('app.timescaledb_retention_days', true)::INTEGER,
        365
    );
    
    retention_interval := retention_days || ' days';
    
    PERFORM add_retention_policy('crypto_coin_price', retention_interval);
    
    RAISE NOTICE 'Retention policy set to % days', retention_days;
END $$;

-- Create update trigger for updated_at
CREATE TRIGGER update_crypto_coin_price_updated_at
    BEFORE UPDATE ON crypto_coin_price
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create HTTP function to fetch OHLCV data from CoinPaprika
CREATE OR REPLACE FUNCTION http_get_coinpaprika_ohlcv(coin_id TEXT, start_date TEXT DEFAULT NULL, end_date TEXT DEFAULT NULL)
RETURNS TABLE (
    time_open TIMESTAMPTZ,
    time_close TIMESTAMPTZ,
    open NUMERIC,
    high NUMERIC,
    low NUMERIC,
    close NUMERIC,
    volume NUMERIC,
    market_cap NUMERIC
) AS $$
DECLARE
    api_url TEXT;
    response JSON;
    api_key TEXT;
BEGIN
    -- Get API key from environment or use default
    api_key := COALESCE(current_setting('app.coinpaprika_api_key', true), 'dummy-key-for-free-tier');
    
    -- Build URL based on parameters
    IF start_date IS NULL AND end_date IS NULL THEN
        -- Get today's data
        api_url := 'https://api.coinpaprika.com/v1/coins/' || coin_id || '/ohlcv/today';
    ELSIF start_date IS NOT NULL AND end_date IS NOT NULL THEN
        -- Get historical data for date range
        api_url := 'https://api.coinpaprika.com/v1/coins/' || coin_id || '/ohlcv/historical?start=' || start_date || '&end=' || end_date;
    ELSE
        -- Get latest data (last 24 hours)
        api_url := 'https://api.coinpaprika.com/v1/coins/' || coin_id || '/ohlcv/latest';
    END IF;
    
    -- Make HTTP request
    response := http_get(
        api_url,
        JSONB_BUILD_OBJECT(
            'User-Agent', 'PostgreSQL/HTTP',
            'Accept', 'application/json'
        )
    )::JSON;
    
    -- Parse and return the response
    RETURN QUERY
    SELECT 
        (item->>'time_open')::TIMESTAMPTZ,
        (item->>'time_close')::TIMESTAMPTZ,
        (item->>'open')::NUMERIC,
        (item->>'high')::NUMERIC,
        (item->>'low')::NUMERIC,
        (item->>'close')::NUMERIC,
        (item->>'volume')::NUMERIC,
        (item->>'market_cap')::NUMERIC
    FROM json_array_elements(response) AS item;
END;
$$ LANGUAGE plpgsql;

-- Function to fetch and store today's price data for a coin
CREATE OR REPLACE FUNCTION refresh_coin_price_today(p_coin_id TEXT)
RETURNS VOID AS $$
BEGIN
    -- Insert or update today's price data
    INSERT INTO crypto_coin_price (
        coin_id, time_open, time_close, open, high, low, close, volume, market_cap
    )
    SELECT 
        p_coin_id, time_open, time_close, open, high, low, close, volume, market_cap
    FROM http_get_coinpaprika_ohlcv(p_coin_id)
    ON CONFLICT (coin_id, time_open, time_close) 
    DO UPDATE SET
        open = EXCLUDED.open,
        high = EXCLUDED.high,
        low = EXCLUDED.low,
        close = EXCLUDED.close,
        volume = EXCLUDED.volume,
        market_cap = EXCLUDED.market_cap,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Create continuous aggregates for different time intervals
-- 1 hour aggregates
CREATE MATERIALIZED VIEW crypto_coin_price_1h
WITH (timescaledb.continuous) AS
SELECT 
    coin_id,
    time_bucket('1 hour', time_close) AS bucket,
    FIRST(open, time_open) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, time_close) AS close,
    SUM(volume) AS volume,
    AVG(market_cap) AS market_cap,
    COUNT(*) AS num_records
FROM crypto_coin_price
GROUP BY coin_id, bucket
WITH NO DATA;

-- 4 hour aggregates
CREATE MATERIALIZED VIEW crypto_coin_price_4h
WITH (timescaledb.continuous) AS
SELECT 
    coin_id,
    time_bucket('4 hours', time_close) AS bucket,
    FIRST(open, time_open) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, time_close) AS close,
    SUM(volume) AS volume,
    AVG(market_cap) AS market_cap,
    COUNT(*) AS num_records
FROM crypto_coin_price
GROUP BY coin_id, bucket
WITH NO DATA;

-- 1 day aggregates
CREATE MATERIALIZED VIEW crypto_coin_price_1d
WITH (timescaledb.continuous) AS
SELECT 
    coin_id,
    time_bucket('1 day', time_close) AS bucket,
    FIRST(open, time_open) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, time_close) AS close,
    SUM(volume) AS volume,
    AVG(market_cap) AS market_cap,
    COUNT(*) AS num_records
FROM crypto_coin_price
GROUP BY coin_id, bucket
WITH NO DATA;

-- 1 week aggregates
CREATE MATERIALIZED VIEW crypto_coin_price_1w
WITH (timescaledb.continuous) AS
SELECT 
    coin_id,
    time_bucket('1 week', time_close) AS bucket,
    FIRST(open, time_open) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, time_close) AS close,
    SUM(volume) AS volume,
    AVG(market_cap) AS market_cap,
    COUNT(*) AS num_records
FROM crypto_coin_price
GROUP BY coin_id, bucket
WITH NO DATA;

-- Create refresh policies for continuous aggregates
-- The refresh window must be larger than the bucket interval
-- For 1h buckets: window must be > 1 hour
SELECT add_continuous_aggregate_policy('crypto_coin_price_1h',
    start_offset => INTERVAL '4 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '30 minutes');

-- For 4h buckets: window must be > 4 hours  
SELECT add_continuous_aggregate_policy('crypto_coin_price_4h',
    start_offset => INTERVAL '24 hours',
    end_offset => INTERVAL '4 hours',
    schedule_interval => INTERVAL '2 hours');

-- For 1d buckets: window must be > 1 day
SELECT add_continuous_aggregate_policy('crypto_coin_price_1d',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '6 hours');

-- For 1w buckets: window must be > 1 week
SELECT add_continuous_aggregate_policy('crypto_coin_price_1w',
    start_offset => INTERVAL '4 weeks',
    end_offset => INTERVAL '1 week',
    schedule_interval => INTERVAL '1 day');

-- Create indexes on continuous aggregates
CREATE INDEX idx_crypto_coin_price_1h_coin_bucket ON crypto_coin_price_1h(coin_id, bucket DESC);
CREATE INDEX idx_crypto_coin_price_4h_coin_bucket ON crypto_coin_price_4h(coin_id, bucket DESC);
CREATE INDEX idx_crypto_coin_price_1d_coin_bucket ON crypto_coin_price_1d(coin_id, bucket DESC);
CREATE INDEX idx_crypto_coin_price_1w_coin_bucket ON crypto_coin_price_1w(coin_id, bucket DESC);

-- Function to get latest price for a coin
CREATE OR REPLACE FUNCTION get_latest_coin_price(p_coin_id TEXT)
RETURNS TABLE (
    coin_id VARCHAR(255),
    price DECIMAL(20, 8),
    volume DECIMAL(30, 2),
    market_cap DECIMAL(30, 2),
    high_24h DECIMAL(20, 8),
    low_24h DECIMAL(20, 8),
    change_24h DECIMAL(20, 8),
    last_updated TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cp.coin_id,
        cp.close as price,
        cp.volume,
        cp.market_cap,
        cp.high as high_24h,
        cp.low as low_24h,
        CASE 
            WHEN cp.open > 0 THEN ((cp.close - cp.open) / cp.open * 100)
            ELSE 0
        END as change_24h,
        cp.time_close as last_updated
    FROM crypto_coin_price cp
    WHERE cp.coin_id = p_coin_id
    ORDER BY cp.time_close DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to update TimescaleDB policies dynamically
CREATE OR REPLACE FUNCTION update_timescaledb_policies(
    p_retention_days INTEGER DEFAULT NULL,
    p_compress_after_days INTEGER DEFAULT NULL
)
RETURNS TABLE (
    policy_type TEXT,
    old_value TEXT,
    new_value TEXT,
    status TEXT
) AS $$
DECLARE
    current_retention INTERVAL;
    current_compression INTERVAL;
    new_retention INTERVAL;
    new_compression INTERVAL;
BEGIN
    -- Update retention policy if provided
    IF p_retention_days IS NOT NULL THEN
        -- Get current retention policy
        SELECT schedule_interval INTO current_retention
        FROM timescaledb_information.policies
        WHERE hypertable = 'crypto_coin_price'::regclass
        AND policy_name LIKE '%retention%'
        LIMIT 1;
        
        new_retention := p_retention_days || ' days';
        
        -- Remove old policy and add new one
        PERFORM remove_retention_policy('crypto_coin_price', if_exists => TRUE);
        PERFORM add_retention_policy('crypto_coin_price', new_retention);
        
        RETURN QUERY SELECT 
            'retention'::TEXT,
            COALESCE(current_retention::TEXT, 'none'),
            new_retention::TEXT,
            'updated'::TEXT;
    END IF;
    
    -- Update compression policy if provided
    IF p_compress_after_days IS NOT NULL THEN
        -- Get current compression policy
        SELECT schedule_interval INTO current_compression
        FROM timescaledb_information.policies
        WHERE hypertable = 'crypto_coin_price'::regclass
        AND policy_name LIKE '%compress%'
        LIMIT 1;
        
        new_compression := p_compress_after_days || ' days';
        
        -- Remove old policy and add new one
        PERFORM remove_compression_policy('crypto_coin_price', if_exists => TRUE);
        PERFORM add_compression_policy('crypto_coin_price', new_compression);
        
        RETURN QUERY SELECT 
            'compression'::TEXT,
            COALESCE(current_compression::TEXT, 'none'),
            new_compression::TEXT,
            'updated'::TEXT;
    END IF;
    
    -- Save configuration to database settings
    IF p_retention_days IS NOT NULL THEN
        PERFORM set_config('app.timescaledb_retention_days', p_retention_days::TEXT, FALSE);
    END IF;
    
    IF p_compress_after_days IS NOT NULL THEN
        PERFORM set_config('app.timescaledb_compress_after_days', p_compress_after_days::TEXT, FALSE);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to refresh price data for all active coins
CREATE OR REPLACE FUNCTION refresh_all_coin_prices()
RETURNS VOID AS $$
DECLARE
    coin RECORD;
    success_count INT := 0;
    error_count INT := 0;
BEGIN
    -- Iterate through all coins that have been linked to Cosmos assets
    FOR coin IN 
        SELECT DISTINCT cc.id as coin_id
        FROM crypto_coins cc
        JOIN asset_symbol_links asl ON asl.crypto_coin_id = cc.id
        WHERE cc.is_active = TRUE
        LIMIT 100  -- Process in batches to avoid overwhelming the API
    LOOP
        BEGIN
            PERFORM refresh_coin_price_today(coin.coin_id);
            success_count := success_count + 1;
        EXCEPTION WHEN OTHERS THEN
            error_count := error_count + 1;
            RAISE NOTICE 'Error refreshing price for coin %: %', coin.coin_id, SQLERRM;
        END;
    END LOOP;
    
    RAISE NOTICE 'Price refresh completed. Success: %, Errors: %', success_count, error_count;
END;
$$ LANGUAGE plpgsql;

-- Create pg_cron jobs for automated data fetching
-- Refresh prices every 5 minutes during market hours
SELECT cron.schedule(
    'refresh-coin-prices-frequent',
    '*/5 * * * *',  -- Every 5 minutes
    $$SELECT refresh_all_coin_prices();$$
);

-- Refresh continuous aggregates every hour
SELECT cron.schedule(
    'refresh-continuous-aggregates-1h',
    '0 * * * *',  -- Every hour at minute 0
    $$CALL refresh_continuous_aggregate('crypto_coin_price_1h', NULL, NULL);$$
);

-- Function to get cosmos asset price through symbol linking
CREATE OR REPLACE FUNCTION get_cosmos_asset_price(p_chain_name TEXT, p_base_denom TEXT)
RETURNS TABLE (
    chain_name TEXT,
    base_denom TEXT,
    symbol TEXT,
    coin_id VARCHAR(255),
    price DECIMAL(20, 8),
    volume DECIMAL(30, 2),
    market_cap DECIMAL(30, 2),
    high_24h DECIMAL(20, 8),
    low_24h DECIMAL(20, 8),
    change_24h DECIMAL(20, 8),
    last_updated TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ca.chain_id as chain_name,
        ca.base as base_denom,
        asl.symbol,
        lcp.coin_id,
        lcp.price,
        lcp.volume,
        lcp.market_cap,
        lcp.high_24h,
        lcp.low_24h,
        lcp.change_24h,
        lcp.last_updated
    FROM asset_symbol_links asl
    JOIN chain_assets ca ON asl.chain_asset_id = ca.id
    JOIN LATERAL get_latest_coin_price(asl.crypto_coin_id) lcp ON TRUE
    WHERE ca.chain_id = p_chain_name 
    AND ca.base = p_base_denom;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop cron jobs
SELECT cron.unschedule('refresh-coin-prices-frequent');
SELECT cron.unschedule('refresh-continuous-aggregates-1h');

-- Drop functions
DROP FUNCTION IF EXISTS get_cosmos_asset_price(TEXT, TEXT);
DROP FUNCTION IF EXISTS get_latest_coin_price(TEXT);
DROP FUNCTION IF EXISTS update_timescaledb_policies(INTEGER, INTEGER);
DROP FUNCTION IF EXISTS refresh_all_coin_prices();
DROP FUNCTION IF EXISTS refresh_coin_price_today(TEXT);
DROP FUNCTION IF EXISTS http_get_coinpaprika_ohlcv(TEXT, TEXT, TEXT);

-- Drop continuous aggregate policies
SELECT remove_continuous_aggregate_policy('crypto_coin_price_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('crypto_coin_price_4h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('crypto_coin_price_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('crypto_coin_price_1w', if_exists => TRUE);

-- Drop continuous aggregates
DROP MATERIALIZED VIEW IF EXISTS crypto_coin_price_1h CASCADE;
DROP MATERIALIZED VIEW IF EXISTS crypto_coin_price_4h CASCADE;
DROP MATERIALIZED VIEW IF EXISTS crypto_coin_price_1d CASCADE;
DROP MATERIALIZED VIEW IF EXISTS crypto_coin_price_1w CASCADE;

-- Drop TimescaleDB policies
SELECT remove_retention_policy('crypto_coin_price', if_exists => TRUE);
SELECT remove_compression_policy('crypto_coin_price', if_exists => TRUE);

-- Drop triggers
DROP TRIGGER IF EXISTS update_crypto_coin_price_updated_at ON crypto_coin_price;

-- Drop hypertable (this will drop all associated objects)
DROP TABLE IF EXISTS crypto_coin_price CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS pg_cron CASCADE;
DROP EXTENSION IF EXISTS timescaledb CASCADE;

-- +goose StatementEnd