package jobs

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"

	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// PublishEconomicNews generates and posts the daily economic bulletin.
func PublishEconomicNews(ctx context.Context, sess *discordgo.Session, composer *imaging.Composer, log zerolog.Logger) {
	// TODO: Fetch latest snapshot from DB to populate data
	data := layouts.EconomicNewsData{
		EconDay:         142,
		PriceIndex:      1.05,
		IndexChange:     0.02,
		Velocity:        0.45,
		GiniCoeff:       0.42,
		TopEarners:      []string{"Player1", "Player2", "Player3"},
		MoneySupply:     1000000,
	}

	imgBytes, err := composer.Compose("economic_news", data)
	if err != nil {
		log.Error().Err(err).Msg("failed to compose economic news image")
		return
	}

	// TODO: Fetch configured news channel ID from config
	newsChannelID := "000000000000000000"

	_, err = sess.ChannelMessageSendComplex(newsChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "Daily Economic Bulletin",
				Description: fmt.Sprintf("Market closing report for Day %d.", data.EconDay),
				Color:       0xFFD700, // Gold for economy
				Image:       &discordgo.MessageEmbedImage{URL: "attachment://economic_news.png"},
			},
		},
		Files: []*discordgo.File{
			{Name: "economic_news.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to publish economic news to Discord")
	} else {
		log.Info().Msg("daily economic news published successfully")
	}
}