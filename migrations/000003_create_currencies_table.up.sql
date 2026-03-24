-- Create currencies table for configurable currency registry
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    precision INTEGER NOT NULL DEFAULT 2,
    currency_type VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) DEFAULT '',
    base_unit VARCHAR(20) DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    provider_prefs JSONB DEFAULT '[]',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Create index for active currencies
CREATE INDEX idx_currencies_active ON currencies(is_active);
CREATE INDEX idx_currencies_type ON currencies(currency_type);

-- Insert default currencies
INSERT INTO currencies (code, name, precision, currency_type, symbol, base_unit, is_active, provider_prefs, created_at, updated_at) VALUES
-- Fiat currencies
('USD', 'US Dollar', 2, 'fiat', '$', '', true, '[{"provider": "exchangerate-api", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('EUR', 'Euro', 2, 'fiat', '€', '', true, '[{"provider": "exchangerate-api", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('RUB', 'Russian Ruble', 2, 'fiat', '₽', '', true, '[{"provider": "exchangerate-api", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('CNY', 'Chinese Yuan', 2, 'fiat', '¥', '', true, '[{"provider": "exchangerate-api", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('GBP', 'British Pound', 2, 'fiat', '£', '', true, '[{"provider": "exchangerate-api", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

-- Cryptocurrencies
('BTC', 'Bitcoin', 8, 'crypto', '₿', '', true, '[{"provider": "coingecko", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('ETH', 'Ethereum', 8, 'crypto', 'Ξ', '', true, '[{"provider": "coingecko", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),

-- Precious metals
('GOLD', 'Gold (grams)', 3, 'metal', '🪙', 'gram', true, '[{"provider": "goldapi.io", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
('SILVER', 'Silver (grams)', 3, 'metal', '🪙', 'gram', true, '[{"provider": "goldapi.io", "priority": 1}]', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (code) DO NOTHING;
