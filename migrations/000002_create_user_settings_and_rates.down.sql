DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_exchange_rates_history;
DROP INDEX IF EXISTS idx_exchange_rates_latest;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS user_settings;
