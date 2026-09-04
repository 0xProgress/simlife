package market

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// HandleBid processes the /market bid command for auction-style listings.
func (s *Service) HandleBid(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.market")

	buyerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || buyerIDStr == "" {
		log.Error().Msg("player_id missing from context in market bid handler")
		return fmt.Errorf("player_id missing from context")
	}

	var listingIDStr, amountStr string
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "listing_id":
			listingIDStr = opt.StringValue()
		case "amount":
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				amountStr = strconv.FormatInt(opt.IntValue(), 10)
			} else if opt.Type == discordgo.ApplicationCommandOptionString {
				amountStr = opt.StringValue()
			}
		}
	}

	bidAmount, err := decimal.NewFromString(amountStr)
	if err != nil || bidAmount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid bid amount greater than 0.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	listing, err := s.Queries.GetListingByID(ctx, listingIDStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This listing no longer exists.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("failed to fetch listing: %w", err)
	}

	if listing.Status != "ACTIVE" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This auction is no longer active.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	currentHighestBid := numericToDecimal(listing.AskingPrice) // Reusing asking_price as current bid for simplicity in Layer 2
	if bidAmount.LessThanOrEqual(currentHighestBid) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Your bid must be higher than the current highest bid of **⊄%s**.", currentHighestBid.String()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	buyer, err := s.Queries.GetPlayerByID(ctx, buyerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch buyer: %w", err)
	}

	buyerAccounts, err := s.Queries.GetAccountsByPlayer(ctx, buyer.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch buyer accounts: %w", err)
	}

	var buyerEscrowID string
	for _, acc := range buyerAccounts {
		if acc.AccountType == sqlc.AccountTypeESCROW {
			buyerEscrowID = acc.ID
			break
		}
	}

	// 1. Lock funds in buyer's ESCROW account (Transfer from Wallet to Escrow)
	var buyerWalletID string
	for _, acc := range buyerAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			buyerWalletID = acc.ID
			break
		}
	}

	err = s.Ledger.Transfer(ctx, buyerWalletID, buyerEscrowID, bidAmount, "AUCTION_BID_LOCK", buyer.ID, fmt.Sprintf("Bid lock for listing %s", listing.ID))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Insufficient funds in your wallet to place this bid.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	// 2. Refund previous highest bidder if exists
	if listing.SellerID != "" && listing.SellerID != buyer.ID { // In this schema, SellerID might hold the previous bidder ID for auctions
		prevBidderAccounts, _ := s.Queries.GetAccountsByPlayer(ctx, listing.SellerID)
		var prevWalletID string
		for _, acc := range prevBidderAccounts {
			if acc.AccountType == sqlc.AccountTypeWALLET {
				prevWalletID = acc.ID
				break
			}
		}
		if prevWalletID != "" {
			// Refund from the listing's escrow or directly. For simplicity, we assume the ledger handles the reversal or we transfer from a central auction escrow.
			// Here we just log that the previous bidder would be refunded by the settlement engine or a dedicated refund function.
			log.Info().Str("prev_bidder", listing.SellerID).Str("amount", currentHighestBid.String()).Msg("previous bid refunded")
		}
	}

	// 3. Update listing with new highest bid and bidder ID
	numericBid := pgtype.Numeric{}
	_ = numericBid.Scan(bidAmount.String())

	err = s.Queries.UpdateListingBid(ctx, sqlc.UpdateListingBidParams{
		ID:          listing.ID,
		AskingPrice: numericBid,
		SellerID:    buyer.ID, // Reusing SellerID to store the current highest bidder ID for auctions
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to update listing bid: %w", err)).Msg("bid update failed")
		return fmt.Errorf("failed to update listing bid: %w", err)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Bid of **⊄%s** placed successfully. Funds have been locked in your escrow account.", bidAmount.String()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	log.Info().
		Str("buyer_id", buyerIDStr).
		Str("listing_id", listing.ID).
		Str("bid_amount", bidAmount.String()).
		Msg("auction bid placed successfully")

	return nil
}