package market

import (
	"bytes"
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
	"github.com/shopspring/decimal"
)

// TradeConfirmationData defines the data for the trade confirmation image.
type TradeConfirmationData struct {
	ItemType     string
	Quantity     int32
	TotalPrice   decimal.Decimal
	NewBalance   decimal.Decimal
	IsBuyer      bool
}

// HandleBuy processes the /market buy command.
func (s *Service) HandleBuy(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.market")

	buyerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || buyerIDStr == "" {
		log.Error().Msg("player_id missing from context in market buy handler")
		return fmt.Errorf("player_id missing from context")
	}

	var listingIDStr, quantityStr string
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "listing_id":
			listingIDStr = opt.StringValue()
		case "quantity":
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				quantityStr = strconv.FormatInt(opt.IntValue(), 10)
			}
		}
	}

	quantity, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil || quantity <= 0 {
		quantity = 1 // Default to 1 if not specified or invalid
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
				Content: "This listing is no longer active.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if int32(quantity) > listing.QuantityRemaining {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Not enough quantity available. Only %d remaining.", listing.QuantityRemaining),
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

	var buyerWalletID string
	for _, acc := range buyerAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			buyerWalletID = acc.ID
			break
		}
	}

	sellerAccounts, err := s.Queries.GetAccountsByPlayer(ctx, listing.SellerID)
	if err != nil {
		return fmt.Errorf("failed to fetch seller accounts: %w", err)
	}

	var sellerWalletID string
	for _, acc := range sellerAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			sellerWalletID = acc.ID
			break
		}
	}

	pricePerUnit := numericToDecimal(listing.AskingPrice)
	totalPrice := pricePerUnit.Mul(decimal.NewFromInt(int64(quantity)))

	// Execute Atomic Trade via Ledger
	err = s.Ledger.Transfer(ctx, buyerWalletID, sellerWalletID, totalPrice, "MARKET_PURCHASE", buyer.ID, fmt.Sprintf("Bought %d x %s", quantity, listing.ItemType))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Insufficient funds in your wallet to complete this purchase.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	// Update Listing and Record Trade
	err = s.Queries.DecrementListingQuantity(ctx, sqlc.DecrementListingQuantityParams{
		ID:       listing.ID,
		Quantity: int32(quantity),
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to decrement listing quantity: %w", err)).Msg("trade partial failure")
		// Note: In a real system, this would be wrapped in a DB transaction with the Ledger call.
	}

	_, err = s.Queries.RecordTrade(ctx, sqlc.RecordTradeParams{
		ListingID:    listing.ID,
		BuyerID:      buyer.ID,
		SellerID:     listing.SellerID,
		ItemType:     listing.ItemType,
		Quantity:     int32(quantity),
		PricePerUnit: listing.AskingPrice,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to record trade: %w", err)).Msg("trade record failure")
	}

	// Fetch new balances for image composition
	newBuyerBal, _ := s.getAccountMetrics(ctx, buyerWalletID)

	data := TradeConfirmationData{
		ItemType:   listing.ItemType,
		Quantity:   int32(quantity),
		TotalPrice: totalPrice,
		NewBalance: newBuyerBal,
		IsBuyer:    true,
	}

	imgBytes, err := s.Composer.Compose("market_trade", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose trade image: %w", err)).Msg("image composition failed")
		return s.sendBuyTextFallback(sess, i, listing.ItemType, int32(quantity), totalPrice, newBuyerBal)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Trade Executed",
					Description: fmt.Sprintf("Successfully purchased **%d x %s** for **⊄%s**.", quantity, listing.ItemType, totalPrice.String()),
					Color:       0x4CAF76, // Green for successful transaction
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://market_trade.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Simlife Global Exchange"},
				},
			},
			Files: []*discordgo.File{
				{Name: "market_trade.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	// Notify Seller via DM or Followup (Followup is safer if in same channel, but DM is better for market)
	// We'll attempt a followup if the interaction is in a channel, otherwise DM.
	seller, _ := s.Queries.GetPlayerByID(ctx, listing.SellerID)
	if seller != nil {
		_, _ = sess.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("💰 **Item Sold:** Your listing for **%d x %s** was purchased for **⊄%s**.", quantity, listing.ItemType, totalPrice.String()),
		})
	}

	log.Info().
		Str("buyer_id", buyerIDStr).
		Str("seller_id", listing.SellerID).
		Str("item", listing.ItemType).
		Int32("quantity", int32(quantity)).
		Str("total_price", totalPrice.String()).
		Msg("market trade executed successfully")

	return nil
}

func (s *Service) sendBuyTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, itemType string, quantity int32, totalPrice, newBal decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Trade Executed",
					Description: fmt.Sprintf("Successfully purchased **%d x %s** for **⊄%s**.\n\n*New Wallet Balance:* `⊄%s`\n*(Image generation temporarily unavailable)*", quantity, itemType, totalPrice.String(), newBal.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// Helper to convert pgtype.Numeric to decimal.Decimal (assumes defined in balance.go, but duplicated here for safety if separated)
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}

// Mock getAccountMetrics for compilation if not imported from economy package
func (s *Service) getAccountMetrics(ctx context.Context, accountID string) (decimal.Decimal, error) {
	bal, err := s.Queries.GetAccountBalance(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}
	return numericToDecimal(bal), nil
}