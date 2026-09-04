package market

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
)

// Service holds the dependencies for all market command handlers.
type Service struct {
	Queries  *db.Queries
	Market   *economy.MarketEngine
	Pricing  *economy.PricingEngine
	Composer *imaging.Composer
	Log      zerolog.Logger
}

// NewService initializes the market command service.
func NewService(q *db.Queries, m *economy.MarketEngine, p *economy.PricingEngine, c *imaging.Composer, log zerolog.Logger) *Service {
	return &Service{
		Queries:  q,
		Market:   m,
		Pricing:  p,
		Composer: c,
		Log:      log.With().Str("package", "commands.market").Logger(),
	}
}

// HandleList processes the /market list command.
func (s *Service) HandleList(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	playerIDStr := ctx.Value(commands.PlayerIDKey).(string)
	
	itemType := getStringOption(i, "item")
	quantity := getIntOption(i, "quantity")
	price := getFloatOption(i, "price")

	if itemType == "" || quantity <= 0 || price <= 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide valid item, quantity (>0), and price (>0).",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByDiscordID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	walletID, escrowID, err := s.getAccountIDs(ctx, player.ID)
	if err != nil {
		return err
	}

	listingID, err := s.Market.CreateListing(ctx, player.ID, walletID, escrowID, itemType, int32(quantity), price)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to create listing: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	data := layouts.MarketData{
		ItemType:     itemType,
		Quantity:     quantity,
		PricePerUnit: int64(price),
		ExpiryHours:  48,
	}

	imgBytes, err := s.Composer.Compose("market", data)
	if err != nil {
		s.Log.Error().Err(err).Msg("failed to compose market image")
	}

	_ = listingID // Used in real implementation to generate cancel/buy buttons

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Market Listing Created",
					Description: fmt.Sprintf("Successfully listed %d x %s for $%.2f each.", quantity, itemType, price),
					Color:       0x0000FF, // Blue for market commands
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://market.png"},
				},
			},
			Files: []*discordgo.File{
				{Name: "market.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func (s *Service) getAccountIDs(ctx context.Context, playerID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	pgPlayerID := pgtype.UUID{Bytes: playerID, Valid: true}
	accounts, err := s.Queries.GetAccountsByPlayer(ctx, pgPlayerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID, escrowID uuid.UUID
	for _, acc := range accounts {
		switch acc.Type {
		case db.AccountTypeWALLET:
			walletID = acc.ID
		case db.AccountTypeESCROW:
			escrowID = acc.ID
		}
	}
	return walletID, escrowID, nil
}

// Helper functions for Discord interaction options
func getStringOption(i *discordgo.InteractionCreate, name string) string {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionString {
			if val, ok := opt.Value.(string); ok {
				return val
			}
		}
	}
	return ""
}

func getIntOption(i *discordgo.InteractionCreate, name string) int {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionInteger {
			if val, ok := opt.Value.(int64); ok {
				return int(val)
			}
		}
	}
	return 0
}

func getFloatOption(i *discordgo.InteractionCreate, name string) float64 {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name && opt.Type == discordgo.ApplicationCommandOptionNumber {
			if val, ok := opt.Value.(float64); ok {
				return val
			}
		}
	}
	return 0
}