// Package scheduler handles periodic background jobs.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/ports/service"
)

// RateScheduler periodically fetches exchange rates.
type RateScheduler struct {
	rateService service.ExchangeRateService
	interval    time.Duration
	timeout     time.Duration
	ticker      *time.Ticker
	stopCh      chan struct{}
	isRunning   atomic.Bool
	stopOnce    sync.Once
	log         logger.Logger
}

// NewRateScheduler creates a new scheduler.
func NewRateScheduler(
	rateService service.ExchangeRateService,
	interval, timeout time.Duration,
	log logger.Logger,
) *RateScheduler {
	return &RateScheduler{
		rateService: rateService,
		interval:    interval,
		timeout:     timeout,
		stopCh:      make(chan struct{}),
		log:         log.With(logger.String(logger.FieldComponent, "scheduler")),
	}
}

// Start begins the scheduler in a goroutine.
func (s *RateScheduler) Start() {
	s.log.Info("starting_rate_scheduler",
		logger.String("interval", s.interval.String()),
		logger.String("timeout", s.timeout.String()),
	)

	go s.safeFetch()

	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.safeFetch()
			case <-s.stopCh:
				s.log.Info("rate_scheduler_stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler.
func (s *RateScheduler) Stop() {
	s.stopOnce.Do(func() {
		s.log.Info("stopping_rate_scheduler")
		close(s.stopCh)
		if s.ticker != nil {
			s.ticker.Stop()
		}
	})
}

func (s *RateScheduler) safeFetch() {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("panic_recovered_in_rate_fetch",
				logger.Any("panic", r),
			)
			s.isRunning.Store(false)
		}
	}()

	s.fetchWithTimeout()
}

func (s *RateScheduler) fetchWithTimeout() {
	if !s.isRunning.CompareAndSwap(false, true) {
		s.log.Warn("rate_fetch_skipped_previous_still_running")
		return
	}
	defer s.isRunning.Store(false)

	s.log.Debug("fetching_exchange_rates")

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	start := time.Now()
	err := s.rateService.FetchRates(ctx)
	duration := time.Since(start)

	if err != nil {
		s.log.Error("rate_fetch_failed",
			logger.Error(err),
			logger.Int64(logger.FieldDuration, duration.Milliseconds()),
		)
	} else {
		s.log.Info("rate_fetch_completed",
			logger.Int64(logger.FieldDuration, duration.Milliseconds()),
		)
	}
}

// IsRunning returns true if a fetch is currently in progress.
func (s *RateScheduler) IsRunning() bool {
	return s.isRunning.Load()
}

// ParseInterval parses interval string like "30m", "1h".
func ParseInterval(s string, minDuration time.Duration) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid interval '%s': %w", s, err)
	}
	if d < minDuration {
		return 0, fmt.Errorf("interval must be at least %v, got %v", minDuration, d)
	}
	return d, nil
}
