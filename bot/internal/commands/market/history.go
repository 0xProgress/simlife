package market

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// HandleHistory processes the /market history command, rendering a sparkline chart.
func (s *Service) HandleHistory(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	itemType := getStringOption(i, "item")
	if itemType == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify an item type.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	history, err := s.Pricing.GetPriceHistory(ctx, itemType)
	if err != nil {
		return fmt.Errorf("failed to fetch price history: %w", err)
	}

	currentPrice, _ := s.Pricing.GetPrice(ctx, itemType)

	data := layouts.MarketData{
		ItemType:      itemType,
		PricePerUnit:  int64(currentPrice),
		HistoryValues: history,
	}

	imgBytes, err := s.Composer.Compose("market", data)
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to compose market history image")
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("Market History: %s", itemType),
					Description: "30-day price trend and current market valuation.",
					Color:       0x0000FF, // Blue for market
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://market.png"},
				},
			},
			Files: []*discordgo.File{
				{Name: "market.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
		},
	})
}
