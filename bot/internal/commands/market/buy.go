package market

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// HandleBuy processes the /market buy command.
func (s *Service) HandleBuy(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	buyerIDStr := ctx.Value(commands.PlayerIDKey).(string)
	listingIDStr := getStringOption(i, "listing_id")
	quantity := getIntOption(i, "quantity")

	if quantity <= 0 {
		quantity = 1 // Default to 1 if not specified
	}

	listingID, err := uuid.Parse(listingIDStr)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid listing ID.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	buyer, err := s.Queries.GetPlayerByDiscordID(ctx, buyerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch buyer: %w", err)
	}

	buyerWallet, _, err := s.getAccountIDs(ctx, buyer.ID)
	if err != nil {
		return err
	}

	// In a real implementation, fetch listing to get seller ID and escrow ID
	// Stubbed seller IDs for compilation
	sellerWallet := uuid.New()
	sellerEscrow := uuid.New()

	err = s.Market.ExecuteTrade(ctx, listingID, buyer.ID, buyerWallet, sellerWallet, sellerEscrow, int32(quantity))
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Trade failed: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	data := layouts.MarketData{
		ItemType:     "Iron Ore", // Would come from listing
		Quantity:     quantity,
		PricePerUnit: 1000, // Would come from listing
	}

	imgBytes, err := s.Composer.Compose("market", data)
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to compose market image")
	}

	// Notify buyer
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Trade Confirmation",
					Color: 0x0000FF, // Blue for market
					Image: &discordgo.MessageEmbedImage{URL: "attachment://market.png"},
				},
			},
			Files: []*discordgo.File{
				{Name: "market.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	// TODO: Send DM or followup to seller with their confirmation image

	return err
}
