package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
)

const (
	// settlementOuterTimeout allows 5 extra minutes beyond the engine's 10-minute
	// hard timeout so the news publisher and post-settlement hooks can complete.
	settlementOuterTimeout = 15 * time.Minute
	// settlementKey is the Redis key updated after each successful settlement,
	// used by the analytics health check to detect stale data.
	settlementKey = "analytics:last_settlement"
)

// RegisterDailySettlement schedules the daily settlement engine using the cron
// expression from configuration. After a successful settlement, it triggers the
// economic news publisher. On failure, it logs a FATAL event and notifies the
// developer's configured Discord channel.
func (s *Scheduler) RegisterDailySettlement(sess *discordgo.Session, composer *imaging.Composer) error {
	if s.settlement == nil {
		return fmt.Errorf("settlement engine not registered; call RegisterSettlementEngine before RegisterDailySettlement")
	}

	_, err := s.cron.Cron(s.cfg.SettlementCron).Do(func() {
		log := logger.FromContext(s.ctx, "jobs.settlement")
		log.Info().Str("cron", s.cfg.SettlementCron).Msg("daily settlement triggered")

		ctx, cancel := context.WithTimeout(s.ctx, settlementOuterTimeout)
		defer cancel()

		start := time.Now()

		// Run the settlement engine (has its own 10-minute internal timeout)
		if err := s.settlement.RunDailySettlement(ctx); err != nil {
			log.Error().Err(fmt.Errorf("settlement engine failed: %w", err)).
				Dur("elapsed", time.Since(start)).
				Msg("daily settlement failed")

			// Notify developer channel if configured
			s.notifyDeveloperChannel(ctx, sess, fmt.Sprintf("❌ **Daily Settlement Failed**\nError: `%v`\nElapsed: %s", err, time.Since(start).Round(time.Second)))
			return
		}

		// Mark settlement as complete in Redis for the analytics health check
		if err := s.redis.Set(ctx, settlementKey, time.Now().UTC().Format(time.RFC3339), 26*time.Hour).Err(); err != nil {
			log.Warn().Err(fmt.Errorf("failed to update settlement timestamp: %w", err)).
				Msg("settlement completed but Redis timestamp update failed")
		}

		log.Info().
			Dur("elapsed", time.Since(start)).
			Msg("daily settlement completed successfully")

		// Publish the daily economic news bulletin
		PublishEconomicNews(ctx, sess, composer, s.queries, s.log)
	})

	if err != nil {
		return fmt.Errorf("failed to register daily settlement job: %w", err)
	}

	s.log.Info().
		Str("cron", s.cfg.SettlementCron).
		Msg("daily settlement job registered")

	return nil
}

// notifyDeveloperChannel sends a message to the developer's configured Discord
// channel. Failures here are logged but never propagated — the settlement result
// is the primary concern.
func (s *Scheduler) notifyDeveloperChannel(ctx context.Context, sess *discordgo.Session, message string) {
	if s.cfg.DeveloperChannelID == "" {
		return
	}

	log := logger.FromContext(ctx, "jobs.settlement")

	_, err := sess.ChannelMessageSend(s.cfg.DeveloperChannelID, message)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to notify developer channel: %w", err)).
			Str("channel_id", s.cfg.DeveloperChannelID).
			Msg("developer notification failed")
	}
}