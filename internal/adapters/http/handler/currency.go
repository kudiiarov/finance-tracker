package handler

import (
	"net/http"

	"github.com/sereozha/finance-tracker/internal/adapters/http/dto"
	"github.com/sereozha/finance-tracker/internal/ports/repository"
	"github.com/sereozha/finance-tracker/pkg/response"
)

// CurrencyHandler handles currency requests.
type CurrencyHandler struct {
	currencyRepo repository.CurrencyRepository
}

// NewCurrencyHandler creates a new handler.
func NewCurrencyHandler(currencyRepo repository.CurrencyRepository) *CurrencyHandler {
	return &CurrencyHandler{currencyRepo: currencyRepo}
}

// List godoc
// @Summary List all currencies
// @Description Get all supported currencies with their properties
// @Tags currencies
// @Produce json
// @Success 200 {object} dto.ListCurrenciesResponse
// @Router /currencies [get]
func (h *CurrencyHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currencies, err := h.currencyRepo.ListActive(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]dto.CurrencyResponse, len(currencies))
	for i, c := range currencies {
		resp[i] = dto.CurrencyResponse{
			Code:      c.Code,
			Name:      c.Name,
			Precision: c.Precision,
		}
	}

	response.JSON(w, http.StatusOK, dto.ListCurrenciesResponse{Currencies: resp})
}
