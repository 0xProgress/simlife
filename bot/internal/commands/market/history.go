package market

import (
	"bytes"
	"context"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// MarketHistoryData defines the data for the market history sparkline image.
type MarketHistoryData struct {
	ItemType      string
	CurrentPrice  decimal.Decimal
	HistoryValues []float64 // Sparkline expects float64 for drawing, converted safely from decimal
}

// HandleHistory processes the /market history command.
func (s *Service) HandleHistory(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.market")

	var itemType string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "item" {
			itemType = opt.StringValue()
			break
.		}
	}

	if itemType == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify an item type to view history for.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Fetch last 30 trades for this item
	trades, err := s.Queries.GetRecentTradesByItem(ctx, sqlc.GetRecentTradesByItemParams{
		ItemType: itemType,
		Limit:    30,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch price history: %w", err)).Str("item", itemType).Msg("history fetch failed")
		return fmt.Errorf("failed to fetch price history: %w", err)
	}

	if len(trades) == 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("No trade history found for **%s**.", itemType),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Build history values for sparkline (reverse to get chronological order for drawing)
	historyValues := make([]float64, 0, len(trades))
	var currentPrice decimal.Decimal

	for j := len(trades) - 1; j >= 0; j-- {
		price := numericToDecimal(trades[j].PricePerUnit)
		historyValues = append(historyValues, price.InexactFloat64()) // Safe for sparkline drawing only
		if j == len(trades)-1 {
			currentPrice = price
		}
	}

	data := MarketHistoryData{
		ItemType:      itemType,
		CurrentPrice:  currentPrice,
		HistoryValues: historyValues,
	}

	imgBytes, err := s.Composer.Compose("market_history", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose market history image: %w", err)).Msg("image composition failed")
		return s.sendHistoryTextFallback(sess, i, itemType, currentPrice, historyValues)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("Market History: %s", itemType),
					Description: "30-day price trend and current market valuation.",
					Color:       0x4A90D9, // Blue for market
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://market_history.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Simlife Global Exchange"},
				},
			},
			Files: []*discordgo.File{
				{Name: "market_history.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	return nil
}

func (s *Service) sendHistoryTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, itemType string, currentPrice decimal.Decimal, history []float64) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("Market History: %s", itemType),
					Description: fmt.Sprintf("Current Price: **⊄%s**\n\n*(Chart generation temporarily unavailable)*", currentPrice.String()),
					Color:       0x4A90D9,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}