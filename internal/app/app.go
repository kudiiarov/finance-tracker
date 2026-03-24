package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sereozha/finance-tracker/internal/adapters/config"
	"github.com/sereozha/finance-tracker/internal/adapters/exchange"
	httpAdapter "github.com/sereozha/finance-tracker/internal/adapters/http"
	"github.com/sereozha/finance-tracker/internal/adapters/http/handler"
	loggerAdapter "github.com/sereozha/finance-tracker/internal/adapters/logger"
	"github.com/sereozha/finance-tracker/internal/adapters/persistence/postgres"
	"github.com/sereozha/finance-tracker/internal/adapters/scheduler"
	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/services/account"
	exchangeSvc "github.com/sereozha/finance-tracker/internal/services/exchange"
	"github.com/sereozha/finance-tracker/internal/services/settings"
	"github.com/sereozha/finance-tracker/internal/services/statistics"
)

// App represents the application with all its dependencies.
type App struct {
	config            *config.Config
	db                *pgxpool.Pool
	rateScheduler     *scheduler.RateScheduler
	snapshotScheduler *scheduler.SnapshotScheduler
	server            *http.Server
	log               logger.Logger
}

// New creates a new application instance.
func New(cfg *config.Config) (*App, error) {
	// Initialize logger first (for all subsequent logging)
	log, err := loggerAdapter.NewZapLogger(cfg.Server.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("application_starting",
		logger.String("mode", cfg.Server.Mode),
		logger.String("port", cfg.Server.Port),
	)

	// Database connection
	log.Info("connecting_to_database")
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Error("database_connection_failed", logger.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info("database_connected")

	// Repositories
	accountRepo := postgres.NewAccountRepository(db)
	rateRepo := postgres.NewExchangeRateRepository(db)
	settingsRepo := postgres.NewSettingsRepository(db)
	currencyRepo := postgres.NewCurrencyRepository(db)
	snapshotRepo := postgres.NewPortfolioSnapshotRepository(db)

	// Provider registry
	providerRegistry := exchange.NewCapabilityRegistry()

	// Services
	accountSvc := account.NewService(accountRepo)
	rateSvc := exchangeSvc.NewService(currencyRepo, rateRepo, providerRegistry, log)
	settingsSvc := settings.NewService(settingsRepo)
	statsSvc := statistics.NewService(accountRepo, rateSvc, settingsSvc)
	snapshotSvc := statistics.NewSnapshotService(accountRepo, snapshotRepo, rateRepo, currencyRepo, log)

	// Schedulers
	schedulerCfg, err := LoadSchedulerConfig()
	if err != nil {
		log.Error("scheduler_config_failed", logger.Error(err))
		db.Close()
		return nil, err
	}
	rateScheduler := scheduler.NewRateScheduler(rateSvc, schedulerCfg.Interval, schedulerCfg.Timeout, log)
	snapshotScheduler := scheduler.NewSnapshotScheduler(snapshotSvc, log)

	// Handlers
	handlers := &httpAdapter.RouterConfig{
		AccountHandler:       handler.NewAccountHandler(accountSvc, settingsSvc, rateSvc, currencyRepo),
		CurrencyHandler:      handler.NewCurrencyHandler(currencyRepo),
		SettingsHandler:      handler.NewSettingsHandler(settingsSvc),
		StatisticsHandler:    handler.NewStatisticsHandler(statsSvc, rateSvc, snapshotSvc),
		SnapshotHandler:      handler.NewSnapshotHandler(snapshotSvc),
		AdminCurrencyHandler: handler.NewAdminCurrencyHandler(currencyRepo),
		Log:                  log,
	}

	// Router
	router := httpAdapter.NewRouter(handlers)

	// HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &App{
		config:            cfg,
		db:                db,
		rateScheduler:     rateScheduler,
		snapshotScheduler: snapshotScheduler,
		server:            server,
		log:               log,
	}, nil
}

// Run starts the application and blocks until shutdown.
func (a *App) Run() error {
	a.log.Info("starting_server",
		logger.String("address", a.server.Addr),
	)

	// Start rate fetch scheduler
	a.rateScheduler.Start()

	// Start snapshot scheduler (for daily portfolio history)
	a.snapshotScheduler.Start()

	// Start server in goroutine
	go func() {
		a.log.Info("server_listening",
			logger.String("url", fmt.Sprintf("http://localhost%s", a.server.Addr)),
		)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal("server_error", logger.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.log.Info("shutdown_signal_received")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.rateScheduler.Stop()
	a.snapshotScheduler.Stop()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("server_shutdown_error", logger.Error(err))
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	a.log.Info("server_shutdown_complete")
	return nil
}

// Close cleans up resources.
func (a *App) Close() {
	if a.snapshotScheduler != nil {
		a.snapshotScheduler.Stop()
	}
	if a.rateScheduler != nil {
		a.rateScheduler.Stop()
	}
	if a.db != nil {
		a.log.Info("closing_database_connection")
		a.db.Close()
	}
	if a.log != nil {
		a.log.Sync()
	}
}
