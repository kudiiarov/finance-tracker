package handler

// Standard error messages for HTTP responses.
// These constants ensure consistency across all handlers.
const (
	// Validation errors
	ErrMsgInvalidRequestBody = "invalid request body"
	ErrMsgInvalidID          = "invalid ID format"
	ErrMsgInvalidCurrency    = "invalid currency"
	ErrMsgInvalidDate        = "invalid date format"

	// Not found errors
	ErrMsgAccountNotFound  = "account not found"
	ErrMsgCurrencyNotFound = "currency not found"
	ErrMsgSnapshotNotFound = "snapshot not found"
	ErrMsgSettingsNotFound = "settings not found"

	// Server errors
	ErrMsgInternalServer     = "internal server error"
	ErrMsgServiceUnavailable = "service temporarily unavailable"
)
