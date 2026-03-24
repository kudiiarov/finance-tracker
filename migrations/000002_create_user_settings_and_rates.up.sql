-- User settings table (main currency, etc.)
CREATE TABLE user_settings (
    user_id VARCHAR(255) PRIMARY KEY,
    main_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Exchange rates table (historical data)
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency VARCHAR(10) NOT NULL,  -- Source currency (BTC, EUR, etc.)
    to_currency VARCHAR(10) NOT NULL,    -- Target currency (always USD for our fetch strategy)
    rate DECIMAL(24, 12) NOT NULL,       -- Exchange rate with high precision
    source VARCHAR(50) NOT NULL,         -- 'coingecko', 'frankfurter'
    fetched_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookup of latest rate
CREATE INDEX idx_exchange_rates_latest ON exchange_rates(from_currency, to_currency, fetched_at DESC);

-- Index for historical queries
CREATE INDEX idx_exchange_rates_history ON exchange_rates(from_currency, to_currency, fetched_at);

CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default settings for hardcoded user
INSERT INTO user_settings (user_id, main_currency) VALUES ('user-123', 'USD')
ON CONFLICT (user_id) DO NOTHING;
