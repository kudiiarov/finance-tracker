package domain

// PortfolioSnapshot represents user's wealth at a specific point in time.
type PortfolioSnapshot struct {
	ID                    string  `json:"id" db:"id"`
	UserID                string  `json:"user_id" db:"user_id"`
	SnapshotDate          string  `json:"snapshot_date" db:"snapshot_date"` // YYYY-MM-DD
	SnapshotTimestamp     int64   `json:"snapshot_timestamp" db:"snapshot_timestamp"`
	TotalBalanceUSD       int64   `json:"total_balance_usd" db:"total_balance_usd"`
	TotalBalanceFloat     float64 `json:"total_balance_float" db:"total_balance_float"`
	AccountCount          int     `json:"account_count" db:"account_count"`
	CurrencyCount         int     `json:"currency_count" db:"currency_count"`
	CalculationDurationMs int     `json:"calculation_duration_ms" db:"calculation_duration_ms"`
	CreatedAt             int64   `json:"created_at" db:"created_at"`

	// Joined data
	Accounts []PortfolioSnapshotAccount `json:"accounts,omitempty" db:"-"`
}

// PortfolioSnapshotAccount represents a single account's state in a snapshot.
type PortfolioSnapshotAccount struct {
	ID                   string  `json:"id" db:"id"`
	SnapshotID           string  `json:"snapshot_id" db:"snapshot_id"`
	AccountID            string  `json:"account_id" db:"account_id"`
	AccountName          string  `json:"account_name" db:"account_name"`
	AccountType          string  `json:"account_type" db:"account_type"`
	CurrencyCode         string  `json:"currency_code" db:"currency_code"`
	CurrencyPrecision    int     `json:"currency_precision" db:"currency_precision"`
	Amount               int64   `json:"amount" db:"amount"`
	AmountFloat          float64 `json:"amount_float" db:"amount_float"`
	RateToUSD            float64 `json:"rate_to_usd" db:"rate_to_usd"`
	RateSource           string  `json:"rate_source" db:"rate_source"`
	RateTimestamp        int64   `json:"rate_timestamp" db:"rate_timestamp"`
	ConvertedAmountUSD   int64   `json:"converted_amount_usd" db:"converted_amount_usd"`
	ConvertedAmountFloat float64 `json:"converted_amount_float" db:"converted_amount_float"`
	CreatedAt            int64   `json:"created_at" db:"created_at"`
}

// PortfolioHistoryQuery parameters for querying historical data.
type PortfolioHistoryQuery struct {
	UserID    string
	FromDate  string // YYYY-MM-DD
	ToDate    string // YYYY-MM-DD
	OrderDesc bool   // true = newest first
}

// PortfolioTrend represents change over a period.
type PortfolioTrend struct {
	Period          string  `json:"period"` // "7d", "30d", "1y"
	StartBalance    float64 `json:"start_balance"`
	EndBalance      float64 `json:"end_balance"`
	ChangeAmount    float64 `json:"change_amount"`
	ChangePercent   float64 `json:"change_percent"`
	DailyGrowthRate float64 `json:"daily_growth_rate"` // CAGR for the period
}

// AccountPerformance tracks individual account performance.
type AccountPerformance struct {
	AccountID       string  `json:"account_id"`
	AccountName     string  `json:"account_name"`
	CurrencyCode    string  `json:"currency_code"`
	CurrentAmount   float64 `json:"current_amount"`
	CurrentValueUSD float64 `json:"current_value_usd"`
	Change7d        float64 `json:"change_7d"`
	Change30d       float64 `json:"change_30d"`
	Change90d       float64 `json:"change_90d"`
}
