package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/internal/services/statistics"
)

// SnapshotScheduler creates daily portfolio snapshots for all users.
type SnapshotScheduler struct {
	snapshotService *statistics.SnapshotService
	log             logger.Logger
	ticker          *time.Ticker
	stopCh          chan struct{}
	isRunning       atomic.Bool
	stopOnce        sync.Once
}

// NewSnapshotScheduler creates a new snapshot scheduler.
func NewSnapshotScheduler(
	snapshotService *statistics.SnapshotService,
	log logger.Logger,
) *SnapshotScheduler {
	return &SnapshotScheduler{
		snapshotService: snapshotService,
		log:             log.With(logger.String(logger.FieldComponent, "snapshot_scheduler")),
		stopCh:          make(chan struct{}),
	}
}

// Start begins the scheduler.
// By default runs daily at midnight UTC.
func (s *SnapshotScheduler) Start() {
	s.log.Info("starting_snapshot_scheduler")

	// Calculate time until next midnight UTC
	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	durationUntilMidnight := nextMidnight.Sub(now)

	s.log.Info("first_snapshot_scheduled",
		logger.String("at", nextMidnight.Format(time.RFC3339)),
		logger.String("in", durationUntilMidnight.String()),
	)

	// Wait until midnight, then start ticker
	go func() {
		time.Sleep(durationUntilMidnight)
		s.createSnapshotForAllUsers()

		// Then start daily ticker
		s.ticker = time.NewTicker(24 * time.Hour)
		for {
			select {
			case <-s.ticker.C:
				s.createSnapshotForAllUsers()
			case <-s.stopCh:
				s.log.Info("snapshot_scheduler_stopped")
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler.
func (s *SnapshotScheduler) Stop() {
	s.stopOnce.Do(func() {
		s.log.Info("stopping_snapshot_scheduler")
		close(s.stopCh)
		if s.ticker != nil {
			s.ticker.Stop()
		}
	})
}

// TriggerNow manually triggers snapshot creation immediately.
func (s *SnapshotScheduler) TriggerNow() {
	go s.createSnapshotForAllUsers()
}

func (s *SnapshotScheduler) createSnapshotForAllUsers() {
	if !s.isRunning.CompareAndSwap(false, true) {
		s.log.Warn("snapshot_creation_already_running")
		return
	}
	defer s.isRunning.Store(false)

	startTime := time.Now()
	today := time.Now().UTC().Format("2006-01-02")

	s.log.Info("creating_daily_snapshots",
		logger.String("date", today),
	)

	// For now, we only have one hardcoded user
	// In future, this would iterate over all active users
	users := []string{"user-123"}

	var wg sync.WaitGroup
	var successCount, failCount int
	var mu sync.Mutex

	for _, userID := range users {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			_, err := s.snapshotService.CreateDailySnapshot(ctx, uid, today)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				s.log.Error("snapshot_failed",
					logger.Error(err),
					logger.String(logger.FieldUserID, uid),
				)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
				s.log.Info("snapshot_created",
					logger.String(logger.FieldUserID, uid),
				)
			}
		}(userID)
	}

	wg.Wait()

	duration := time.Since(startTime)
	s.log.Info("daily_snapshots_completed",
		logger.String("date", today),
		logger.Int("success", successCount),
		logger.Int("failed", failCount),
		logger.Int64("duration_ms", duration.Milliseconds()),
	)
}
