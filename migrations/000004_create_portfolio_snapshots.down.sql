-- Drop portfolio snapshot tables
DROP INDEX IF EXISTS idx_snapshot_accounts_currency;
DROP INDEX IF EXISTS idx_snapshot_accounts_snapshot;
DROP INDEX IF EXISTS idx_portfolio_snapshots_date;
DROP INDEX IF EXISTS idx_portfolio_snapshots_user_created;
DROP INDEX IF EXISTS idx_portfolio_snapshots_user_date;

DROP TABLE IF EXISTS portfolio_snapshot_accounts;
DROP TABLE IF EXISTS portfolio_snapshots;
