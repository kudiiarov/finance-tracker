package handler

import (
	"net/http"
	"time"

	"github.com/sereozha/finance-tracker/internal/ports/service"
	"github.com/sereozha/finance-tracker/internal/services/statistics"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// StatisticsHandler handles portfolio statistics requests.
type StatisticsHandler struct {
	statsSvc    service.StatisticsService
	rateSvc     service.ExchangeRateService
	snapshotSvc *statistics.SnapshotService
}

// NewStatisticsHandler creates a new StatisticsHandler instance.
func NewStatisticsHandler(
	statsSvc service.StatisticsService,
	rateSvc service.ExchangeRateService,
	snapshotSvc *statistics.SnapshotService,
) *StatisticsHandler {
	return &StatisticsHandler{
		statsSvc:    statsSvc,
		rateSvc:     rateSvc,
		snapshotSvc: snapshotSvc,
	}
}

// parsePeriod converts period string to days.
func parsePeriod(period string) int {
	switch period {
	case "7d":
		return 7
	case "90d":
		return 90
	case "1y":
		return 365
	default:
		return 30
	}
}

// GetPortfolio godoc
// @Summary Get current portfolio statistics
// @Description Get real-time total balance and breakdown across all accounts
// @Tags portfolio
// @Produce json
// @Param currency query string false "Target currency for conversion (default: user's main currency)"
// @Success 200 {object} service.StatisticsResult
// @Failure 500 {object} map[string]string
// @Router /portfolio [get]
func (h *StatisticsHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.statsSvc.GetTotalBalance(ctx, contextkey.UserID(ctx), r.URL.Query().Get("currency"))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// RefreshRates godoc
// @Summary Refresh exchange rates
// @Description Trigger immediate fetch of exchange rates from all providers
// @Tags rates
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rates/refresh [post]
func (h *StatisticsHandler) RefreshRates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.rateSvc.FetchRates(ctx); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to refresh rates: "+err.Error())
		return
	}

	response.Success(w, "rates refreshed successfully")
}

// GetHistory godoc
// @Summary Get portfolio history
// @Description Get daily portfolio snapshots for a date range
// @Tags portfolio
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD, default: 30 days ago)"
// @Param to query string false "End date (YYYY-MM-DD, default: today)"
// @Success 200 {object} map[string]interface{} "{from: string, to: string, count: int, data: []PortfolioSnapshot}"
// @Failure 500 {object} map[string]string
// @Router /portfolio/history [get]
func (h *StatisticsHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	fromDate, toDate := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if fromDate == "" || toDate == "" {
		now := time.Now()
		fromDate = now.AddDate(0, 0, -30).Format("2006-01-02")
		toDate = now.Format("2006-01-02")
	}

	history, err := h.snapshotSvc.GetHistory(ctx, userID, fromDate, toDate)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"from":  fromDate,
		"to":    toDate,
		"count": len(history),
		"data":  history,
	})
}

// GetTrend godoc
// @Summary Get portfolio trend
// @Description Calculate growth metrics over a time period
// @Tags portfolio
// @Produce json
// @Param period query string false "Time period: 7d, 30d, 90d, 1y (default: 30d)"
// @Success 200 {object} domain.PortfolioTrend
// @Failure 500 {object} map[string]string
// @Router /portfolio/trend [get]
func (h *StatisticsHandler) GetTrend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	trend, err := h.snapshotSvc.GetTrend(ctx, contextkey.UserID(ctx), parsePeriod(r.URL.Query().Get("period")))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, trend)
}
