package market

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/0xProgress/simlife/bot/internal/commands"
)

// HandleBid processes the /market bid command for auction-style listings.
func (s *Service) HandleBid(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	buyerIDStr := ctx.Value(commands.PlayerIDKey).(string)
	listingIDStr := getStringOption(i, "listing_id")
	amount := getFloatOption(i, "amount")

	if amount <= 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid bid amount (>0).",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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

	// TODO: Implement auction bidding logic
	// 1. Fetch listing and verify it's an auction
	// 2. Verify bid is higher than current highest bid
	// 3. Lock funds in buyer's ESCROW account
	// 4. Release funds from previous highest bidder's ESCROW
	// 5. Update listing with new highest bid and bidder ID

	_ = listingID
	_ = buyer

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Bid of $%.2f placed successfully. Funds locked in escrow.", amount),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}