// Package app provides application composition and lifecycle management.
package app

import (
	"fmt"
	"os"
	"time"

	"github.com/sereozha/finance-tracker/internal/adapters/scheduler"
)

// SchedulerConfig holds scheduler configuration.
type SchedulerConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

// LoadSchedulerConfig loads scheduler configuration from environment.
func LoadSchedulerConfig() (*SchedulerConfig, error) {
	// Parse interval
	intervalStr := os.Getenv("RATE_FETCH_INTERVAL")
	if intervalStr == "" {
		intervalStr = "30m"
	}
	interval, err := scheduler.ParseInterval(intervalStr, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_FETCH_INTERVAL: %w", err)
	}

	// Parse timeout
	timeoutStr := os.Getenv("RATE_FETCH_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "2m"
	}
	timeout, err := scheduler.ParseInterval(timeoutStr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_FETCH_TIMEOUT: %w", err)
	}

	return &SchedulerConfig{
		Interval: interval,
		Timeout:  timeout,
	}, nil
}
