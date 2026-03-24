package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sereozha/finance-tracker/internal/adapters/http/dto"
	"github.com/sereozha/finance-tracker/internal/ports/service"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// SettingsHandler handles user settings requests.
type SettingsHandler struct {
	service service.SettingsService
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(service service.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

// Get godoc
// @Summary Get user settings
// @Description Get current user settings including main currency preference
// @Tags settings
// @Produce json
// @Success 200 {object} domain.Settings
// @Failure 500 {object} map[string]string
// @Router /settings [get]
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s, err := h.service.Get(ctx, contextkey.UserID(ctx))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, s)
}

// Update godoc
// @Summary Update user settings
// @Description Update all user settings (full replacement)
// @Tags settings
// @Accept json
// @Produce json
// @Param request body dto.UpdateSettingsRequest true "Settings data"
// @Success 200 {object} domain.Settings
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings [put]
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	var req dto.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	// Validate currency
	if req.MainCurrency == "" {
		response.Error(w, http.StatusBadRequest, "main_currency is required")
		return
	}

	if err := h.service.Update(ctx, userID, req.MainCurrency); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return updated settings
	s, err := h.service.Get(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, s)
}
