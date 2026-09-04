package jobs

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
)

// RegisterDailySettlement schedules the daily settlement engine.
func (s *Scheduler) RegisterDailySettlement(engine *economy.Engine, sess *discordgo.Session, composer *imaging.Composer) error {
	_, err := s.cron.Cron(s.cfg.SettlementCron).Do(func() {
		s.log.Info().Msg("starting daily settlement")

		// 15 minute outer timeout to allow for news publishing after the 10 min engine timeout
		ctx, cancel := context.WithTimeout(s.ctx, 15*time.Minute)
		defer cancel()

		if err := engine.RunDailySettlement(ctx); err != nil {
			s.log.Fatal().Err(err).Msg("daily settlement failed")
			// TODO: Notify developer Discord channel of FATAL settlement failure
			return
		}

		// Publish economic news after successful settlement
		PublishEconomicNews(ctx, sess, composer, s.log)
	})

	return err
}