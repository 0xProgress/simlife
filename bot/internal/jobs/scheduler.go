package jobs

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"

	"github.com/0xProgress/simlife/bot/internal/config"
)

// Scheduler manages all background cron jobs for the Simlife bot.
type Scheduler struct {
	cron   *gocron.Scheduler
	cfg    *config.Config
	log    zerolog.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewScheduler initializes the job scheduler.
func NewScheduler(cfg *config.Config, log zerolog.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:   gocron.NewScheduler(time.UTC),
		cfg:    cfg,
		log:    log.With().Str("component", "scheduler").Logger(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins executing scheduled jobs asynchronously.
func (s *Scheduler) Start() {
	s.cron.StartAsync()
	s.log.Info().Msg("background job scheduler started")
}

// Stop gracefully halts all scheduled jobs.
func (s *Scheduler) Stop() {
	s.cancel()
	s.cron.Stop()
	s.log.Info().Msg("background job scheduler stopped")
}