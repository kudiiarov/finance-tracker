package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sereozha/finance-tracker/internal/adapters/http/dto"
	"github.com/sereozha/finance-tracker/internal/domain"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/internal/ports/service"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// AccountHandler handles account-related HTTP requests.
type AccountHandler struct {
	accountSvc   service.AccountService
	settingsSvc  service.SettingsService
	rateSvc      service.ExchangeRateService
	currencyRepo repository.CurrencyRepository
}

// NewAccountHandler creates a new AccountHandler instance.
func NewAccountHandler(
	accountSvc service.AccountService,
	settingsSvc service.SettingsService,
	rateSvc service.ExchangeRateService,
	currencyRepo repository.CurrencyRepository,
) *AccountHandler {
	return &AccountHandler{
		accountSvc:   accountSvc,
		settingsSvc:  settingsSvc,
		rateSvc:      rateSvc,
		currencyRepo: currencyRepo,
	}
}

// getMainCurrency retrieves the user's preferred main currency with USD fallback.
func (h *AccountHandler) getMainCurrency(ctx context.Context) domain.Currency {
	userID := contextkey.UserID(ctx)

	code, err := h.settingsSvc.GetMainCurrency(ctx, userID)
	if err != nil {
		code = "USD"
	}

	if currInfo, err := h.currencyRepo.Get(ctx, code); err == nil {
		return currInfo.ToCurrency()
	}

	return domain.Currency{Code: "USD", Precision: 2}
}

// getConversionRate fetches the exchange rate between two currencies.
// Returns 1.0 if currencies are the same.
func (h *AccountHandler) getConversionRate(ctx context.Context, from, to string) float64 {
	if from == to {
		return 1.0
	}

	rate, err := h.rateSvc.GetRate(ctx, from, to)
	if err != nil {
		return 0
	}

	return rate.Rate
}

// buildAccountResponse creates an AccountResponse with conversion to main currency.
func (h *AccountHandler) buildAccountResponse(ctx context.Context, account *domain.Account) dto.AccountResponse {
	mainCurrency := h.getMainCurrency(ctx)
	rate := h.getConversionRate(ctx, account.Currency.Code, mainCurrency.Code)

	return dto.ToAccountResponse(*account, mainCurrency, rate)
}

// parseUUID parses a UUID string from URL parameters.
func (h *AccountHandler) parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

// Create godoc
// @Summary Create a new account
// @Description Create a new financial account with specified currency and amount
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "Account data"
// @Success 201 {object} dto.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	currInfo, err := h.currencyRepo.Get(ctx, req.Currency)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s: %s", ErrMsgInvalidCurrency, req.Currency))
		return
	}

	account, err := h.accountSvc.Create(ctx, contextkey.UserID(ctx), service.CreateAccountInput{
		Name:        req.Name,
		Type:        domain.AccountType(req.Type),
		Currency:    currInfo.ToCurrency(),
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, h.buildAccountResponse(ctx, account))
}

// Get godoc
// @Summary Get account by ID
// @Description Get a single account with converted amounts
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID (UUID)"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [get]
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := h.parseUUID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidID)
		return
	}

	account, err := h.accountSvc.Get(ctx, contextkey.UserID(ctx), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgAccountNotFound)
		return
	}

	response.JSON(w, http.StatusOK, h.buildAccountResponse(ctx, account))
}

// List godoc
// @Summary List all accounts
// @Description Get all user accounts with converted amounts to main currency
// @Tags accounts
// @Produce json
// @Success 200 {object} dto.ListAccountsResponse
// @Failure 500 {object} map[string]string
// @Router /accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := contextkey.UserID(ctx)

	accounts, err := h.accountSvc.List(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}

	mainCurrency := h.getMainCurrency(ctx)
	rates := h.buildRateMap(ctx, accounts, mainCurrency.Code)

	response.JSON(w, http.StatusOK, dto.ToAccountResponseList(accounts, mainCurrency, rates))
}

// buildRateMap creates a map of currency codes to conversion rates.
func (h *AccountHandler) buildRateMap(ctx context.Context, accounts []domain.Account, mainCurrency string) map[string]float64 {
	rates := make(map[string]float64)

	for _, account := range accounts {
		code := account.Currency.Code
		if _, exists := rates[code]; !exists {
			rates[code] = h.getConversionRate(ctx, code, mainCurrency)
		}
	}

	return rates
}

// Update godoc
// @Summary Update account
// @Description Update account details (name, type, amount, description)
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID (UUID)"
// @Param request body dto.UpdateAccountRequest true "Fields to update"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [put]
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := h.parseUUID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidID)
		return
	}

	var req dto.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequestBody)
		return
	}

	input := h.buildUpdateInput(req)

	account, err := h.accountSvc.Update(ctx, contextkey.UserID(ctx), id, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, h.buildAccountResponse(ctx, account))
}

// buildUpdateInput converts DTO to service input.
func (h *AccountHandler) buildUpdateInput(req dto.UpdateAccountRequest) service.UpdateAccountInput {
	input := service.UpdateAccountInput{}

	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Type != nil {
		accType := domain.AccountType(*req.Type)
		input.Type = &accType
	}
	if req.Amount != nil {
		input.Amount = req.Amount
	}
	if req.Description != nil {
		input.Description = req.Description
	}

	return input
}

// Delete godoc
// @Summary Delete account
// @Description Delete an account permanently
// @Tags accounts
// @Param id path string true "Account ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /accounts/{id} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := h.parseUUID(r, "id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, ErrMsgInvalidID)
		return
	}

	if err := h.accountSvc.Delete(ctx, contextkey.UserID(ctx), id); err != nil {
		response.Error(w, http.StatusNotFound, ErrMsgAccountNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
