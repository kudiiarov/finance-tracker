# Statistics Generation System Design

## 1. Statistics Categories

### A. Wealth Statistics (Real-time)
| Metric | Description | Calculation |
|--------|-------------|-------------|
| `total_balance` | Sum of all accounts in target currency | Σ(account.amount × rate) |
| `net_worth` | Total - liabilities | Assets - Debts |
| `by_currency` | Breakdown per currency | Group by currency code |
| `by_type` | Breakdown by account type | cash, bank, crypto, etc. |
| `by_provider` | Wealth by institution/exchange | Group by account.provider |

### B. Time Series (Historical)
| Metric | Granularity | Storage Strategy |
|--------|-------------|------------------|
| `daily_balance` | Day | Pre-computed at midnight |
| `weekly_change` | Week | Computed from daily |
| `monthly_trend` | Month | Computed from daily |
| `yearly_growth` | Year | Computed from monthly |
| `all_time_high` | Event | Track peak values |

### C. Performance Metrics
| Metric | Description |
|--------|-------------|
| `best_performing_asset` | Highest % growth currency |
| `worst_performing_asset` | Lowest % growth currency |
| `volatility_index` | Standard deviation of rates |
| `correlation_matrix` | How currencies move together |
| `fiat_vs_crypto_ratio` | % allocation |

### D. Cash Flow (Transaction-based - Future)
| Metric | Description |
|--------|-------------|
| `income_vs_expenses` | Monthly cash flow |
| `savings_rate` | % of income saved |
| `top_spending_categories` | Where money goes |

---

## 2. Architecture Approaches

### Option A: Real-time Calculation (Current)
```
API Request → Query accounts → Get rates → Calculate → Return
```
**Pros:** Always up-to-date, simple
**Cons:** Slow for large datasets, repeated calculations

### Option B: Pre-computed Snapshots (Recommended)
```
Scheduler → Daily snapshot → Store in stats table → API reads from cache
```
**Pros:** Fast API response, historical tracking, trend analysis
**Cons:** Slight delay (max 24h), more storage

### Option C: Materialized Views (Hybrid)
```
PostgreSQL Materialized View → Auto-refresh → API queries view
```
**Pros:** Database handles computation, always consistent
**Cons:** Refresh overhead, limited flexibility

---

## 3. Recommended Architecture: Hybrid Approach

```
┌─────────────────────────────────────────────────────────────────┐
│                     STATISTICS SERVICE                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐   │
│  │   Real-time  │     │  Cached/     │     │  Historical  │   │
│  │   Calculator │────▶│  Snapshot    │◀────│   Analyzer   │   │
│  │              │     │   Store      │     │              │   │
│  └──────────────┘     └──────────────┘     └──────────────┘   │
│         │                    │                    │            │
│         ▼                    ▼                    ▼            │
│  ┌──────────────────────────────────────────────────────┐     │
│  │              Statistics Repository                    │     │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │     │
│  │  │ daily_stats  │  │ monthly_stats│  │snapshots  │  │     │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Data Model

### Table: `daily_statistics`
```sql
CREATE TABLE daily_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    target_currency VARCHAR(10) NOT NULL,
    
    -- Totals
    total_balance BIGINT NOT NULL,  -- in smallest units
    total_balance_float DECIMAL(24, 12),
    
    -- Breakdowns (stored as JSON for flexibility)
    by_currency JSONB,      -- {"BTC": 50000, "ETH": 3000, ...}
    by_type JSONB,          -- {"crypto": 60000, "fiat": 10000}
    
    -- Metadata
    account_count INTEGER,
    rate_count INTEGER,
    calculation_time_ms INTEGER,
    
    created_at BIGINT NOT NULL,
    
    UNIQUE(user_id, date, target_currency)
);
```

### Table: `statistics_snapshots`
```sql
CREATE TABLE statistics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(50) NOT NULL,
    timestamp BIGINT NOT NULL,
    
    -- Full snapshot data
    data JSONB NOT NULL,    -- Complete statistics object
    
    -- Index fields for filtering
    snapshot_type VARCHAR(20) NOT NULL,  -- 'hourly', 'daily', 'on_demand'
    
    created_at BIGINT NOT NULL
);
```

---

## 5. API Design

### Endpoints

```
# Current state (real-time or cached)
GET  /api/v1/statistics                    → Current totals
GET  /api/v1/statistics/breakdown          → By currency, type, etc.

# Historical
GET  /api/v1/statistics/history?from=&to=  → Time series
GET  /api/v1/statistics/trend?period=7d    → Trends (7d, 30d, 1y)
GET  /api/v1/statistics/performance        → Best/worst performers

# Comparison
GET  /api/v1/statistics/compare?currencies=BTC,ETH  → Compare assets

# Admin/Trigger
POST /admin/statistics/refresh             → Force recalculation
```

### Response Examples

```json
// GET /api/v1/statistics
{
  "timestamp": 1774289508,
  "target_currency": "USD",
  "total_balance": 157500.50,
  "change_24h": {
    "amount": 1250.00,
    "percentage": 0.80
  },
  "accounts": {
    "total": 5,
    "by_type": {
      "crypto": { "count": 2, "balance": 120000 },
      "fiat": { "count": 3, "balance": 37500 }
    }
  }
}

// GET /api/v1/statistics/history?period=30d
{
  "period": "30d",
  "data_points": 30,
  "values": [
    {"date": "2026-02-23", "balance": 145000, "change": null},
    {"date": "2026-02-24", "balance": 146200, "change": 1.20},
    ...
  ],
  "summary": {
    "start": 145000,
    "end": 157500,
    "growth": 8.62,
    "max": 158000,
    "min": 144500
  }
}
```

---

## 6. Calculation Strategies

### Real-time (For Current Stats)
```go
func (s *StatisticsService) GetCurrent(ctx, userID, targetCurrency string) (*CurrentStats, error) {
    accounts, _ := s.accountRepo.GetByUserID(ctx, userID)
    
    var total int64
    breakdown := make(map[string]float64)
    
    for _, acc := range accounts {
        rate, _ := s.rateSvc.GetRate(ctx, acc.Currency.Code, targetCurrency)
        converted := rate.Convert(acc.Amount, acc.Currency.Precision, targetPrecision)
        total += converted
        breakdown[acc.Currency.Code] += converted
    }
    
    return &CurrentStats{Total: total, Breakdown: breakdown}, nil
}
```

### Scheduled (For Historical)
```go
// Runs daily at 00:00
func (s *StatisticsService) SaveDailySnapshot(ctx, userID string) error {
    stats, _ := s.GetCurrent(ctx, userID, "USD")
    
    snapshot := DailyStatistics{
        UserID: userID,
        Date: time.Now(),
        TotalBalance: stats.Total,
        ByCurrency: stats.Breakdown,
    }
    
    return s.statsRepo.SaveDaily(ctx, snapshot)
}
```

---

## 7. Implementation Phases

### Phase 1: Basic Real-time (Current)
- ✅ Total balance calculation
- ⚠️ Needs: Breakdown by currency/type

### Phase 2: Daily Snapshots
- Daily statistics table
- Midnight scheduler job
- History API endpoint

### Phase 3: Trend Analysis
- Growth calculations
- Performance metrics
- Comparison endpoints

### Phase 4: Advanced Analytics
- Volatility tracking
- Correlation analysis
- Predictive insights

---

## 8. Questions to Decide

1. **Storage**: Store only totals or full breakdowns?
2. **Granularity**: Daily snapshots or hourly?
3. **Retention**: How long keep historical data?
4. **Multi-currency**: Snapshots in all currencies or just USD?
5. **Performance**: Cache real-time stats for X minutes?

## Recommendation

**Start with Phase 2 (Daily Snapshots)** - gives immediate value with historical tracking, enables trend graphs in UI, and provides data foundation for advanced analytics.
