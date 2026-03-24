package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sereozha/finance-tracker/internal/services/statistics"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// SnapshotHandler handles portfolio snapshot requests.
type SnapshotHandler struct {
	snapshotSvc *statistics.SnapshotService
}

// NewSnapshotHandler creates a new SnapshotHandler instance.
func NewSnapshotHandler(snapshotSvc *statistics.SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{
		snapshotSvc: snapshotSvc,
	}
}

// defaultDateRange returns the default 30-day date range.
func defaultDateRange() (from, to string) {
	now := time.Now()
	return now.AddDate(0, 0, -30).Format("2006-01-02"), now.Format("2006-01-02")
}

// List godoc
// @Summary List portfolio snapshots
// @Description Get portfolio snapshots for a date range
// @Tags snapshots
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD, default: 30 days ago)"
// @Param to query string false "End date (YYYY-MM-DD, default: today)"
// @Success 200 {object} map[string]interface{} "{from: string, to: string, count: int, data: []PortfolioSnapshot}"
// @Failure 500 {object} map[string]string
// @Router /snapshots [get]
func (h *SnapshotHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	fromDate, toDate := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	if fromDate == "" || toDate == "" {
		fromDate, toDate = defaultDateRange()
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

// GetByDate godoc
// @Summary Get snapshot by date
// @Description Get detailed portfolio state for a specific date
// @Tags snapshots
// @Produce json
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} domain.PortfolioSnapshot
// @Failure 404 {object} map[string]string
// @Router /snapshots/{date} [get]
func (h *SnapshotHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	date := chi.URLParam(r, "date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	snapshot, err := h.snapshotSvc.GetSnapshot(ctx, userID, date)
	if err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgSnapshotNotFound)
		return
	}

	response.JSON(w, http.StatusOK, snapshot)
}

// Create godoc
// @Summary Create portfolio snapshot
// @Description Create a manual portfolio snapshot for a specific date
// @Tags snapshots
// @Produce json
// @Param date query string false "Date to create snapshot for (YYYY-MM-DD, default: today)"
// @Success 201 {object} domain.PortfolioSnapshot
// @Failure 500 {object} map[string]string
// @Router /snapshots [post]
func (h *SnapshotHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	snapshot, err := h.snapshotSvc.CreateDailySnapshot(ctx, userID, date)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusCreated, snapshot)
}
