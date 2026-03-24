package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sereozha/finance-tracker/internal/adapters/http/dto"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// AdminCurrencyHandler handles admin currency management.
type AdminCurrencyHandler struct {
	currencyRepo repository.CurrencyRepository
}

// NewAdminCurrencyHandler creates a new admin currency handler.
func NewAdminCurrencyHandler(currencyRepo repository.CurrencyRepository) *AdminCurrencyHandler {
	return &AdminCurrencyHandler{currencyRepo: currencyRepo}
}

// parseCurrencyCode extracts and validates currency code from URL.
func (h *AdminCurrencyHandler) parseCurrencyCode(r *http.Request) string {
	return chi.URLParam(r, "code")
}

// List godoc
// @Summary List all currencies (admin)
// @Description Get detailed list of all currencies including provider preferences
// @Tags admin
// @Produce json
// @Success 200 {array} dto.CurrencyInfoResponse
// @Router /admin/currencies [get]
func (h *AdminCurrencyHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currencies, err := h.currencyRepo.ListActive(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, dto.ToCurrencyInfoList(currencies))
}

// Get godoc
// @Summary Get currency details
// @Description Get detailed information about a specific currency
// @Tags admin
// @Produce json
// @Param code path string true "Currency code (e.g., BTC, USD)"
// @Success 200 {object} dto.CurrencyInfoResponse
// @Failure 404 {object} map[string]string
// @Router /admin/currencies/{code} [get]
func (h *AdminCurrencyHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := h.parseCurrencyCode(r)

	currency, err := h.currencyRepo.Get(ctx, code)
	if err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgCurrencyNotFound)
		return
	}

	response.JSON(w, http.StatusOK, dto.ToCurrencyInfoResponse(currency))
}

// Create godoc
// @Summary Create new currency
// @Description Add a new currency to the system with provider preferences
// @Tags admin
// @Accept json
// @Produce json
// @Param request body dto.CreateCurrencyRequest true "Currency data"
// @Success 201 {object} dto.CurrencyInfoResponse
// @Failure 400 {object} map[string]string
// @Router /admin/currencies [post]
func (h *AdminCurrencyHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	if !req.Type.IsValid() {
		response.Error(w, http.StatusBadRequest, "invalid currency type")
		return
	}

	currency := &domain.CurrencyInfo{
		Code:          req.Code,
		Name:          req.Name,
		Precision:     req.Precision,
		Type:          req.Type,
		Symbol:        req.Symbol,
		BaseUnit:      req.BaseUnit,
		IsActive:      true,
		ProviderPrefs: req.ProviderPrefs,
	}

	if err := h.currencyRepo.Create(ctx, currency); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, dto.ToCurrencyInfoResponse(currency))
}

// Update godoc
// @Summary Update currency
// @Description Update currency properties (name, precision, symbol, etc.)
// @Tags admin
// @Accept json
// @Produce json
// @Param code path string true "Currency code"
// @Param request body dto.UpdateCurrencyRequest true "Fields to update"
// @Success 200 {object} dto.CurrencyInfoResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/currencies/{code} [put]
func (h *AdminCurrencyHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := h.parseCurrencyCode(r)

	var req dto.UpdateCurrencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	currency, err := h.currencyRepo.Get(ctx, code)
	if err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgCurrencyNotFound)
		return
	}

	h.applyUpdates(currency, req)

	if err := h.currencyRepo.Update(ctx, currency); err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, dto.ToCurrencyInfoResponse(currency))
}

// applyUpdates modifies currency based on update request.
func (h *AdminCurrencyHandler) applyUpdates(c *domain.CurrencyInfo, req dto.UpdateCurrencyRequest) {
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Precision != nil {
		c.Precision = *req.Precision
	}
	if req.Symbol != nil {
		c.Symbol = *req.Symbol
	}
	if req.BaseUnit != nil {
		c.BaseUnit = *req.BaseUnit
	}
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}
}

// UpdateProviders godoc
// @Summary Update currency provider preferences
// @Description Set primary and fallback providers for a currency
// @Tags admin
// @Accept json
// @Produce json
// @Param code path string true "Currency code"
// @Param request body dto.UpdateProvidersRequest true "Provider preferences"
// @Success 200 {object} dto.CurrencyInfoResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/currencies/{code}/providers [put]
func (h *AdminCurrencyHandler) UpdateProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := h.parseCurrencyCode(r)

	var req dto.UpdateProvidersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	if err := h.validateProviderPrefs(req.Providers); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.currencyRepo.UpdateProviderPrefs(ctx, code, req.Providers); err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgCurrencyNotFound)
		return
	}

	currency, err := h.currencyRepo.Get(ctx, code)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	response.JSON(w, http.StatusOK, dto.ToCurrencyInfoResponse(currency))
}

// validateProviderPrefs validates provider preference entries.
func (h *AdminCurrencyHandler) validateProviderPrefs(prefs []domain.ProviderPreference) error {
	for _, p := range prefs {
		if p.Provider == "" {
			return fmt.Errorf("provider name is required")
		}
		if p.Priority < 1 {
			return fmt.Errorf("priority must be >= 1")
		}
	}
	return nil
}

// Delete godoc
// @Summary Delete currency
// @Description Soft delete a currency (mark as inactive)
// @Tags admin
// @Param code path string true "Currency code"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/currencies/{code} [delete]
func (h *AdminCurrencyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := h.parseCurrencyCode(r)

	if err := h.currencyRepo.Delete(ctx, code); err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgCurrencyNotFound)
		return
	}

	response.Success(w, "currency deleted")
}
