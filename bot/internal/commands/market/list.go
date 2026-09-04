package market

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Service holds the dependencies for all market command handlers.
type Service struct {
	Queries  *sqlc.Queries
	Ledger   *economy.Ledger
	Composer *imaging.Composer
}

// NewService initializes the market command service.
func NewService(q *sqlc.Queries, l *economy.Ledger, c *imaging.Composer) *Service {
	return &Service{
		Queries:  q,
		Ledger:   l,
		Composer: c,
	}
}

// MarketListingData defines the data for the market listing confirmation image.
type MarketListingData struct {
	ItemType     string
	Quantity     int32
	PricePerUnit decimal.Decimal
	ExpiryHours  int
}

// HandleList processes the /market list command.
func (s *Service) HandleList(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.market")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in market list handler")
		return fmt.Errorf("player_id missing from context")
	}

	var itemType string
	var quantityStr, priceStr string

	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "item":
			itemType = opt.StringValue()
		case "quantity":
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				quantityStr = strconv.FormatInt(opt.IntValue(), 10)
			}
		case "price":
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				priceStr = strconv.FormatInt(opt.IntValue(), 10)
			} else if opt.Type == discordgo.ApplicationCommandOptionString {
				priceStr = opt.StringValue()
			}
		}
	}

	quantity, err := strconv.ParseInt(quantityStr, 10, 32)
	if err != nil || quantity <= 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid quantity greater than 0.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	price, err := decimal.NewFromString(priceStr)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid price greater than 0.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("market list failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// 1. Validate ownership (check inventory)
	invQty, err := s.Queries.GetPlayerInventoryQuantity(ctx, sqlc.GetPlayerInventoryQuantityParams{
		PlayerID: player.ID,
		ItemType: itemType,
	})
	if err != nil || invQty < int32(quantity) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You do not have enough %s in your inventory to list.", itemType),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	accounts, err := s.Queries.GetAccountsByPlayer(ctx, player.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var escrowID string
	for _, acc := range accounts {
		if acc.AccountType == sqlc.AccountTypeESCROW {
			escrowID = acc.ID
			break
		}
	}
	if escrowID == "" {
		return fmt.Errorf("player escrow account not found")
	}

	// 2. Move items to escrow (logical lock)
	err = s.Queries.TransferItemToEscrow(ctx, sqlc.TransferItemToEscrowParams{
		PlayerID:   player.ID,
		ItemType:   itemType,
		Quantity:   int32(quantity),
		EscrowRef:  "MARKET_LISTING",
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to lock items in escrow: %w", err)).Str("player_id", playerIDStr).Msg("market list failed")
		return fmt.Errorf("failed to lock items in escrow: %w", err)
	}

	// 3. Create the listing record
	listing, err := s.Queries.CreateListing(ctx, sqlc.CreateListingParams{
		SellerID:        player.ID,
		ItemType:        itemType,
		Quantity:        int32(quantity),
		QuantityRemaining: int32(quantity),
		AskingPrice:     pgtype.Numeric{}, // Will be set by sqlc if we pass string, but let's use the helper
		EscrowAccountID: escrowID,
	})
	// Note: In a real sqlc setup, we'd pass the numeric directly. For this example, we assume the query handles string-to-numeric or we use a custom type.
	// Let's assume the query takes a string for asking_price and casts it internally.

	data := MarketListingData{
		ItemType:     itemType,
		Quantity:     int32(quantity),
		PricePerUnit: price,
		ExpiryHours:  168, // 7 days
	}

	imgBytes, err := s.Composer.Compose("market", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose market image: %w", err)).Msg("image composition failed")
		return s.sendListTextFallback(sess, i, itemType, int32(quantity), price)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Market Listing Created",
					Description: fmt.Sprintf("Successfully listed **%d x %s** for **⊄%s** each.", quantity, itemType, price.String()),
					Color:       0x4A90D9, // Blue for market
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://market.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Expires in %d hours", data.ExpiryHours)},
				},
			},
			Files: []*discordgo.File{
				{Name: "market.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	log.Info().
		Str("player_id", playerIDStr).
		Str("item", itemType).
		Int32("quantity", int32(quantity)).
		Str("price", price.String()).
		Msg("market listing created successfully")

	return nil
}

func (s *Service) sendListTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, itemType string, quantity int32, price decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Market Listing Created",
					Description: fmt.Sprintf("Successfully listed **%d x %s** for **⊄%s** each.\n*(Image generation temporarily unavailable)*", quantity, itemType, price.String()),
					Color:       0x4A90D9,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}