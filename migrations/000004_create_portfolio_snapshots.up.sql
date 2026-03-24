-- Portfolio snapshots for historical wealth tracking
-- Stores daily account states with rates used for conversion

CREATE TABLE portfolio_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(50) NOT NULL,
    snapshot_date DATE NOT NULL,
    snapshot_timestamp BIGINT NOT NULL,  -- Unix timestamp when snapshot was taken
    
    -- Totals in USD (using rates from snapshot_date)
    total_balance_usd BIGINT NOT NULL,  -- Smallest units (cents)
    total_balance_float DECIMAL(24, 12),  -- Human readable
    
    -- Metadata
    account_count INTEGER NOT NULL DEFAULT 0,
    currency_count INTEGER NOT NULL DEFAULT 0,
    calculation_duration_ms INTEGER,  -- How long calculation took
    
    created_at BIGINT NOT NULL,
    
    UNIQUE(user_id, snapshot_date)
);

-- Individual account states within each snapshot
CREATE TABLE portfolio_snapshot_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL REFERENCES portfolio_snapshots(id) ON DELETE CASCADE,
    
    -- Account reference (we store name in case account is deleted later)
    account_id UUID NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    
    -- Currency info
    currency_code VARCHAR(10) NOT NULL,
    currency_precision INTEGER NOT NULL,
    
    -- Amount at snapshot time
    amount BIGINT NOT NULL,  -- In currency's smallest units
    amount_float DECIMAL(24, 12),  -- Human readable
    
    -- Conversion to USD
    rate_to_usd DECIMAL(24, 12) NOT NULL,  -- Rate used for conversion
    rate_source VARCHAR(50) NOT NULL,  -- Which provider
    rate_timestamp BIGINT NOT NULL,  -- When that rate was fetched
    converted_amount_usd BIGINT NOT NULL,  -- amount * rate in USD cents
    converted_amount_float DECIMAL(24, 12),
    
    created_at BIGINT NOT NULL
);

-- Indexes for efficient querying
CREATE INDEX idx_portfolio_snapshots_user_date ON portfolio_snapshots(user_id, snapshot_date);
CREATE INDEX idx_portfolio_snapshots_user_created ON portfolio_snapshots(user_id, created_at DESC);
CREATE INDEX idx_portfolio_snapshots_date ON portfolio_snapshots(snapshot_date);

CREATE INDEX idx_snapshot_accounts_snapshot ON portfolio_snapshot_accounts(snapshot_id);
CREATE INDEX idx_snapshot_accounts_currency ON portfolio_snapshot_accounts(currency_code);

-- Materialized view for quick stats (optional, can be added later)
-- CREATE MATERIALIZED VIEW portfolio_daily_stats AS ...
