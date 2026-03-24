# Finance Tracker Code Style Guide

## Table of Contents
1. [General Principles](#general-principles)
2. [Project Structure](#project-structure)
3. [Naming Conventions](#naming-conventions)
4. [Error Handling](#error-handling)
5. [Logging](#logging)
6. [HTTP Handlers](#http-handlers)
7. [Domain Models](#domain-models)
8. [Testing](#testing)

---

## General Principles

### 1.1 Clean Architecture
- **Dependency Rule**: Dependencies point inward (Domain → Ports → Services → Adapters)
- **Layer Responsibilities**:
  - `domain`: Entities, value objects, business rules (no external deps)
  - `ports`: Interfaces (repository, service contracts)
  - `services`: Business logic, use cases
  - `adapters`: External implementations (HTTP, DB, external APIs)

### 1.2 Code Organization
```
internal/
├── domain/           # Business entities
├── ports/            # Interfaces
│   ├── repository/   # Data access contracts
│   └── service/      # Business service contracts
├── services/         # Business logic implementations
└── adapters/         # External implementations
    ├── http/         # HTTP handlers, middleware
    ├── persistence/  # Database implementations
    └── exchange/     # External API clients
```

---

## Naming Conventions

### 2.1 Packages
- Lowercase, single word when possible
- Plural for collections: `accounts`, `currencies`
- Clear purpose: `exchange` not `ext`

### 2.2 Types
```go
// Interfaces: noun describing capability
type ExchangeRateService interface{}
type AccountRepository interface{}

// Structs: concrete implementation
type accountService struct{}        // private
type AccountService struct{}        // public (rare)

// DTOs: describe purpose
type CreateAccountRequest struct{}
type AccountResponse struct{}
```

### 2.3 Functions
```go
// Constructors: New + TypeName
func NewAccountService(repo AccountRepository) *accountService

// Methods: verb + noun
createAccount(ctx, input) (*Account, error)
getRate(ctx, from, to) (*ExchangeRate, error)

// Private helpers: lowercase, descriptive
calculateViaUSD(ctx, from, to) (*ExchangeRate, error)
```

### 2.4 Constants
```go
// General constants
defaultTimeout = 30 * time.Second
maxRetries     = 3

// Error message constants (reusable)
const (
    ErrInvalidRequestBody = "invalid request body"
    ErrInvalidAccountID   = "invalid account ID"
    ErrCurrencyNotFound   = "currency not found"
)
```

---

## Error Handling

### 3.1 Error Wrapping
Always wrap errors with context:
```go
// Good
if err != nil {
    return fmt.Errorf("failed to create account: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### 3.2 Error Constants
Define reusable error messages:
```go
// internal/adapters/http/handler/errors.go
package handler

const (
    ErrMsgInvalidRequest    = "invalid request body"
    ErrMsgInvalidID         = "invalid ID format"
    ErrMsgNotFound          = "resource not found"
    ErrMsgInternalServer    = "internal server error"
)
```

### 3.3 HTTP Status Codes
```go
// 400 - Bad Request: validation errors, invalid input
response.Error(w, http.StatusBadRequest, ErrMsgInvalidRequest)

// 404 - Not Found: resource doesn't exist
response.Error(w, http.StatusNotFound, ErrMsgNotFound)

// 500 - Internal Server Error: unexpected errors
response.Error(w, http.StatusInternalServerError, ErrMsgInternalServer)
```

---

## Logging

### 4.1 Structured Logging
Use field-based logging:
```go
log.Info("account_created",
    logger.String("account_id", account.ID.String()),
    logger.String("currency", account.Currency.Code),
    logger.Int64("amount", account.Amount),
)

log.Error("rate_fetch_failed",
    logger.Error(err),
    logger.String("provider", providerName),
    logger.Duration("elapsed_ms", duration.Milliseconds()),
)
```

### 4.2 Log Levels
- **Debug**: Detailed flow, variable values (development only)
- **Info**: Significant events, operations completed
- **Warn**: Unexpected but handled situations
- **Error**: Actual errors requiring attention

### 4.3 Standard Fields
```go
logger.FieldCorrelationID  // Request tracing
logger.FieldUserID         // User identification
logger.FieldComponent      // Service/component name
logger.FieldOperation      // Operation being performed
logger.FieldDuration       // Timing in milliseconds
```

---

## HTTP Handlers

### 5.1 Handler Structure
```go
// Single responsibility: parse input, call service, format response
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. Parse & validate input
    req, err := h.parseCreateRequest(r)
    if err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }
    
    // 2. Call service
    account, err := h.accountSvc.Create(ctx, req)
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    // 3. Format & send response
    response.JSON(w, http.StatusCreated, h.toResponse(account))
}
```

### 5.2 DRY Principle
Extract common patterns:
```go
// Bad: Duplicated in every handler
rate := 1.0
if account.Currency.Code != mainCurrency.Code {
    if r, err := h.rateSvc.GetRate(ctx, ...); err == nil {
        rate = r.Rate
    }
}

// Good: Extract to helper
rate := h.getConversionRate(ctx, account.Currency.Code, mainCurrency.Code)
```

### 5.3 Request Parsing
```go
func (h *Handler) parseRequest(r *http.Request, v interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(v); err != nil {
        return fmt.Errorf("%s: %w", ErrMsgInvalidRequest, err)
    }
    return validate.Struct(v)
}
```

---

## Domain Models

### 6.1 Entity Design
```go
// Domain entity: rich behavior
type Account struct {
    ID       uuid.UUID
    UserID   string
    Name     string
    Type     AccountType
    Currency Currency
    Amount   int64  // Always smallest units
}

// Validation method
func (a *Account) Validate() error {
    if a.Name == "" {
        return errors.New("account name is required")
    }
    if !a.Currency.IsValid() {
        return fmt.Errorf("invalid currency: %s", a.Currency.Code)
    }
    return nil
}
```

### 6.2 Value Objects
```go
// Immutable, compared by value
type Currency struct {
    Code      string
    Name      string
    Precision int
}

func (c Currency) ToSmallestUnit(value float64) int64 { ... }
func (c Currency) FromSmallestUnit(amount int64) float64 { ... }
```

---

## Testing

### 7.1 Test Structure
```go
func TestAccountService_Create(t *testing.T) {
    // Arrange
    repo := mock.NewAccountRepository()
    svc := NewAccountService(repo)
    
    // Act
    account, err := svc.Create(ctx, input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, account.Name)
}
```

### 7.2 Table-Driven Tests
```go
func TestCurrency_IsValid(t *testing.T) {
    tests := []struct {
        name     string
        currency string
        want     bool
    }{
        {"valid USD", "USD", true},
        {"valid BTC", "BTC", true},
        {"invalid XXX", "XXX", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := Currency{Code: tt.currency}
            assert.Equal(t, tt.want, c.IsValid())
        })
    }
}
```

---

## Code Review Checklist

### Before Submitting PR
- [ ] No duplicated code (extract helpers)
- [ ] Error messages use constants
- [ ] All errors wrapped with context
- [ ] Structured logging with fields
- [ ] Swagger comments for handlers
- [ ] No hardcoded magic values
- [ ] Functions under 50 lines
- [ ] Clear variable names
- [ ] Unit tests for business logic

### Anti-Patterns to Avoid
```go
// ❌ Magic numbers
if status == 404 { ... }  // Use http.StatusNotFound

// ❌ String concatenation in errors
errors.New("invalid " + field)

// ❌ Silent error ignoring
_ = doSomething()

// ❌ Deep nesting
if err == nil {
    if valid {
        if allowed {
            // ...
        }
    }
}

// ✅ Early returns
if err != nil {
    return err
}
if !valid {
    return ErrInvalid
}
// main logic
```

---

## Example: Refactored Handler

### Before
```go
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req dto.CreateAccountRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }
    // ... 40+ lines with duplication
}
```

### After
```go
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    req, err := h.parseCreateRequest(r)
    if err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }
    
    account, err := h.accountSvc.Create(ctx, req)
    if err != nil {
        h.handleServiceError(w, err)
        return
    }
    
    resp := h.buildAccountResponse(ctx, account)
    response.JSON(w, http.StatusCreated, resp)
}
```
