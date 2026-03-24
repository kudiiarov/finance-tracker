// Package http provides HTTP routing.
package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sereozha/finance-tracker/internal/adapters/http/handler"
	"github.com/sereozha/finance-tracker/internal/adapters/http/middleware"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
)

// RouterConfig holds handler dependencies.
type RouterConfig struct {
	AccountHandler       *handler.AccountHandler
	CurrencyHandler      *handler.CurrencyHandler
	SettingsHandler      *handler.SettingsHandler
	StatisticsHandler    *handler.StatisticsHandler
	SnapshotHandler      *handler.SnapshotHandler
	AdminCurrencyHandler *handler.AdminCurrencyHandler
	Log                  logger.Logger
}

// @title Finance Tracker API
// @version 1.0
// @description Personal finance tracking API with multi-currency support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@financetracker.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// NewRouter creates and configures the HTTP router.
func NewRouter(cfg *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware (order matters!)
	r.Use(chiMiddleware.RequestID)           // Chi's built-in request ID (used as fallback)
	r.Use(middleware.RequestLogger(cfg.Log)) // Structured logging with timing
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth middleware - extracts userID to context
		r.Use(middleware.UserID)

		// Accounts
		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", cfg.AccountHandler.List)
			r.Post("/", cfg.AccountHandler.Create)
			r.Get("/{id}", cfg.AccountHandler.Get)
			r.Put("/{id}", cfg.AccountHandler.Update)
			r.Delete("/{id}", cfg.AccountHandler.Delete)
		})

		// Currencies
		r.Get("/currencies", cfg.CurrencyHandler.List)

		// Portfolio statistics
		r.Route("/portfolio", func(r chi.Router) {
			r.Get("/", cfg.StatisticsHandler.GetPortfolio)
			r.Get("/history", cfg.StatisticsHandler.GetHistory)
			r.Get("/trend", cfg.StatisticsHandler.GetTrend)
		})

		// Exchange rates
		r.Route("/rates", func(r chi.Router) {
			r.Post("/refresh", cfg.StatisticsHandler.RefreshRates)
		})

		// Portfolio snapshots
		r.Route("/snapshots", func(r chi.Router) {
			r.Get("/", cfg.SnapshotHandler.List)
			r.Get("/{date}", cfg.SnapshotHandler.GetByDate)
			r.Post("/", cfg.SnapshotHandler.Create)
		})

		// User settings
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", cfg.SettingsHandler.Get)
			r.Put("/", cfg.SettingsHandler.Update)
		})

		// Admin routes (under /api/v1/admin)
		r.Route("/admin", func(r chi.Router) {
			// TODO: Add admin authentication middleware
			// r.Use(middleware.AdminAuth)

			r.Route("/currencies", func(r chi.Router) {
				r.Get("/", cfg.AdminCurrencyHandler.List)
				r.Post("/", cfg.AdminCurrencyHandler.Create)
				r.Get("/{code}", cfg.AdminCurrencyHandler.Get)
				r.Put("/{code}", cfg.AdminCurrencyHandler.Update)
				r.Delete("/{code}", cfg.AdminCurrencyHandler.Delete)
				r.Put("/{code}/providers", cfg.AdminCurrencyHandler.UpdateProviders)
			})
		})
	})

	return r
}
